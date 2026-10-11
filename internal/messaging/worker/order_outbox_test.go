package worker

import (
	"testing"
	"time"

	"github.com/goccy/go-json"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stickpro/go-store/internal/messaging/contracts"
	"github.com/stickpro/go-store/internal/models"
)

// The relay must put an outbox row on the wire as the documented envelope,
// with the stored payload embedded as-is.
func TestOrderEnvelope(t *testing.T) {
	occurred := time.Date(2026, 10, 11, 12, 30, 0, 0, time.UTC)
	row := &models.OrderOutbox{
		ID:            7,
		EventID:       uuid.MustParse("5b0c1f0e-6f0a-4a57-9d3e-0c1c5a3f8a11"),
		EventType:     contracts.OrderEventPaid,
		SchemaVersion: contracts.OrderSchemaVersion,
		OrderID:       uuid.MustParse("8f2f4f0c-3b0e-4f43-9a52-2d1a6b0d9c10"),
		Payload:       []byte(`{"order_id":"8f2f4f0c-3b0e-4f43-9a52-2d1a6b0d9c10","paid_at":"2026-10-11T12:29:00Z"}`),
		OccurredAt:    pgtype.Timestamptz{Time: occurred, Valid: true},
	}

	data, err := json.Marshal(orderEnvelope(row))
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var got struct {
		EventID       string `json:"event_id"`
		EventType     string `json:"event_type"`
		SchemaVersion int    `json:"schema_version"`
		OccurredAt    string `json:"occurred_at"`
		Payload       struct {
			OrderID string `json:"order_id"`
		} `json:"payload"`
	}
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if got.EventID != row.EventID.String() {
		t.Errorf("event_id = %q, want %q", got.EventID, row.EventID)
	}
	if got.EventType != contracts.OrderEventPaid {
		t.Errorf("event_type = %q, want %q", got.EventType, contracts.OrderEventPaid)
	}
	if got.SchemaVersion != contracts.OrderSchemaVersion {
		t.Errorf("schema_version = %d, want %d", got.SchemaVersion, contracts.OrderSchemaVersion)
	}
	if got.OccurredAt != "2026-10-11T12:30:00Z" {
		t.Errorf("occurred_at = %q, want 2026-10-11T12:30:00Z", got.OccurredAt)
	}
	if got.Payload.OrderID != row.OrderID.String() {
		t.Errorf("payload.order_id = %q, want %q", got.Payload.OrderID, row.OrderID)
	}
}
