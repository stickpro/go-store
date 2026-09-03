package order

import (
	"context"

	"github.com/stickpro/go-store/internal/messaging/contracts"
)

// EventPublisher ships order domain events to downstream consumers (email,
// analytics, ERP sync). Implementations must be safe to call after the order
// transaction has committed; a publish failure is logged, not surfaced to the
// customer.
type EventPublisher interface {
	OrderCreated(ctx context.Context, e contracts.OrderCreatedPayload) error
	OrderPaid(ctx context.Context, e contracts.OrderPaidPayload) error
	OrderCancelled(ctx context.Context, e contracts.OrderCancelledPayload) error
}

// NoopPublisher drops every event. Used until the Kafka producer is wired into
// the DI graph.
type NoopPublisher struct{}

func (NoopPublisher) OrderCreated(context.Context, contracts.OrderCreatedPayload) error { return nil }
func (NoopPublisher) OrderPaid(context.Context, contracts.OrderPaidPayload) error       { return nil }
func (NoopPublisher) OrderCancelled(context.Context, contracts.OrderCancelledPayload) error {
	return nil
}
