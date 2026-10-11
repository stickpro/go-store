package handler

import (
	"context"
	"errors"
	"testing"

	"github.com/goccy/go-json"
	"github.com/stickpro/go-store/internal/dto"
	"github.com/stickpro/go-store/internal/messaging/contracts"
	"github.com/stickpro/go-store/internal/service/order"
	"github.com/stickpro/go-store/pkg/logger"
)

const testOrderID = "8f2f4f0c-3b0e-4f43-9a52-2d1a6b0d9c10"

// fakeOrders records what the handler passes to the order service.
type fakeOrders struct {
	order.IOrderService

	statuses  []dto.OrderERPStatusDTO
	trackings []dto.OrderERPTrackingDTO
	err       error
}

func (f *fakeOrders) ApplyERPStatus(_ context.Context, d dto.OrderERPStatusDTO) error {
	f.statuses = append(f.statuses, d)
	return f.err
}

func (f *fakeOrders) AssignERPTracking(_ context.Context, d dto.OrderERPTrackingDTO) error {
	f.trackings = append(f.trackings, d)
	return f.err
}

func envelope(t *testing.T, raw string) contracts.OrderEnvelope {
	t.Helper()
	var env contracts.OrderEnvelope
	if err := json.Unmarshal([]byte(raw), &env); err != nil {
		t.Fatalf("unmarshal envelope: %v", err)
	}
	return env
}

const statusEvent = `{
	"event_id": "5b0c1f0e-6f0a-4a57-9d3e-0c1c5a3f8a11",
	"event_type": "order.status_changed",
	"schema_version": 1,
	"occurred_at": "2026-10-11T12:30:00+03:00",
	"payload": {"order_id": "8f2f4f0c-3b0e-4f43-9a52-2d1a6b0d9c10", "external_id": "ЗК-000123", "status": "shipped", "comment": "", "changed_at": "2026-10-11T12:29:40+03:00"}
}`

const trackingEvent = `{
	"event_id": "0d9f6c9a-0a35-4b0e-8f3b-6a4a3c0f7b22",
	"event_type": "order.tracking_assigned",
	"schema_version": 1,
	"occurred_at": "2026-10-11T13:00:00+03:00",
	"payload": {"order_id": "8f2f4f0c-3b0e-4f43-9a52-2d1a6b0d9c10", "tracking_number": "10012345678", "provider": "cdek", "assigned_at": "2026-10-11T12:59:00+03:00"}
}`

func TestHandleErpOrder_StatusChanged(t *testing.T) {
	svc := &fakeOrders{}
	h := NewOrderHandler(svc, logger.New())

	if err := h.HandleErpOrder(context.Background(), envelope(t, statusEvent)); err != nil {
		t.Fatalf("handle: %v", err)
	}
	if len(svc.statuses) != 1 {
		t.Fatalf("ApplyERPStatus calls = %d, want 1", len(svc.statuses))
	}
	got := svc.statuses[0]
	if got.EventID.String() != "5b0c1f0e-6f0a-4a57-9d3e-0c1c5a3f8a11" || got.OrderID.String() != testOrderID || got.Status != "shipped" {
		t.Errorf("unexpected dto: %+v", got)
	}
	if got.ExternalID == nil || *got.ExternalID != "ЗК-000123" {
		t.Errorf("external_id = %v, want ЗК-000123", got.ExternalID)
	}
	if got.Comment != nil {
		t.Errorf("empty comment must map to nil, got %q", *got.Comment)
	}
}

func TestHandleErpOrder_TrackingAssigned(t *testing.T) {
	svc := &fakeOrders{}
	h := NewOrderHandler(svc, logger.New())

	if err := h.HandleErpOrder(context.Background(), envelope(t, trackingEvent)); err != nil {
		t.Fatalf("handle: %v", err)
	}
	if len(svc.trackings) != 1 {
		t.Fatalf("AssignERPTracking calls = %d, want 1", len(svc.trackings))
	}
	got := svc.trackings[0]
	if got.OrderID.String() != testOrderID || got.TrackingNumber != "10012345678" {
		t.Errorf("unexpected dto: %+v", got)
	}
	if got.Provider == nil || *got.Provider != "cdek" {
		t.Errorf("provider = %v, want cdek", got.Provider)
	}
	if got.TrackingURL != nil {
		t.Errorf("tracking_url = %v, want nil", got.TrackingURL)
	}
}

// Events the store can never apply are acknowledged without reaching the service.
func TestHandleErpOrder_InvalidEventSkipped(t *testing.T) {
	cases := map[string]string{
		"unknown type":      `{"event_id":"5b0c1f0e-6f0a-4a57-9d3e-0c1c5a3f8a11","event_type":"order.exploded","schema_version":1,"payload":{}}`,
		"missing event id":  `{"event_type":"order.status_changed","schema_version":1,"payload":{"order_id":"8f2f4f0c-3b0e-4f43-9a52-2d1a6b0d9c10","status":"shipped"}}`,
		"schema version":    `{"event_id":"5b0c1f0e-6f0a-4a57-9d3e-0c1c5a3f8a11","event_type":"order.status_changed","schema_version":2,"payload":{"order_id":"8f2f4f0c-3b0e-4f43-9a52-2d1a6b0d9c10","status":"shipped"}}`,
		"no order id":       `{"event_id":"5b0c1f0e-6f0a-4a57-9d3e-0c1c5a3f8a11","event_type":"order.status_changed","schema_version":1,"payload":{"status":"shipped"}}`,
		"no tracking":       `{"event_id":"5b0c1f0e-6f0a-4a57-9d3e-0c1c5a3f8a11","event_type":"order.tracking_assigned","schema_version":1,"payload":{"order_id":"8f2f4f0c-3b0e-4f43-9a52-2d1a6b0d9c10"}}`,
		"malformed payload": `{"event_id":"5b0c1f0e-6f0a-4a57-9d3e-0c1c5a3f8a11","event_type":"order.tracking_assigned","schema_version":1,"payload":{"order_id":"abc"}}`,
	}

	for name, raw := range cases {
		t.Run(name, func(t *testing.T) {
			svc := &fakeOrders{}
			h := NewOrderHandler(svc, logger.New())

			if err := h.HandleErpOrder(context.Background(), envelope(t, raw)); err != nil {
				t.Fatalf("handle: %v, want nil (skip)", err)
			}
			if len(svc.statuses)+len(svc.trackings) != 0 {
				t.Errorf("service must not be called for an invalid event")
			}
		})
	}
}

func TestHandleErpOrder_ServiceErrors(t *testing.T) {
	infra := errors.New("connection reset")
	cases := []struct {
		name    string
		err     error
		wantErr bool
	}{
		{"order not found is skipped", order.ErrNotFound, false},
		{"invalid transition is skipped", order.ErrInvalidTransition, false},
		{"infrastructure error is returned", infra, true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h := NewOrderHandler(&fakeOrders{err: tc.err}, logger.New())

			err := h.HandleErpOrder(context.Background(), envelope(t, statusEvent))
			if (err != nil) != tc.wantErr {
				t.Errorf("err = %v, wantErr = %v", err, tc.wantErr)
			}
		})
	}
}
