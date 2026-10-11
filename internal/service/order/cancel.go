package order

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/stickpro/go-store/internal/constant"
	"github.com/stickpro/go-store/internal/dto"
	"github.com/stickpro/go-store/internal/messaging/contracts"
	"github.com/stickpro/go-store/internal/models"
	"github.com/stickpro/go-store/internal/storage/repository"
	"github.com/stickpro/go-store/internal/storage/repository/repository_order_status_history"
	"github.com/stickpro/go-store/pkg/dbutils/pgtypeutils"
)

// Cancel moves an order to "cancelled" and, if it had decremented stock, returns
// every line to the catalogue, and queues the order.cancelled event — all in
// one transaction. Used by the customer,
// the admin panel, and the pending-order expiry worker.
func (s *Service) Cancel(ctx context.Context, number int64, actor, reason string) (*dto.OrderDTO, error) {
	var (
		updated *models.Order
		items   []*models.OrderItem
	)

	txErr := repository.BeginTxFunc(ctx, s.storage.PSQLConn(), pgx.TxOptions{}, func(tx pgx.Tx) error {
		o, err := s.storage.Orders(repository.WithTx(tx)).GetByNumberForUpdate(ctx, number)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrNotFound
			}
			return fmt.Errorf("order: lock for cancel: %w", err)
		}

		if !canTransition(constant.OrderStatus(o.Status), constant.OrderCancelled) {
			return fmt.Errorf("%w: %s -> %s", ErrInvalidTransition, o.Status, constant.OrderCancelled)
		}

		// Restock before flipping status so a failure here rolls the whole thing back.
		if err := s.storage.Products(repository.WithTx(tx)).RestockOrderItems(ctx, o.ID); err != nil {
			return fmt.Errorf("order: restock on cancel: %w", err)
		}

		updated, err = s.storage.Orders(repository.WithTx(tx)).MarkCancelled(ctx, o.ID)
		if err != nil {
			return fmt.Errorf("order: mark cancelled: %w", err)
		}

		if _, err := s.storage.OrderStatusHistory(repository.WithTx(tx)).Create(ctx, repository_order_status_history.CreateParams{
			OrderID:    o.ID,
			FromStatus: pgtype.Text{String: o.Status, Valid: true},
			ToStatus:   constant.OrderCancelled.String(),
			Actor:      actor,
			Comment:    pgtypeutils.EncodeText(strPtr(reason)),
		}); err != nil {
			return fmt.Errorf("order: cancel history: %w", err)
		}

		if err := s.enqueueEvent(ctx, tx, contracts.OrderEventCancelled, updated.ID, contracts.OrderCancelledPayload{
			OrderID:     updated.ID,
			Reason:      reason,
			CancelledAt: updated.CancelledAt.Time,
		}); err != nil {
			return err
		}

		items, err = s.storage.OrderItems(repository.WithTx(tx)).ListByOrderID(ctx, o.ID)
		if err != nil {
			return fmt.Errorf("order: load items after cancel: %w", err)
		}
		return nil
	})
	if txErr != nil {
		return nil, txErr
	}

	return dto.OrderDTOFromModel(updated, items), nil
}
