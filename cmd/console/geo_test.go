package console

import (
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stickpro/go-store/internal/dto"
)

func TestCdekCoordRows(t *testing.T) {
	const fias = "267fd50d-0077-4041-9594-18b1de5f2acb"
	rows := cdekCoordRows([]dto.CDEKCityDTO{
		{FiasGUID: fias, Latitude: 65.994148, Longitude: 57.55701, TimeZone: "Europe/Moscow"},
		{FiasGUID: fias, Latitude: 1, Longitude: 1},        // duplicate fias_guid
		{FiasGUID: "", Latitude: 55, Longitude: 37},        // not linked to ФИАС
		{FiasGUID: "aeedcbcc-8ca1-4cd8-b961-6d3c7d355197"}, // no coordinates
		{FiasGUID: "1ac46b49-3209-4814-b7bf-a509ea1aecd9", Latitude: 55.3, Longitude: 79.7, TimeZone: "No/Such_Zone"},
	})
	if len(rows) != 2 {
		t.Fatalf("got %d rows, want 2", len(rows))
	}
	if lat, ok := rows[0][1].(decimal.Decimal); !ok || !lat.Equal(decimal.RequireFromString("65.994148")) {
		t.Fatalf("lat = %v", rows[0][1])
	}
	if tz, ok := rows[0][3].(*string); !ok || tz == nil || *tz != "UTC+3" {
		t.Fatalf("timezone = %v, want UTC+3", rows[0][3])
	}
	if tz, ok := rows[1][3].(*string); !ok || tz != nil {
		t.Fatalf("unknown zone must give nil timezone, got %v", rows[1][3])
	}
}

func TestUTCOffset(t *testing.T) {
	at := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	for zone, want := range map[string]string{
		"Europe/Kaliningrad": "UTC+2", "Asia/Novosibirsk": "UTC+7", "Asia/Kamchatka": "UTC+12",
	} {
		got := utcOffset(zone, at)
		if got == nil || *got != want {
			t.Fatalf("%s: got %v, want %s", zone, got, want)
		}
	}
	if utcOffset("", at) != nil {
		t.Fatal("empty zone must give nil")
	}
}
