package contracts

import (
	"testing"

	"github.com/goccy/go-json"
)

func TestProductPayload_DimensionsUnmarshal(t *testing.T) {
	raw := `{
		"external_id": "19c9f92b-81de-11f1-b9d9-00155ddf6800",
		"price_retail": 1290,
		"is_enable": "true",
		"length": "8.4",
		"width": "7.4",
		"height": "3.4",
		"weight": "0.007"
	}`

	var p ProductPayload
	if err := json.Unmarshal([]byte(raw), &p); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if got := p.Length.String(); got != "8.4" {
		t.Errorf("length = %q, want 8.4", got)
	}
	if got := p.Width.String(); got != "7.4" {
		t.Errorf("width = %q, want 7.4", got)
	}
	if got := p.Height.String(); got != "3.4" {
		t.Errorf("height = %q, want 3.4", got)
	}
	if got := p.Weight.String(); got != "0.007" {
		t.Errorf("weight = %q, want 0.007", got)
	}
}

func TestFlexDecimal_EmptyAndMissing(t *testing.T) {
	var p ProductPayload
	if err := json.Unmarshal([]byte(`{"weight": "", "length": null}`), &p); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if !p.Weight.IsZero() || !p.Length.IsZero() || !p.Height.IsZero() {
		t.Errorf("expected zero decimals, got weight=%s length=%s height=%s",
			p.Weight, p.Length, p.Height)
	}
}
