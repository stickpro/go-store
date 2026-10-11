package order

import (
	"context"
	"fmt"

	"github.com/goccy/go-json"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/stickpro/go-store/internal/dto"
	"github.com/stickpro/go-store/internal/messaging/contracts"
	"github.com/stickpro/go-store/internal/models"
	"github.com/stickpro/go-store/internal/storage/repository"
	"github.com/stickpro/go-store/internal/storage/repository/repository_order_outbox"
)

// enqueueEvent records an order event for the ERP in the outbox, inside the
// transaction that makes the change — the event exists if and only if the
// change committed. The relay worker publishes it to Kafka afterwards. A
// failure here fails the surrounding transaction. No-op while the order
// exchange is disabled.
func (s *Service) enqueueEvent(ctx context.Context, tx pgx.Tx, eventType string, orderID uuid.UUID, payload any) error {
	if !s.cfg.Kafka.OrderOutbox.Enabled {
		return nil
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("order: marshal %s event: %w", eventType, err)
	}

	if err := s.storage.OrderOutbox(repository.WithTx(tx)).Enqueue(ctx, repository_order_outbox.EnqueueParams{
		EventID:       uuid.New(),
		EventType:     eventType,
		SchemaVersion: contracts.OrderSchemaVersion,
		OrderID:       orderID,
		Payload:       data,
	}); err != nil {
		return fmt.Errorf("order: enqueue %s event: %w", eventType, err)
	}
	return nil
}

// enqueueSnapshot queues an event carrying the whole order (order.created /
// order.updated). order and items must be the rows as written by the
// surrounding transaction.
func (s *Service) enqueueSnapshot(ctx context.Context, tx pgx.Tx, eventType string, o *models.Order, items []*models.OrderItem) error {
	return s.enqueueEvent(ctx, tx, eventType, o.ID, orderSnapshot(dto.OrderDTOFromModel(o, items)))
}

// orderSnapshot maps an order to the full-snapshot body of the order.created /
// order.updated events.
func orderSnapshot(o *dto.OrderDTO) contracts.OrderSnapshotPayload {
	its := make([]contracts.OrderEventItem, 0, len(o.Items))
	for _, it := range o.Items {
		its = append(its, contracts.OrderEventItem{
			ProductID:  it.ProductID,
			VariantID:  it.VariantID,
			ExternalID: it.ExternalID,
			Sku:        it.Sku,
			Name:       it.Name,
			UnitPrice:  it.UnitPrice,
			Quantity:   it.Quantity,
			LineTotal:  it.LineTotal,
		})
	}

	return contracts.OrderSnapshotPayload{
		OrderID:       o.ID,
		OrderNumber:   o.Number,
		Version:       o.Version,
		UserID:        o.UserID,
		Source:        o.Source,
		Status:        o.Status,
		PaymentStatus: o.PaymentStatus,
		PaymentMethod: o.PaymentMethod,
		Email:         o.Email,
		Phone:         o.Phone,
		Shipping: contracts.OrderEventShipping{
			CityName:   o.Shipping.CityName,
			Address:    o.Shipping.Address,
			Postcode:   o.Shipping.Postcode,
			Recipient:  o.Shipping.Recipient,
			Method:     o.Shipping.Method,
			Provider:   o.Shipping.Provider,
			TariffCode: o.Shipping.TariffCode,
			PointCode:  o.Shipping.PointCode,
		},
		Currency:      o.Currency,
		Subtotal:      o.Subtotal,
		DiscountTotal: o.DiscountTotal,
		ShippingTotal: o.ShippingTotal,
		TaxTotal:      o.TaxTotal,
		GrandTotal:    o.GrandTotal,
		Items:         its,
		Comment:       o.Comment,
		CreatedAt:     o.CreatedAt,
	}
}
