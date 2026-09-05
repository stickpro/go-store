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
	"github.com/stickpro/go-store/internal/storage/repository/repository_orders"
	"github.com/stickpro/go-store/pkg/dbutils/pgtypeutils"
)

// UpdateStatus drives an admin-initiated order status transition: pending ->
// paid -> processing -> shipped -> delivered, plus refund. Cancelling is
// delegated to Cancel so cancelled_at/restock/the order.cancelled event stay
// defined in one place.
func (s *Service) UpdateStatus(ctx context.Context, number int64, d dto.OrderStatusUpdateDTO) (*dto.OrderDTO, error) {
	to := constant.OrderStatus(d.Status)
	if to == constant.OrderCancelled {
		reason := ""
		if d.Comment != nil {
			reason = *d.Comment
		}
		return s.Cancel(ctx, number, d.Actor, reason)
	}

	var (
		updated *models.Order
		items   []*models.OrderItem
		from    constant.OrderStatus
	)

	txErr := repository.BeginTxFunc(ctx, s.storage.PSQLConn(), pgx.TxOptions{}, func(tx pgx.Tx) error {
		o, err := s.storage.Orders(repository.WithTx(tx)).GetByNumberForUpdate(ctx, number)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrNotFound
			}
			return fmt.Errorf("order: lock for status update: %w", err)
		}
		from = constant.OrderStatus(o.Status)

		if !canTransition(from, to) {
			return fmt.Errorf("%w: %s -> %s", ErrInvalidTransition, from, to)
		}

		if restocksOn(to) {
			if err := s.storage.Products(repository.WithTx(tx)).RestockOrderItems(ctx, o.ID); err != nil {
				return fmt.Errorf("order: restock on %s: %w", to, err)
			}
		}

		switch to {
		case constant.OrderPaid:
			paymentMethod := o.PaymentMethod
			if d.PaymentMethod != nil {
				paymentMethod = pgtypeutils.EncodeText(d.PaymentMethod)
			}
			updated, err = s.storage.Orders(repository.WithTx(tx)).MarkPaid(ctx, repository_orders.MarkPaidParams{
				ID:            o.ID,
				Status:        to.String(),
				PaymentMethod: paymentMethod,
			})
		case constant.OrderRefunded:
			updated, err = s.storage.Orders(repository.WithTx(tx)).MarkRefunded(ctx, o.ID)
		default:
			updated, err = s.storage.Orders(repository.WithTx(tx)).UpdateStatus(ctx, repository_orders.UpdateStatusParams{
				ID:     o.ID,
				Status: to.String(),
			})
		}
		if err != nil {
			return fmt.Errorf("order: update status: %w", err)
		}

		if _, err := s.storage.OrderStatusHistory(repository.WithTx(tx)).Create(ctx, repository_order_status_history.CreateParams{
			OrderID:    o.ID,
			FromStatus: pgtype.Text{String: o.Status, Valid: true},
			ToStatus:   to.String(),
			Actor:      d.Actor,
			Comment:    pgtypeutils.EncodeText(d.Comment),
		}); err != nil {
			return fmt.Errorf("order: status history: %w", err)
		}

		items, err = s.storage.OrderItems(repository.WithTx(tx)).ListByOrderID(ctx, o.ID)
		return err
	})
	if txErr != nil {
		return nil, txErr
	}

	result := dto.OrderDTOFromModel(updated, items)

	if to == constant.OrderPaid {
		if err := s.publisher.OrderPaid(ctx, contracts.OrderPaidPayload{
			OrderNumber:   result.Number,
			PaymentMethod: result.PaymentMethod,
			PaidAt:        derefOrNow(result.PaidAt),
		}); err != nil {
			s.logger.Errorw("order: publish order.paid", "order", result.Number, "error", err)
		}
	}

	return result, nil
}
