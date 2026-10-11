package contracts

import (
	"time"

	"github.com/goccy/go-json"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stickpro/go-store/internal/constant"
)

// OrderSchemaVersion is the current version of the order exchange contract,
// carried in OrderEnvelope.SchemaVersion. Bump it on a breaking change only;
// adding an optional field is not one.
const OrderSchemaVersion = 1

// Order event types carried in OrderEnvelope.EventType.
const (
	// Store -> ERP (topics.Orders).
	OrderEventCreated   = "order.created"
	OrderEventUpdated   = "order.updated"
	OrderEventPaid      = "order.paid"
	OrderEventCancelled = "order.cancelled"

	// ERP -> store (topics.ErpOrders).
	OrderEventStatusChanged    = "order.status_changed"
	OrderEventTrackingAssigned = "order.tracking_assigned"
)

// OrderEnvelope wraps every message of the order exchange, in both directions.
// The Kafka message key is the order id, so events of one order stay ordered
// within a partition.
type OrderEnvelope struct {
	// EventID is unique per event; the receiver deduplicates on it.
	EventID       uuid.UUID `json:"event_id"`
	EventType     string    `json:"event_type"`
	SchemaVersion int       `json:"schema_version"`
	OccurredAt    time.Time `json:"occurred_at"`
	// Payload is the event body; its shape is selected by EventType.
	Payload json.RawMessage `json:"payload"`
}

type OrderEventItem struct {
	ProductID *uuid.UUID `json:"product_id,omitempty"`
	VariantID *uuid.UUID `json:"variant_id,omitempty"`
	// ExternalID is products.external_id — the key the ERP matches its
	// nomenclature on. Empty for a product that did not come from the ERP.
	ExternalID *string         `json:"external_id,omitempty"`
	Sku        *string         `json:"sku,omitempty"`
	Name       string          `json:"name"`
	UnitPrice  decimal.Decimal `json:"unit_price"`
	Quantity   int64           `json:"quantity"`
	LineTotal  decimal.Decimal `json:"line_total"`
}

type OrderEventShipping struct {
	CityName  string  `json:"city_name"`
	Address   string  `json:"address"`
	Postcode  *string `json:"postcode,omitempty"`
	Recipient string  `json:"recipient"`
	Method    *string `json:"method,omitempty"`
	// Provider / TariffCode / PointCode are the carrier snapshot picked at
	// checkout (e.g. "cdek" + tariff + pickup point).
	Provider   *string `json:"provider,omitempty"`
	TariffCode *string `json:"tariff_code,omitempty"`
	PointCode  *string `json:"point_code,omitempty"`
}

// OrderSnapshotPayload is the whole order as it stands after the change. It is
// the body of both order.created and order.updated: the receiver overwrites
// its copy with it rather than applying a delta.
type OrderSnapshotPayload struct {
	// OrderID identifies the order in every event of the exchange.
	OrderID uuid.UUID `json:"order_id"`
	// OrderNumber is the human-facing number; informational only.
	OrderNumber int64 `json:"order_number"`
	// Version grows with every change to the order; a snapshot with a lower
	// version than one already applied is stale and can be dropped.
	Version int64      `json:"version"`
	UserID  *uuid.UUID `json:"user_id,omitempty"`
	// Source is the acquisition channel: "checkout" or "quick" (a quick order
	// still needs a manager to collect address / delivery / payment).
	Source        string  `json:"source"`
	Status        string  `json:"status"`
	PaymentStatus string  `json:"payment_status"`
	PaymentMethod *string `json:"payment_method,omitempty"`

	Email    string             `json:"email"`
	Phone    *string            `json:"phone,omitempty"`
	Shipping OrderEventShipping `json:"shipping"`

	Currency      string           `json:"currency"`
	Subtotal      decimal.Decimal  `json:"subtotal"`
	DiscountTotal decimal.Decimal  `json:"discount_total"`
	ShippingTotal decimal.Decimal  `json:"shipping_total"`
	TaxTotal      decimal.Decimal  `json:"tax_total"`
	GrandTotal    decimal.Decimal  `json:"grand_total"`
	Items         []OrderEventItem `json:"items"`

	Comment   *string   `json:"comment,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type OrderPaidPayload struct {
	OrderID       uuid.UUID `json:"order_id"`
	PaymentMethod *string   `json:"payment_method,omitempty"`
	PaidAt        time.Time `json:"paid_at"`
}

type OrderCancelledPayload struct {
	OrderID     uuid.UUID `json:"order_id"`
	Reason      string    `json:"reason"`
	CancelledAt time.Time `json:"cancelled_at"`
}

// OrderStatusChangedPayload is sent by the ERP when it moves an order along
// the fulfilment flow. Only "processing", "shipped" and "delivered" are
// accepted from the ERP; the transition still goes through the order state
// machine.
type OrderStatusChangedPayload struct {
	OrderID uuid.UUID `json:"order_id"`
	// ExternalID is the ERP's own document number / GUID for this order.
	ExternalID *string              `json:"external_id,omitempty"`
	Status     constant.OrderStatus `json:"status"`
	Comment    *string              `json:"comment,omitempty"`
	ChangedAt  time.Time            `json:"changed_at"`
}

// OrderTrackingAssignedPayload is sent by the ERP once the parcel has a
// carrier tracking number. A repeated event for the same order replaces the
// previous number.
type OrderTrackingAssignedPayload struct {
	OrderID        uuid.UUID `json:"order_id"`
	TrackingNumber string    `json:"tracking_number"`
	// Provider is the carrier code (e.g. "cdek"); omitted = the carrier picked
	// at checkout.
	Provider    *string   `json:"provider,omitempty"`
	TrackingURL *string   `json:"tracking_url,omitempty"`
	AssignedAt  time.Time `json:"assigned_at"`
}
