// Package order owns the customer order lifecycle: checkout (cart -> persisted
// order with stock decremented in one transaction), lookup for the account that
// placed it, and status transitions with a restock on cancel/refund.
package order

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/stickpro/go-store/internal/config"
	"github.com/stickpro/go-store/internal/dto"
	"github.com/stickpro/go-store/internal/service/cart"
	"github.com/stickpro/go-store/internal/service/mail"
	"github.com/stickpro/go-store/internal/service/shipping"
	"github.com/stickpro/go-store/internal/service/user"
	"github.com/stickpro/go-store/internal/storage"
	"github.com/stickpro/go-store/internal/storage/base"
	"github.com/stickpro/go-store/pkg/logger"
)

type IOrderService interface { //nolint:interfacebloat
	// CreateOrder turns the caller's cart into a persisted order, decrements
	// stock, clears the cart, and fires the confirmation email + order.created
	// event. Idempotent when CreateOrderDTO.IdempotencyKey is set.
	CreateOrder(ctx context.Context, d dto.CreateOrderDTO) (*dto.OrderDTO, error)
	// CreateQuickOrder turns the caller's cart into a "quick order" (name + phone
	// only): persisted in status "new" / source "quick", stock decremented and
	// cart cleared as usual, but with shipping unresolved and grand total = item
	// subtotal. A manager completes it and moves it to "pending".
	CreateQuickOrder(ctx context.Context, d dto.CreateQuickOrderDTO) (*dto.OrderDTO, error)
	// PreviewCheckout returns the server-computed cart total (subtotal +
	// shipping + …) for a delivery choice, without creating an order.
	PreviewCheckout(ctx context.Context, d dto.CheckoutPreviewDTO) (*dto.CheckoutPreviewResultDTO, error)
	// GetByNumber returns an order by its human-facing number, scoped to userID
	// (guest orders are not reachable this way).
	GetByNumber(ctx context.Context, userID uuid.UUID, number int64) (*dto.OrderDTO, error)
	// GetByID returns an order by its internal id, with no ownership check.
	GetByID(ctx context.Context, id uuid.UUID) (*dto.OrderDTO, error)
	// GetForPayment resolves an order to start/inspect a payment attempt
	// against; see the implementation comment for why id (not order_number)
	// is what guest requests are authorized by.
	GetForPayment(ctx context.Context, owner dto.Owner, id uuid.UUID) (*dto.OrderDTO, error)
	// ListForUser returns the account's orders, newest first.
	ListForUser(ctx context.Context, userID uuid.UUID, d dto.GetDTO) (*base.FindResponseWithFullPagination[*dto.OrderDTO], error)
	// Cancel moves an order to cancelled and restocks it if stock had been
	// decremented. actor is recorded in the status history.
	Cancel(ctx context.Context, number int64, actor, reason string) (*dto.OrderDTO, error)
	// ListAdmin returns every order matching the filter, across all accounts
	// and guests, newest first. Admin only.
	ListAdmin(ctx context.Context, f dto.AdminOrderFilter) (*base.FindResponseWithFullPagination[*dto.OrderDTO], error)
	// GetByNumberAdmin returns an order by number without GetByNumber's
	// ownership check. Admin only.
	GetByNumberAdmin(ctx context.Context, number int64) (*dto.OrderDTO, error)
	// UpdateStatus drives an admin-initiated status transition. Cancelling is
	// delegated to Cancel.
	UpdateStatus(ctx context.Context, number int64, d dto.OrderStatusUpdateDTO) (*dto.OrderDTO, error)
	// UpdateDetails is the admin order edit: contact, shipping address, carrier,
	// payment method, comment. Item lines are never touched. Only "new" and
	// "pending" orders are editable; editing a "new" order confirms it into
	// "pending". Admin only.
	UpdateDetails(ctx context.Context, number int64, d dto.OrderDetailsUpdateDTO) (*dto.OrderDTO, error)
	// RecordRefund brings the order in line with money returned on its
	// payment; refundedTotal is the running total. A full refund moves the
	// order to "refunded" (with restock) when its status allows it, otherwise
	// — and for any partial refund — only payment_status / refunded_total
	// change. Repeating it for an order already marked refunded is a no-op.
	RecordRefund(ctx context.Context, orderID uuid.UUID, refundedTotal decimal.Decimal, full bool, actor string) (*dto.OrderDTO, error)
	// EditItems replaces the order's lines with the given target set (admin
	// only), moving stock and recomputing totals in one transaction. Allowed
	// up to "processing"; a paid order's total may only go down, with the
	// difference reported as RefundDue. DryRun computes the same result
	// without saving it.
	EditItems(ctx context.Context, number int64, d dto.OrderItemsEditDTO) (*dto.OrderItemsEditResultDTO, error)
	// ListEdits returns the order's item-edit audit trail, newest first.
	ListEdits(ctx context.Context, number int64) ([]*dto.OrderEditDTO, error)
}

type Service struct {
	cfg       *config.Config
	logger    logger.Logger
	storage   storage.IStorage
	cart      cart.ICartService
	users     user.IUserService
	mail      mail.IMailService
	publisher EventPublisher
	shipping  *shipping.Registry

	flatShipping     decimal.Decimal
	freeShippingFrom decimal.Decimal
}

func New(
	cfg *config.Config,
	l logger.Logger,
	st storage.IStorage,
	cartService cart.ICartService,
	userService user.IUserService,
	mailService mail.IMailService,
	shippingRegistry *shipping.Registry,
	publisher EventPublisher,
) (*Service, error) {
	flat, err := decimal.NewFromString(orDefault(cfg.Order.FlatShipping, "0"))
	if err != nil {
		return nil, fmt.Errorf("order: parse flat_shipping %q: %w", cfg.Order.FlatShipping, err)
	}
	free, err := decimal.NewFromString(orDefault(cfg.Order.FreeShippingFrom, "0"))
	if err != nil {
		return nil, fmt.Errorf("order: parse free_shipping_from %q: %w", cfg.Order.FreeShippingFrom, err)
	}

	if publisher == nil {
		publisher = NoopPublisher{}
	}

	return &Service{
		cfg:              cfg,
		logger:           l,
		storage:          st,
		cart:             cartService,
		users:            userService,
		mail:             mailService,
		publisher:        publisher,
		shipping:         shippingRegistry,
		flatShipping:     flat,
		freeShippingFrom: free,
	}, nil
}

func (s *Service) currency() string {
	if s.cfg.Order.Currency != "" {
		return s.cfg.Order.Currency
	}
	return "RUB"
}

func orDefault(v, def string) string {
	if v == "" {
		return def
	}
	return v
}
