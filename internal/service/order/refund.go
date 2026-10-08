package order

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"

	"github.com/stickpro/go-store/internal/constant"
	"github.com/stickpro/go-store/internal/dto"
	"github.com/stickpro/go-store/internal/storage/repository/repository_orders"
)

func (s *Service) RecordRefund(ctx context.Context, orderID uuid.UUID, refundedTotal decimal.Decimal, full bool, actor string) (*dto.OrderDTO, error) {
	o, err := s.GetByID(ctx, orderID)
	if err != nil {
		return nil, err
	}
	if o.PaymentStatus == constant.PaymentRefunded.String() {
		return o, nil
	}

	if full && canTransition(constant.OrderStatus(o.Status), constant.OrderRefunded) {
		return s.UpdateStatus(ctx, o.Number, dto.OrderStatusUpdateDTO{
			Status: constant.OrderRefunded.String(),
			Actor:  actor,
		})
	}

	paymentStatus := constant.PaymentPartiallyRefunded
	if full {
		paymentStatus = constant.PaymentRefunded
	}

	updated, err := s.storage.Orders().SetRefundState(ctx, repository_orders.SetRefundStateParams{
		ID:            orderID,
		PaymentStatus: paymentStatus.String(),
		RefundedTotal: refundedTotal,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%w: order %d has payment status %q", ErrInvalidTransition, o.Number, o.PaymentStatus)
		}
		return nil, fmt.Errorf("order: record refund: %w", err)
	}
	return s.assemble(ctx, updated)
}
