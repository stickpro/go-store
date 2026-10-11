package handler

import (
	"context"
	"errors"
	"fmt"
	"unicode/utf8"

	"github.com/goccy/go-json"
	"github.com/google/uuid"
	"github.com/stickpro/go-store/internal/dto"
	"github.com/stickpro/go-store/internal/messaging/contracts"
	"github.com/stickpro/go-store/internal/service/order"
	"github.com/stickpro/go-store/pkg/logger"
)

type OrderHandler struct {
	svc    order.IOrderService
	logger logger.Logger
}

func NewOrderHandler(svc order.IOrderService, log logger.Logger) *OrderHandler {
	return &OrderHandler{svc: svc, logger: log}
}

// HandleErpOrder applies an order event sent by the ERP (status change or
// tracking number). It returns an error only for a failure worth redelivering;
// an event the store can never apply — malformed, unknown type, unknown order,
// forbidden transition — is logged and acknowledged, since a retry would fail
// the same way.
func (h *OrderHandler) HandleErpOrder(ctx context.Context, env contracts.OrderEnvelope) error {
	err := h.apply(ctx, env)
	if err == nil {
		return nil
	}
	if errors.Is(err, errSkipEvent) || errors.Is(err, order.ErrNotFound) || errors.Is(err, order.ErrInvalidTransition) {
		h.logger.Warnw("kafka: erp order event skipped",
			"event_id", env.EventID,
			"event_type", env.EventType,
			"error", err,
		)
		return nil
	}
	return err
}

// Column widths of the fields the ERP fills in (orders.external_id,
// tracking_number, tracking_provider, tracking_url).
const (
	maxExternalIDLen       = 128
	maxTrackingNumberLen   = 128
	maxTrackingProviderLen = 32
	maxTrackingURLLen      = 512
)

// errSkipEvent marks an event that is invalid by itself, regardless of store state.
var errSkipEvent = errors.New("invalid erp order event")

func (h *OrderHandler) apply(ctx context.Context, env contracts.OrderEnvelope) error {
	if env.EventID == uuid.Nil {
		return fmt.Errorf("%w: event_id is required", errSkipEvent)
	}
	if env.SchemaVersion != contracts.OrderSchemaVersion {
		return fmt.Errorf("%w: unsupported schema_version %d", errSkipEvent, env.SchemaVersion)
	}

	switch env.EventType {
	case contracts.OrderEventStatusChanged:
		var p contracts.OrderStatusChangedPayload
		if err := json.Unmarshal(env.Payload, &p); err != nil {
			return fmt.Errorf("%w: payload: %w", errSkipEvent, err)
		}
		if p.OrderID == uuid.Nil {
			return fmt.Errorf("%w: order_id is required", errSkipEvent)
		}
		if tooLong(p.ExternalID, maxExternalIDLen) {
			return fmt.Errorf("%w: external_id is longer than %d", errSkipEvent, maxExternalIDLen)
		}
		return h.svc.ApplyERPStatus(ctx, dto.OrderERPStatusDTO{
			EventID:    env.EventID,
			OrderID:    p.OrderID,
			ExternalID: nonEmpty(p.ExternalID),
			Status:     p.Status.String(),
			Comment:    nonEmpty(p.Comment),
		})
	case contracts.OrderEventTrackingAssigned:
		var p contracts.OrderTrackingAssignedPayload
		if err := json.Unmarshal(env.Payload, &p); err != nil {
			return fmt.Errorf("%w: payload: %w", errSkipEvent, err)
		}
		if p.OrderID == uuid.Nil {
			return fmt.Errorf("%w: order_id is required", errSkipEvent)
		}
		if p.TrackingNumber == "" {
			return fmt.Errorf("%w: tracking_number is required", errSkipEvent)
		}
		if tooLong(&p.TrackingNumber, maxTrackingNumberLen) ||
			tooLong(p.Provider, maxTrackingProviderLen) ||
			tooLong(p.TrackingURL, maxTrackingURLLen) {
			return fmt.Errorf("%w: tracking field is too long", errSkipEvent)
		}
		return h.svc.AssignERPTracking(ctx, dto.OrderERPTrackingDTO{
			EventID:        env.EventID,
			OrderID:        p.OrderID,
			TrackingNumber: p.TrackingNumber,
			Provider:       nonEmpty(p.Provider),
			TrackingURL:    nonEmpty(p.TrackingURL),
		})
	default:
		return fmt.Errorf("%w: unknown event_type %q", errSkipEvent, env.EventType)
	}
}

// tooLong reports whether s holds more than limit characters.
func tooLong(s *string, limit int) bool {
	return s != nil && utf8.RuneCountInString(*s) > limit
}

// nonEmpty treats an empty string the same as an absent value.
func nonEmpty(s *string) *string {
	if s == nil || *s == "" {
		return nil
	}
	return s
}
