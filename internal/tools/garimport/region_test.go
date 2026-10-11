package garimport

import (
	"archive/zip"
	"context"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/stickpro/go-store/pkg/gar"
)

const (
	guidRegion   = "11111111-1111-1111-1111-111111111111"
	guidArea     = "22222222-2222-2222-2222-222222222222"
	guidCity     = "33333333-3333-3333-3333-333333333333"
	guidVillage  = "44444444-4444-4444-4444-444444444444"
	guidInCity   = "55555555-5555-5555-5555-555555555555"
	guidOrphan   = "66666666-6666-6666-6666-666666666666"
	guidStreet   = "77777777-7777-7777-7777-777777777777"
	guidReplaced = "88888888-8888-8888-8888-888888888888"
)

const addrObj = `<?xml version="1.0" encoding="utf-8"?><ADDRESSOBJECTS>` +
	`<OBJECT ID="1" OBJECTID="100" OBJECTGUID="` + guidRegion + `" NAME="Тверская" TYPENAME="обл." LEVEL="1" ISACTUAL="1" ISACTIVE="1" />` +
	`<OBJECT ID="2" OBJECTID="200" OBJECTGUID="` + guidArea + `" NAME="Калининский" TYPENAME="р-н" LEVEL="2" ISACTUAL="1" ISACTIVE="1" />` +
	`<OBJECT ID="3" OBJECTID="300" OBJECTGUID="` + guidCity + `" NAME="Тверь" TYPENAME="г." LEVEL="5" ISACTUAL="1" ISACTIVE="1" />` +
	`<OBJECT ID="4" OBJECTID="400" OBJECTGUID="` + guidVillage + `" NAME="Москва" TYPENAME="д." LEVEL="6" ISACTUAL="1" ISACTIVE="1" />` +
	`<OBJECT ID="5" OBJECTID="500" OBJECTGUID="` + guidInCity + `" NAME="Химинститут" TYPENAME="п." LEVEL="6" ISACTUAL="1" ISACTIVE="1" />` +
	`<OBJECT ID="6" OBJECTID="600" OBJECTGUID="` + guidOrphan + `" NAME="Сирота" TYPENAME="с." LEVEL="6" ISACTUAL="1" ISACTIVE="1" />` +
	`<OBJECT ID="7" OBJECTID="700" OBJECTGUID="` + guidStreet + `" NAME="Ленина" TYPENAME="ул." LEVEL="8" ISACTUAL="1" ISACTIVE="1" />` +
	`<OBJECT ID="8" OBJECTID="800" OBJECTGUID="` + guidReplaced + `" NAME="Старое" TYPENAME="д." LEVEL="6" ISACTUAL="0" ISACTIVE="0" />` +
	`</ADDRESSOBJECTS>`

const admHierarchy = `<?xml version="1.0" encoding="utf-8"?><ITEMS>` +
	`<ITEM ID="1" OBJECTID="300" PARENTOBJID="100" ISACTIVE="1" PATH="100.300" />` +
	`<ITEM ID="2" OBJECTID="400" PARENTOBJID="300" ISACTIVE="0" PATH="100.300.400" />` +
	`<ITEM ID="3" OBJECTID="400" PARENTOBJID="200" ISACTIVE="1" PATH="100.200.400" />` +
	`<ITEM ID="4" OBJECTID="500" PARENTOBJID="300" ISACTIVE="1" PATH="100.300.500" />` +
	`<ITEM ID="5" OBJECTID="700" PARENTOBJID="500" ISACTIVE="1" PATH="100.300.500.700" />` +
	`</ITEMS>`

const addrObjParams = `<?xml version="1.0" encoding="utf-8"?><PARAMS>` +
	`<PARAM ID="1" OBJECTID="300" TYPEID="5" VALUE="170000" STARTDATE="2011-01-01" ENDDATE="2079-06-06" />` +
	`<PARAM ID="2" OBJECTID="300" TYPEID="7" VALUE="28701000" STARTDATE="2011-01-01" ENDDATE="2015-01-01" />` +
	`<PARAM ID="3" OBJECTID="300" TYPEID="7" VALUE="28701000001" STARTDATE="2015-01-01" ENDDATE="2079-06-06" />` +
	`<PARAM ID="4" OBJECTID="300" TYPEID="10" VALUE="6900000100000" STARTDATE="2011-01-01" ENDDATE="2079-06-06" />` +
	`<PARAM ID="5" OBJECTID="300" TYPEID="22" VALUE="1" STARTDATE="2024-09-28" ENDDATE="2079-06-06" />` +
	`<PARAM ID="6" OBJECTID="300" TYPEID="1" VALUE="6952" STARTDATE="2011-01-01" ENDDATE="2079-06-06" />` +
	`<PARAM ID="7" OBJECTID="400" TYPEID="6" VALUE="28220000001" STARTDATE="2011-01-01" ENDDATE="2079-06-06" />` +
	`<PARAM ID="8" OBJECTID="400" TYPEID="23" VALUE="1" STARTDATE="2024-09-28" ENDDATE="2079-06-06" />` +
	`<PARAM ID="9" OBJECTID="700" TYPEID="5" VALUE="170001" STARTDATE="2011-01-01" ENDDATE="2079-06-06" />` +
	`</PARAMS>`

func must(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

func equal[T comparable](t *testing.T, got, want T, msg string) {
	t.Helper()
	if got != want {
		t.Fatalf("%s: got %v, want %v", msg, got, want)
	}
}

func str(t *testing.T, got *string, want, msg string) {
	t.Helper()
	if got == nil {
		t.Fatalf("%s: got nil, want %q", msg, want)
	}
	equal(t, *got, want, msg)
}

func null(t *testing.T, got *string, msg string) {
	t.Helper()
	if got != nil {
		t.Fatalf("%s: got %q, want nil", msg, *got)
	}
}

func writeArchive(t *testing.T, files map[string]string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "gar_xml.zip")
	f, err := os.Create(path)
	must(t, err)
	zw := zip.NewWriter(f)
	for name, body := range files {
		w, err := zw.Create(name)
		must(t, err)
		_, err = w.Write([]byte(body))
		must(t, err)
	}
	must(t, zw.Close())
	must(t, f.Close())
	return path
}

func TestLoadRegion(t *testing.T) {
	path := writeArchive(t, map[string]string{
		"version.txt": "2026.10.09",
		"69/AS_ADDR_OBJ_20261009_0a1b2c3d-0000-0000-0000-000000000001.XML":        addrObj,
		"69/AS_ADM_HIERARCHY_20261009_0a1b2c3d-0000-0000-0000-000000000002.XML":   admHierarchy,
		"69/AS_ADDR_OBJ_PARAMS_20261009_0a1b2c3d-0000-0000-0000-000000000003.XML": addrObjParams,
	})
	arc, err := gar.Open(context.Background(), path, http.DefaultClient)
	must(t, err)
	defer func() { _ = arc.Close() }()
	equal(t, slices.Equal(arc.Regions(), []string{"69"}), true, "arc.Regions()")

	res, err := loadRegion(context.Background(), arc, "69", "Центральный", "2026-10-09")
	must(t, err)
	equal(t, res.regionName, "Тверская", "res.regionName")
	equal(t, res.federalCity, false, "res.federalCity")
	equal(t, res.noHierarchy, 1, "res.noHierarchy")

	byGUID := make(map[string]row)
	for _, r := range res.rows {
		byGUID[r.FiasID.String()] = r
	}
	equal(t, len(byGUID), 4, "street, region and inactive objects are not settlements")

	city := byGUID[guidCity]
	equal(t, city.Address, "Тверская обл., г. Тверь", "city.Address")
	str(t, city.City, "Тверь", "city.City")
	null(t, city.Settlement, "city.Settlement")
	null(t, city.Area, "city.Area")
	equal(t, city.FiasLevel, int16(fiasLevelCity), "city.FiasLevel")
	equal(t, city.CapitalMarker, int16(capitalRegion), "city.CapitalMarker")
	str(t, city.PostalCode, "170000", "city.PostalCode")
	str(t, city.Oktmo, "28701000001", "the ended OKTMO value is ignored")
	str(t, city.KladrID, "6900000100000", "city.KladrID")
	str(t, city.TaxOffice, "6952", "city.TaxOffice")
	null(t, city.Okato, "city.Okato")
	equal(t, city.FederalDistrict, "Центральный", "city.FederalDistrict")

	village := byGUID[guidVillage]
	equal(t, village.Address, "Тверская обл., Калининский р-н, д. Москва", "only the active hierarchy item counts")
	null(t, village.City, "village.City")
	str(t, village.Settlement, "Москва", "village.Settlement")
	equal(t, village.FiasLevel, int16(fiasLevelSettlement), "village.FiasLevel")
	equal(t, village.CapitalMarker, int16(capitalArea), "village.CapitalMarker")
	null(t, village.PostalCode, "village.PostalCode")

	inCity := byGUID[guidInCity]
	equal(t, inCity.Address, "Тверская обл., г. Тверь, п. Химинститут", "inCity.Address")
	str(t, inCity.City, "Тверь", "inCity.City")
	str(t, inCity.Settlement, "Химинститут", "inCity.Settlement")

	orphan := byGUID[guidOrphan]
	equal(t, orphan.Address, "Тверская обл., с. Сирота", "orphan.Address")
}

func TestAddressRegionAdjective(t *testing.T) {
	city := "Ижевск"
	cityType := "г."
	r := row{RegionType: "респ.", Region: "Удмуртская", City: &city, CityType: &cityType}
	equal(t, address(&r), "Удмуртская респ., г. Ижевск", "adjective region")
	r.Region = "Адыгея"
	equal(t, address(&r), "респ. Адыгея, г. Ижевск", "noun region")
}

func TestLoadRegionFederalCity(t *testing.T) {
	path := writeArchive(t, map[string]string{
		"77/AS_ADDR_OBJ_20261009_0a1b2c3d-0000-0000-0000-000000000001.XML": `<ADDRESSOBJECTS>` +
			`<OBJECT OBJECTID="1" OBJECTGUID="0c5b2444-70a0-4932-980c-b4dc0d3f02b5" NAME="Москва" TYPENAME="г." LEVEL="1" ISACTUAL="1" ISACTIVE="1" />` +
			`<OBJECT OBJECTID="2" OBJECTGUID="` + guidCity + `" NAME="Зеленоград" TYPENAME="г." LEVEL="5" ISACTUAL="1" ISACTIVE="1" />` +
			`</ADDRESSOBJECTS>`,
		"77/AS_ADM_HIERARCHY_20261009_0a1b2c3d-0000-0000-0000-000000000002.XML": `<ITEMS>` +
			`<ITEM OBJECTID="2" ISACTIVE="1" PATH="1.2" />` +
			`</ITEMS>`,
	})
	arc, err := gar.Open(context.Background(), path, http.DefaultClient)
	must(t, err)
	defer func() { _ = arc.Close() }()

	res, err := loadRegion(context.Background(), arc, "77", "Центральный", "2026-10-09")
	must(t, err)
	equal(t, res.federalCity, true, "res.federalCity")
	equal(t, len(res.rows), 2, "len")

	for _, r := range res.rows {
		if r.FiasID == uuid.MustParse("0c5b2444-70a0-4932-980c-b4dc0d3f02b5") {
			equal(t, r.Address, "г. Москва", "r.Address")
			null(t, r.City, "r.City")
			equal(t, r.FiasLevel, int16(fiasLevelRegion), "r.FiasLevel")
		} else {
			equal(t, r.Address, "г. Москва, г. Зеленоград", "r.Address")
		}
	}
}
