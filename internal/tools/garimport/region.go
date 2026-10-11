package garimport

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/google/uuid"

	"github.com/stickpro/go-store/pkg/gar"
)

const country = "Россия"

// FIAS levels as stored in cities.fias_level (the pre-GAR ФИАС AOLEVEL that
// hflabs/city and DaData use), not GAR's AS_OBJECT_LEVELS.
const (
	fiasLevelRegion     = 1
	fiasLevelCity       = 4
	fiasLevelSettlement = 6
)

// capital_marker values, as in hflabs/city.
const (
	capitalNone   = 0
	capitalArea   = 1 // центр района (муниципального образования)
	capitalRegion = 2 // центр региона
)

// row is one cities row as GAR describes it.
type row struct {
	Address         string
	PostalCode      *string
	FederalDistrict string
	RegionType      string
	Region          string
	AreaType        *string
	Area            *string
	CityType        *string
	City            *string
	SettlementType  *string
	Settlement      *string
	KladrID         *string
	FiasID          uuid.UUID
	FiasLevel       int16
	CapitalMarker   int16
	Okato           *string
	Oktmo           *string
	TaxOffice       *string
}

type object struct {
	guid     uuid.UUID
	name     string
	typeName string
	level    int64
}

type param struct {
	value     string
	startDate string
}

type regionResult struct {
	code          string
	rows          []row
	noHierarchy   int // settlements missing from AS_ADM_HIERARCHY
	badGUID       int
	federalCity   bool
	regionName    string
	compressedLen int64
}

var wantedParams = map[int64]bool{
	gar.ParamIFNSFL: true, gar.ParamPostIndex: true, gar.ParamOKATO: true,
	gar.ParamOKTMO: true, gar.ParamKLADR: true, gar.ParamIsAdmCapital: true,
	gar.ParamIsMunCapital: true,
}

// loadRegion reads one region folder and returns its settlements. today
// (YYYY-MM-DD) filters out parameter values that have already ended.
func loadRegion(ctx context.Context, arc *gar.Archive, code, district, today string) (*regionResult, error) {
	res := &regionResult{code: code}
	for _, kind := range []string{gar.KindAddrObj, gar.KindAddrObjParams, gar.KindAdmHierarchy} {
		res.compressedLen += arc.CompressedSize(code, kind)
	}

	objects, err := readObjects(ctx, arc, code)
	if err != nil {
		return nil, err
	}

	var (
		region   *object
		regionID int64
	)
	targets := make(map[int64]*object)
	for id, o := range objects {
		switch o.level {
		case gar.LevelRegion:
			region, regionID = o, id
		case gar.LevelCity, gar.LevelSettlement:
			targets[id] = o
		}
	}
	if region == nil {
		return nil, fmt.Errorf("region %s: no active region object", code)
	}
	res.regionName = region.name
	// Federal cities (Москва, Санкт-Петербург, Севастополь) are regions in
	// GAR but cities for customers; hflabs lists them as such too.
	res.federalCity = strings.HasPrefix(region.typeName, "г")
	if res.federalCity {
		targets[regionID] = region
	}

	paths, err := readHierarchy(ctx, arc, code, targets)
	if err != nil {
		return nil, err
	}
	params, err := readParams(ctx, arc, code, targets, today)
	if err != nil {
		return nil, err
	}

	res.rows = make([]row, 0, len(targets))
	for id, o := range targets {
		if o.guid == uuid.Nil {
			res.badGUID++
			continue
		}
		r := row{
			FederalDistrict: district,
			RegionType:      region.typeName,
			Region:          region.name,
			FiasID:          o.guid,
		}

		switch o.level {
		case gar.LevelRegion:
			r.FiasLevel = fiasLevelRegion
		case gar.LevelCity:
			r.FiasLevel = fiasLevelCity
			r.CityType, r.City = ptr(o.typeName), ptr(o.name)
		default:
			r.FiasLevel = fiasLevelSettlement
			r.SettlementType, r.Settlement = ptr(o.typeName), ptr(o.name)
		}

		if o.level != gar.LevelRegion {
			path, ok := paths[id]
			if !ok {
				res.noHierarchy++
			}
			// Nearest ancestors win: walk from the parent up to the region.
			for i := len(path) - 2; i >= 0; i-- {
				a := objects[path[i]]
				if a == nil {
					continue
				}
				switch {
				case (a.level == gar.LevelAdmArea || a.level == gar.LevelMunArea) && r.Area == nil:
					r.AreaType, r.Area = ptr(a.typeName), ptr(a.name)
				case a.level == gar.LevelCity && r.City == nil:
					r.CityType, r.City = ptr(a.typeName), ptr(a.name)
				}
			}
		}

		p := params[id]
		r.PostalCode = paramPtr(p, gar.ParamPostIndex)
		r.Okato = paramPtr(p, gar.ParamOKATO)
		r.Oktmo = paramPtr(p, gar.ParamOKTMO)
		r.KladrID = paramPtr(p, gar.ParamKLADR)
		r.TaxOffice = paramPtr(p, gar.ParamIFNSFL)
		switch {
		case isTrue(p[gar.ParamIsAdmCapital].value):
			r.CapitalMarker = capitalRegion
		case isTrue(p[gar.ParamIsMunCapital].value):
			r.CapitalMarker = capitalArea
		default:
			r.CapitalMarker = capitalNone
		}

		r.Address = address(&r)
		res.rows = append(res.rows, r)
	}
	return res, nil
}

// readObjects returns the active regions, areas, cities and settlements of a
// region folder, keyed by OBJECTID. Streets and other lower levels are skipped.
func readObjects(ctx context.Context, arc *gar.Archive, code string) (map[int64]*object, error) {
	objects := make(map[int64]*object)
	err := scanFile(ctx, arc, code, gar.KindAddrObj, "OBJECT", func(s *gar.Scanner) {
		level, _ := s.Int("LEVEL")
		if level < gar.LevelRegion || level > gar.LevelSettlement || level == 4 {
			return
		}
		if !s.Is("ISACTUAL", "1") || !s.Is("ISACTIVE", "1") {
			return
		}
		id, ok := s.Int("OBJECTID")
		if !ok {
			return
		}
		guid, _ := uuid.ParseBytes(s.Raw("OBJECTGUID"))
		objects[id] = &object{
			guid:     guid,
			name:     strings.TrimSpace(s.String("NAME")),
			typeName: strings.TrimSpace(s.String("TYPENAME")),
			level:    level,
		}
	})
	return objects, err
}

// readHierarchy returns, for each target, its active administrative PATH as
// OBJECTIDs from the region down to the target itself.
func readHierarchy(ctx context.Context, arc *gar.Archive, code string, targets map[int64]*object) (map[int64][]int64, error) {
	paths := make(map[int64][]int64, len(targets))
	err := scanFile(ctx, arc, code, gar.KindAdmHierarchy, "ITEM", func(s *gar.Scanner) {
		// Most items are houses and steads; check the cheap key first.
		id, ok := s.Int("OBJECTID")
		if !ok || targets[id] == nil || !s.Is("ISACTIVE", "1") {
			return
		}
		var path []int64
		for part := range bytes.SplitSeq(s.Raw("PATH"), []byte{'.'}) {
			if n, ok := gar.ParseInt(part); ok {
				path = append(path, n)
			}
		}
		paths[id] = path
	})
	return paths, err
}

// readParams returns the current value of each wanted parameter per target.
func readParams(ctx context.Context, arc *gar.Archive, code string, targets map[int64]*object, today string) (map[int64]map[int64]param, error) {
	params := make(map[int64]map[int64]param, len(targets))
	err := scanFile(ctx, arc, code, gar.KindAddrObjParams, "PARAM", func(s *gar.Scanner) {
		id, ok := s.Int("OBJECTID")
		if !ok || targets[id] == nil {
			return
		}
		typeID, _ := s.Int("TYPEID")
		if !wantedParams[typeID] || string(s.Raw("ENDDATE")) <= today {
			return
		}
		start := string(s.Raw("STARTDATE"))
		m := params[id]
		if m == nil {
			m = make(map[int64]param)
			params[id] = m
		}
		if prev, ok := m[typeID]; ok && prev.startDate > start {
			return
		}
		m[typeID] = param{value: strings.TrimSpace(s.String("VALUE")), startDate: start}
	})
	return params, err
}

func scanFile(ctx context.Context, arc *gar.Archive, code, kind, tag string, fn func(*gar.Scanner)) error {
	rc, err := arc.OpenFile(ctx, code, kind)
	if errors.Is(err, gar.ErrNotFound) {
		return nil
	}
	if err != nil {
		return err
	}
	defer func() { _ = rc.Close() }()

	s := gar.NewScanner(&ctxReader{ctx: ctx, r: rc}, tag)
	for s.Next() {
		fn(s)
	}
	if err := s.Err(); err != nil {
		return fmt.Errorf("read %s/%s: %w", code, kind, err)
	}
	return nil
}

// ctxReader stops a long scan promptly when ctx is canceled.
type ctxReader struct {
	ctx context.Context //nolint:containedctx // adapts io.Reader to ctx
	r   io.Reader
}

func (c *ctxReader) Read(p []byte) (int, error) {
	if err := c.ctx.Err(); err != nil {
		return 0, err
	}
	return c.r.Read(p)
}

// address formats the row like hflabs/DaData: "Респ Адыгея, г Адыгейск",
// "Московская обл, Ленинский р-н, д Ивановка".
func address(r *row) string {
	parts := make([]string, 0, 4)
	// "респ. Адыгея", but "Удмуртская респ.": adjectives take the type after.
	if isPrefixType(r.RegionType) && !strings.HasSuffix(r.Region, "ая") {
		parts = append(parts, join(r.RegionType, r.Region))
	} else {
		parts = append(parts, join(r.Region, r.RegionType))
	}
	if r.Area != nil {
		if isPrefixType(*r.AreaType) {
			parts = append(parts, join(*r.AreaType, *r.Area))
		} else {
			parts = append(parts, join(*r.Area, *r.AreaType))
		}
	}
	if r.City != nil {
		parts = append(parts, join(*r.CityType, *r.City))
	}
	if r.Settlement != nil {
		parts = append(parts, join(*r.SettlementType, *r.Settlement))
	}
	return strings.Join(parts, ", ")
}

// isPrefixType reports whether a region/area type is written before the name
// ("респ. Адыгея", "г. Сочи") rather than after it ("Московская обл.").
func isPrefixType(t string) bool {
	t = strings.ToLower(strings.TrimSuffix(t, "."))
	return t == "респ" || t == "г" || t == "г.ф.з" || t == "г.о" || t == "м.о"
}

func join(a, b string) string {
	return strings.TrimSpace(a + " " + b)
}

func ptr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func paramPtr(p map[int64]param, typeID int64) *string {
	return ptr(p[typeID].value)
}

func isTrue(v string) bool {
	return v == "1" || strings.EqualFold(v, "true")
}
