// Package garimport loads every Russian settlement (cities, towns, villages…)
// from the ФНС ГАР БД ФИАС export into the cities table.
//
// Each region folder of gar_xml.zip is read independently: AS_ADDR_OBJ gives
// the active objects of levels region…settlement, AS_ADM_HIERARCHY the
// administrative path (region → district → city → settlement) and
// AS_ADDR_OBJ_PARAMS the postal code, OKATO, OKTMO, KLADR and tax office.
// Rows are upserted by fias_id, so a re-import keeps ids (and
// orders.ship_city_id) stable and never deletes rows that left GAR.
//
// GAR has no coordinates, population, timezone or foundation year. Those
// columns are left untouched on existing rows (e.g. filled by the hflabs
// import) and null on new ones.
package garimport

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/stickpro/go-store/pkg/gar"
	"github.com/stickpro/go-store/pkg/logger"
)

const (
	// regionAttempts bounds retries of one region when a remote stream drops.
	regionAttempts = 3
	retryDelay     = 5 * time.Second
)

// Options controls what a Run does.
type Options struct {
	// Source is a local gar_xml.zip path or its URL. Empty means the latest
	// full export, read remotely.
	Source string
	// Regions limits the import to these folder codes ("01", "77"); empty
	// means all.
	Regions []string
	// Workers is how many regions are read concurrently.
	Workers int
	DryRun  bool
}

// Report summarizes what a Run did.
type Report struct {
	Source      string
	Regions     int
	Settlements int
	Inserted    int64
	Updated     int64
	// NoHierarchy counts settlements without an administrative path; they
	// are imported with only the region filled in.
	NoHierarchy int
	Skipped     []string
}

// Run imports the settlements into the cities table of pool.
func Run(ctx context.Context, pool *pgxpool.Pool, l logger.Logger, opts Options) (*Report, error) {
	client := &http.Client{Timeout: 0} // region streams run for minutes; ctx bounds them
	src := opts.Source
	if src == "" {
		var err error
		if src, err = gar.LatestFullURL(ctx, client); err != nil {
			return nil, err
		}
	}
	report := &Report{Source: src}
	l.Infow("opening gar archive", "source", src)

	arc, err := gar.Open(ctx, src, client)
	if err != nil {
		return report, err
	}
	defer func() { _ = arc.Close() }()

	codes, err := selectRegions(arc.Regions(), opts.Regions)
	if err != nil {
		return report, err
	}

	today := time.Now().Format(time.DateOnly)
	results, skipped, err := loadRegions(ctx, arc, l, codes, today, max(opts.Workers, 1))
	report.Skipped = skipped
	if err != nil {
		return report, err
	}

	var rows []row
	seen := make(map[uuid.UUID]bool)
	for _, res := range results {
		report.Regions++
		report.NoHierarchy += res.noHierarchy
		// ON CONFLICT can't touch the same row twice in one statement.
		for _, r := range res.rows {
			if !seen[r.FiasID] {
				seen[r.FiasID] = true
				rows = append(rows, r)
			}
		}
	}
	report.Settlements = len(rows)
	if len(rows) == 0 {
		return report, ErrNoRows
	}
	if opts.DryRun {
		return report, nil
	}

	report.Inserted, report.Updated, err = upsert(ctx, pool, rows)
	return report, err
}

func selectRegions(available, wanted []string) ([]string, error) {
	if len(wanted) == 0 {
		return available, nil
	}
	for _, c := range wanted {
		if !slices.Contains(available, c) {
			return nil, fmt.Errorf("region %q is not in the archive (have %v)", c, available)
		}
	}
	return wanted, nil
}

func loadRegions(ctx context.Context, arc *gar.Archive, l logger.Logger, codes []string, today string, workers int) ([]*regionResult, []string, error) {
	var (
		mu       sync.Mutex
		results  []*regionResult
		skipped  []string
		firstErr error
		wg       sync.WaitGroup
	)
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	jobs := make(chan string)
	for range workers {
		wg.Go(func() {
			for code := range jobs {
				district, ok := federalDistricts[code]
				if !ok {
					mu.Lock()
					skipped = append(skipped, code)
					mu.Unlock()
					l.Infow("skipping region without federal district", "region", code)
					continue
				}
				started := time.Now()
				res, err := loadRegionWithRetry(ctx, arc, l, code, district, today)
				mu.Lock()
				if err != nil {
					if firstErr == nil {
						firstErr = err
						cancel()
					}
				} else {
					results = append(results, res)
				}
				mu.Unlock()
				if err == nil {
					l.Infow("region loaded",
						"region", code, "name", res.regionName, "settlements", len(res.rows),
						"no_hierarchy", res.noHierarchy, "compressed_mb", res.compressedLen>>20,
						"took", time.Since(started).Round(time.Second).String())
				}
			}
		})
	}

feed:
	for _, code := range codes {
		select {
		case jobs <- code:
		case <-ctx.Done():
			break feed
		}
	}
	close(jobs)
	wg.Wait()

	if firstErr != nil {
		return nil, skipped, firstErr
	}
	if err := ctx.Err(); err != nil {
		return nil, skipped, err
	}
	sort.Slice(results, func(i, j int) bool { return results[i].code < results[j].code })
	sort.Strings(skipped)
	return results, skipped, nil
}

func loadRegionWithRetry(ctx context.Context, arc *gar.Archive, l logger.Logger, code, district, today string) (*regionResult, error) {
	var err error
	for attempt := 1; attempt <= regionAttempts; attempt++ {
		var res *regionResult
		if res, err = loadRegion(ctx, arc, code, district, today); err == nil {
			return res, nil
		}
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		l.Warnw("region load failed", "region", code, "attempt", attempt, "error", err)
		select {
		case <-time.After(retryDelay):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
	return nil, fmt.Errorf("region %s: %w", code, err)
}

var upsertColumns = []string{
	"address", "postal_code", "country", "federal_district", "region_type", "region",
	"area_type", "area", "city_type", "city", "settlement_type", "settlement",
	"kladr_id", "fias_id", "fias_level", "capital_marker", "okato", "oktmo", "tax_office",
}

// upsertSQL inserts new settlements and refreshes GAR-owned columns of
// existing ones. Codes GAR doesn't have for a settlement and a zero capital
// marker don't overwrite values another source (hflabs) provided.
const upsertSQL = `
INSERT INTO cities (address, postal_code, country, federal_district, region_type, region,
                    area_type, area, city_type, city, settlement_type, settlement,
                    kladr_id, fias_id, fias_level, capital_marker, okato, oktmo, tax_office)
SELECT address, postal_code, country, federal_district, region_type, region,
       area_type, area, city_type, city, settlement_type, settlement,
       kladr_id, fias_id, fias_level, capital_marker, okato, oktmo, tax_office
FROM cities_gar_import
ON CONFLICT (fias_id) DO UPDATE SET
    address          = excluded.address,
    postal_code      = coalesce(excluded.postal_code, cities.postal_code),
    country          = excluded.country,
    federal_district = excluded.federal_district,
    region_type      = excluded.region_type,
    region           = excluded.region,
    area_type        = excluded.area_type,
    area             = excluded.area,
    city_type        = excluded.city_type,
    city             = excluded.city,
    settlement_type  = excluded.settlement_type,
    settlement       = excluded.settlement,
    kladr_id         = coalesce(excluded.kladr_id, cities.kladr_id),
    fias_level       = excluded.fias_level,
    capital_marker   = CASE WHEN excluded.capital_marker > 0 THEN excluded.capital_marker ELSE cities.capital_marker END,
    okato            = coalesce(excluded.okato, cities.okato),
    oktmo            = coalesce(excluded.oktmo, cities.oktmo),
    tax_office       = coalesce(excluded.tax_office, cities.tax_office)
RETURNING (xmax = 0) AS inserted`

func upsert(ctx context.Context, pool *pgxpool.Pool, rows []row) (inserted, updated int64, err error) {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return 0, 0, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if _, err = tx.Exec(ctx, `CREATE TEMP TABLE cities_gar_import (LIKE cities INCLUDING DEFAULTS) ON COMMIT DROP`); err != nil {
		return 0, 0, fmt.Errorf("create temp table: %w", err)
	}
	_, err = tx.CopyFrom(ctx, pgx.Identifier{"cities_gar_import"}, upsertColumns,
		pgx.CopyFromSlice(len(rows), func(i int) ([]any, error) {
			r := &rows[i]
			return []any{
				r.Address, r.PostalCode, country, r.FederalDistrict, r.RegionType, r.Region,
				r.AreaType, r.Area, r.CityType, r.City, r.SettlementType, r.Settlement,
				r.KladrID, r.FiasID, r.FiasLevel, r.CapitalMarker, r.Okato, r.Oktmo, r.TaxOffice,
			}, nil
		}))
	if err != nil {
		return 0, 0, fmt.Errorf("copy settlements: %w", err)
	}

	res, err := tx.Query(ctx, upsertSQL)
	if err != nil {
		return 0, 0, fmt.Errorf("upsert settlements: %w", err)
	}
	for res.Next() {
		var isNew bool
		if err := res.Scan(&isNew); err != nil {
			res.Close()
			return 0, 0, err
		}
		if isNew {
			inserted++
		} else {
			updated++
		}
	}
	res.Close()
	if err := res.Err(); err != nil {
		return 0, 0, fmt.Errorf("upsert settlements: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return 0, 0, err
	}
	return inserted, updated, nil
}

// ErrNoRows is returned when the selected regions held no settlements.
var ErrNoRows = errors.New("no settlements found")
