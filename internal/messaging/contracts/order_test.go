package contracts

import (
	"testing"

	"github.com/goccy/go-json"
	"github.com/stickpro/go-store/internal/constant"
)

func TestOrderEnvelope_StatusChangedUnmarshal(t *testing.T) {
	raw := `{
		"event_id": "5b0c1f0e-6f0a-4a57-9d3e-0c1c5a3f8a11",
		"event_type": "order.status_changed",
		"schema_version": 1,
		"occurred_at": "2026-10-11T12:30:00+03:00",
		"payload": {
			"order_id": "8f2f4f0c-3b0e-4f43-9a52-2d1a6b0d9c10",
			"external_id": "ЗК-000123",
			"status": "shipped",
			"changed_at": "2026-10-11T12:29:40+03:00"
		}
	}`

	var env OrderEnvelope
	if err := json.Unmarshal([]byte(raw), &env); err != nil {
		t.Fatalf("unmarshal envelope: %v", err)
	}
	if env.EventType != OrderEventStatusChanged {
		t.Fatalf("event_type = %q, want %q", env.EventType, OrderEventStatusChanged)
	}

	var p OrderStatusChangedPayload
	if err := json.Unmarshal(env.Payload, &p); err != nil {
		t.Fatalf("unmarshal payload: %v", err)
	}
	if got := p.OrderID.String(); got != "8f2f4f0c-3b0e-4f43-9a52-2d1a6b0d9c10" {
		t.Errorf("order_id = %s, want 8f2f4f0c-3b0e-4f43-9a52-2d1a6b0d9c10", got)
	}
	if p.Status != constant.OrderShipped {
		t.Errorf("status = %q, want %q", p.Status, constant.OrderShipped)
	}
	if p.ExternalID == nil || *p.ExternalID != "ЗК-000123" {
		t.Errorf("external_id = %v, want ЗК-000123", p.ExternalID)
	}
	if p.Comment != nil {
		t.Errorf("comment = %v, want nil", p.Comment)
	}
}

func TestOrderEnvelope_TrackingAssignedUnmarshal(t *testing.T) {
	raw := `{
		"event_id": "0d9f6c9a-0a35-4b0e-8f3b-6a4a3c0f7b22",
		"event_type": "order.tracking_assigned",
		"schema_version": 1,
		"occurred_at": "2026-10-11T13:00:00+03:00",
		"payload": {
			"order_id": "8f2f4f0c-3b0e-4f43-9a52-2d1a6b0d9c10",
			"tracking_number": "10012345678",
			"provider": "cdek",
			"assigned_at": "2026-10-11T12:59:00+03:00"
		}
	}`

	var env OrderEnvelope
	if err := json.Unmarshal([]byte(raw), &env); err != nil {
		t.Fatalf("unmarshal envelope: %v", err)
	}

	var p OrderTrackingAssignedPayload
	if err := json.Unmarshal(env.Payload, &p); err != nil {
		t.Fatalf("unmarshal payload: %v", err)
	}
	if p.TrackingNumber != "10012345678" {
		t.Errorf("tracking_number = %q, want 10012345678", p.TrackingNumber)
	}
	if p.Provider == nil || *p.Provider != "cdek" {
		t.Errorf("provider = %v, want cdek", p.Provider)
	}
	if p.TrackingURL != nil {
		t.Errorf("tracking_url = %v, want nil", p.TrackingURL)
	}
}
