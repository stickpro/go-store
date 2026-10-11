package order

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/stickpro/go-store/internal/constant"
	"github.com/stickpro/go-store/internal/dto"
	"github.com/stickpro/go-store/internal/messaging/contracts"
	"github.com/stickpro/go-store/internal/models"
	"github.com/stickpro/go-store/internal/storage/repository"
	"github.com/stickpro/go-store/internal/storage/repository/repository_order_inbound_events"
	"github.com/stickpro/go-store/internal/storage/repository/repository_order_status_history"
	"github.com/stickpro/go-store/internal/storage/repository/repository_orders"
	"github.com/stickpro/go-store/pkg/dbutils/pgtypeutils"
)

// erpAccepts reports whether the ERP may drive an order into this status.
// Payment, cancellation and refunds stay with the store.
func erpAccepts(to constant.OrderStatus) bool {
	switch to {
	case constant.OrderProcessing, constant.OrderShipped, constant.OrderDelivered:
		return true
	default:
		return false
	}
}

// ApplyERPStatus applies a fulfilment status reported by the ERP. A repeated
// event and a status the order is already in are no-ops. The ERP's document
// number is stored even when the transition itself is rejected, so the two
// systems stay linked.
func (s *Service) ApplyERPStatus(ctx context.Context, d dto.OrderERPStatusDTO) error {
	to := constant.OrderStatus(d.Status)
	if !erpAccepts(to) {
		return fmt.Errorf("%w: status %q is not accepted from the ERP", ErrInvalidTransition, d.Status)
	}

	var rejected error

	txErr := s.inERPEvent(ctx, d.EventID, contracts.OrderEventStatusChanged, d.OrderID, func(tx pgx.Tx, o *models.Order) error {
		if d.ExternalID != nil && pgtypeutils.DecodeText(o.ExternalID) == nil {
			if _, err := s.storage.Orders(repository.WithTx(tx)).SetExternalID(ctx, repository_orders.SetExternalIDParams{
				ID:         o.ID,
				ExternalID: pgtypeutils.EncodeText(d.ExternalID),
			}); err != nil {
				return fmt.Errorf("order: set external id: %w", err)
			}
		}

		from := constant.OrderStatus(o.Status)
		if from == to {
			return nil
		}
		if !canTransition(from, to) {
			rejected = fmt.Errorf("%w: %s -> %s", ErrInvalidTransition, from, to)
			return nil
		}

		if _, err := s.storage.Orders(repository.WithTx(tx)).UpdateStatus(ctx, repository_orders.UpdateStatusParams{
			ID:     o.ID,
			Status: to.String(),
		}); err != nil {
			return fmt.Errorf("order: update status: %w", err)
		}

		if _, err := s.storage.OrderStatusHistory(repository.WithTx(tx)).Create(ctx, repository_order_status_history.CreateParams{
			OrderID:    o.ID,
			FromStatus: pgtype.Text{String: o.Status, Valid: true},
			ToStatus:   to.String(),
			Actor:      constant.OrderActorERP,
			Comment:    pgtypeutils.EncodeText(d.Comment),
		}); err != nil {
			return fmt.Errorf("order: status history: %w", err)
		}
		return nil
	})
	if txErr != nil {
		return txErr
	}
	return rejected
}

// AssignERPTracking stores the carrier tracking number reported by the ERP,
// replacing a previous one. It never changes the order status. A repeated
// event is a no-op.
func (s *Service) AssignERPTracking(ctx context.Context, d dto.OrderERPTrackingDTO) error {
	return s.inERPEvent(ctx, d.EventID, contracts.OrderEventTrackingAssigned, d.OrderID, func(tx pgx.Tx, o *models.Order) error {
		provider := o.ShipProvider
		if d.Provider != nil {
			provider = pgtypeutils.EncodeText(d.Provider)
		}

		if _, err := s.storage.Orders(repository.WithTx(tx)).SetTracking(ctx, repository_orders.SetTrackingParams{
			ID:               o.ID,
			TrackingNumber:   pgtype.Text{String: d.TrackingNumber, Valid: true},
			TrackingProvider: provider,
			TrackingUrl:      pgtypeutils.EncodeText(d.TrackingURL),
		}); err != nil {
			return fmt.Errorf("order: set tracking: %w", err)
		}
		return nil
	})
}

// inERPEvent runs apply against the row-locked order inside one transaction
// that also registers the inbound event id, so an event is applied at most
// once: a redelivered one is skipped without calling apply.
func (s *Service) inERPEvent(
	ctx context.Context,
	eventID uuid.UUID,
	eventType string,
	orderID uuid.UUID,
	apply func(tx pgx.Tx, o *models.Order) error,
) error {
	return repository.BeginTxFunc(ctx, s.storage.PSQLConn(), pgx.TxOptions{}, func(tx pgx.Tx) error {
		o, err := s.storage.Orders(repository.WithTx(tx)).GetByIDForUpdate(ctx, orderID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrNotFound
			}
			return fmt.Errorf("order: lock for erp event: %w", err)
		}

		registered, err := s.storage.OrderInboundEvents(repository.WithTx(tx)).Register(ctx, repository_order_inbound_events.RegisterParams{
			EventID:   eventID,
			EventType: eventType,
			OrderID:   orderID,
		})
		if err != nil {
			return fmt.Errorf("order: register erp event: %w", err)
		}
		if registered == 0 {
			return nil
		}

		return apply(tx, o)
	})
}
