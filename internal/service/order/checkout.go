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
	"github.com/stickpro/go-store/internal/service/mail"
	"github.com/stickpro/go-store/internal/storage/repository"
	"github.com/stickpro/go-store/internal/storage/repository/repository_order_items"
	"github.com/stickpro/go-store/internal/storage/repository/repository_order_status_history"
	"github.com/stickpro/go-store/internal/storage/repository/repository_orders"
	"github.com/stickpro/go-store/internal/storage/repository/repository_products"
	"github.com/stickpro/go-store/pkg/dbutils/pgerror"
	"github.com/stickpro/go-store/pkg/dbutils/pgtypeutils"
)

// errDuplicateOrder is raised inside the transaction when the idempotency key
// was claimed by a concurrent checkout; the outer call re-reads that order.
var errDuplicateOrder = errors.New("order: duplicate idempotency key")

func (s *Service) CreateOrder(ctx context.Context, d dto.CreateOrderDTO) (*dto.OrderDTO, error) {
	if d.IdempotencyKey != nil && *d.IdempotencyKey != "" {
		if existing, err := s.storage.Orders().GetByIdempotencyKey(ctx, pgtypeutils.EncodeText(d.IdempotencyKey)); err == nil {
			return s.assemble(ctx, existing)
		} else if !errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("order: idempotency lookup: %w", err)
		}
	}

	rawCart, err := s.cart.RawCart(ctx, d.Owner)
	if err != nil {
		return nil, fmt.Errorf("order: load cart: %w", err)
	}
	if len(rawCart.Items) == 0 {
		return nil, ErrCartEmpty
	}

	// Sum requested quantity per variant, keeping first-seen order for stable
	// line ordering and error reporting.
	requested := make(map[uuid.UUID]int64, len(rawCart.Items))
	variantIDs := make([]uuid.UUID, 0, len(rawCart.Items))
	for _, ci := range rawCart.Items {
		if _, seen := requested[ci.VariantID]; !seen {
			variantIDs = append(variantIDs, ci.VariantID)
		}
		requested[ci.VariantID] += int64(ci.Quantity)
	}

	var (
		created *models.Order
		items   []*models.OrderItem
	)

	txErr := repository.BeginTxFunc(ctx, s.storage.PSQLConn(), pgx.TxOptions{}, func(tx pgx.Tx) error {
		created, items, err = s.buildOrder(ctx, tx, d, variantIDs, requested)
		return err
	})
	if txErr != nil {
		if errors.Is(txErr, errDuplicateOrder) {
			o, err := s.storage.Orders().GetByIdempotencyKey(ctx, pgtypeutils.EncodeText(d.IdempotencyKey))
			if err != nil {
				return nil, fmt.Errorf("order: reload duplicate: %w", err)
			}
			return s.assemble(ctx, o)
		}
		return nil, txErr
	}

	result := dto.OrderDTOFromModel(created, items)
	s.afterCommit(ctx, d.Owner, result)
	return result, nil
}

// buildOrder does everything that must be atomic: lock product rows, validate
// and price each line, insert the order + items + first history row, decrement
// stock.
func (s *Service) buildOrder(
	ctx context.Context,
	tx pgx.Tx,
	d dto.CreateOrderDTO,
	variantIDs []uuid.UUID,
	requested map[uuid.UUID]int64,
) (*models.Order, []*models.OrderItem, error) {
	rows, err := s.storage.Products(repository.WithTx(tx)).GetOrderLinesByVariantIDs(ctx, variantIDs)
	if err != nil {
		return nil, nil, fmt.Errorf("order: lock product lines: %w", err)
	}

	byVariant := make(map[uuid.UUID]*repository_products.GetOrderLinesByVariantIDsRow, len(rows))
	perProduct := make(map[uuid.UUID]int64)
	subtractByProduct := make(map[uuid.UUID]bool)
	for _, r := range rows {
		byVariant[r.VariantID] = r
		perProduct[r.ProductID] += requested[r.VariantID]
		subtractByProduct[r.ProductID] = r.Subtract
	}

	lines := make([]checkoutLine, 0, len(variantIDs))
	for _, vid := range variantIDs {
		r, ok := byVariant[vid]
		if !ok {
			return nil, nil, &LineError{VariantID: vid, Reason: LineUnavailable}
		}
		qty := requested[vid]
		if !r.ProductEnabled || !r.VariantEnabled {
			return nil, nil, &LineError{VariantID: vid, Reason: LineUnavailable}
		}
		if qty < r.Minimum {
			return nil, nil, &LineError{VariantID: vid, Reason: LineBelowMinimum, Requested: qty, Available: r.Minimum}
		}
		if r.Subtract && perProduct[r.ProductID] > r.StockQuantity {
			return nil, nil, &LineError{
				VariantID: vid, Reason: LineInsufficientStock,
				Requested: perProduct[r.ProductID], Available: r.StockQuantity,
			}
		}

		unit := resolveUnitPrice(d.User, linePricing{r.PriceRetail, r.PriceBusiness, r.PriceWholesale})
		lines = append(lines, checkoutLine{
			productID: r.ProductID,
			variantID: r.VariantID,
			sku:       pgtypeutils.DecodeText(r.Sku),
			name:      r.Name,
			slug:      strPtr(r.Slug),
			imagePath: pgtypeutils.DecodeText(r.ImagePath),
			unitPrice: unit,
			quantity:  qty,
		})
	}

	totals := computeTotals(lines, s.shippingFor(sumSubtotal(lines)))
	if d.ExpectedTotal != nil && !d.ExpectedTotal.Equal(totals.grand) {
		return nil, nil, ErrPriceChanged
	}

	createdOrder, err := s.storage.Orders(repository.WithTx(tx)).Create(ctx, s.createParams(d, totals))
	if err != nil {
		if uc := new(pgerror.UniqueConstraintError); errors.As(pgerror.ParseError(err), &uc) {
			return nil, nil, errDuplicateOrder
		}
		return nil, nil, fmt.Errorf("order: insert: %w", err)
	}

	orderItems := make([]*models.OrderItem, 0, len(lines))
	for _, l := range lines {
		it, err := s.storage.OrderItems(repository.WithTx(tx)).Create(ctx, repository_order_items.CreateParams{
			OrderID:   createdOrder.ID,
			ProductID: uuid.NullUUID{UUID: l.productID, Valid: true},
			VariantID: uuid.NullUUID{UUID: l.variantID, Valid: true},
			Sku:       pgtypeutils.EncodeText(l.sku),
			Name:      l.name,
			Slug:      pgtypeutils.EncodeText(l.slug),
			ImagePath: pgtypeutils.EncodeText(l.imagePath),
			UnitPrice: l.unitPrice,
			Quantity:  l.quantity,
			LineTotal: l.lineTotal(),
		})
		if err != nil {
			return nil, nil, fmt.Errorf("order: insert item: %w", err)
		}
		orderItems = append(orderItems, it)
	}

	for productID, total := range perProduct {
		if !subtractByProduct[productID] {
			continue
		}
		n, err := s.storage.Products(repository.WithTx(tx)).DecrementProductStock(ctx, repository_products.DecrementProductStockParams{
			ID:       productID,
			Quantity: total,
		})
		if err != nil {
			return nil, nil, fmt.Errorf("order: decrement stock: %w", err)
		}
		if n != 1 {
			// Lost the race despite FOR UPDATE — should be unreachable, but never
			// let an order commit that oversold.
			return nil, nil, &LineError{Reason: LineInsufficientStock}
		}
	}

	if _, err := s.storage.OrderStatusHistory(repository.WithTx(tx)).Create(ctx, repository_order_status_history.CreateParams{
		OrderID:    createdOrder.ID,
		FromStatus: pgtype.Text{},
		ToStatus:   constant.OrderPending.String(),
		Actor:      constant.OrderActorCustomer,
	}); err != nil {
		return nil, nil, fmt.Errorf("order: insert history: %w", err)
	}

	return createdOrder, orderItems, nil
}

func (s *Service) createParams(d dto.CreateOrderDTO, t orderTotals) repository_orders.CreateParams {
	var userID uuid.NullUUID
	if d.User != nil {
		userID = uuid.NullUUID{UUID: d.User.ID, Valid: true}
	} else if d.Owner.UserID != nil {
		userID = uuid.NullUUID{UUID: *d.Owner.UserID, Valid: true}
	}

	var cityID uuid.NullUUID
	if d.ShipCityID != nil {
		cityID = uuid.NullUUID{UUID: *d.ShipCityID, Valid: true}
	}

	return repository_orders.CreateParams{
		UserID:         userID,
		Status:         constant.OrderPending.String(),
		PaymentStatus:  constant.PaymentUnpaid.String(),
		PaymentMethod:  pgtypeutils.EncodeText(strOrNil(d.PaymentMethod)),
		Currency:       s.currency(),
		Email:          d.Email,
		Phone:          pgtypeutils.EncodeText(d.Phone),
		ShipCityID:     cityID,
		ShipCityName:   d.ShipCityName,
		ShipAddress:    d.ShipAddress,
		ShipPostcode:   pgtypeutils.EncodeText(d.ShipPostcode),
		ShipRecipient:  d.ShipRecipient,
		ShippingMethod: pgtypeutils.EncodeText(d.ShippingMethod),
		Subtotal:       t.subtotal,
		DiscountTotal:  t.discount,
		ShippingTotal:  t.shipping,
		TaxTotal:       t.tax,
		GrandTotal:     t.grand,
		Comment:        pgtypeutils.EncodeText(d.Comment),
		IdempotencyKey: pgtypeutils.EncodeText(d.IdempotencyKey),
	}
}

// afterCommit runs the best-effort side effects. Failures are logged, never
// returned — the order is already durable.
func (s *Service) afterCommit(ctx context.Context, owner dto.Owner, o *dto.OrderDTO) {
	if err := s.cart.ClearCart(ctx, owner); err != nil {
		s.logger.Errorw("order: clear cart after checkout", "order", o.Number, "error", err)
	}

	if err := s.mail.Enqueue(ctx, o.Email, mail.OrderConfirmation{
		OrderNumber: o.Number,
		Currency:    o.Currency,
		GrandTotal:  o.GrandTotal.StringFixed(2),
	}); err != nil {
		s.logger.Errorw("order: enqueue confirmation email", "order", o.Number, "error", err)
	}

	if err := s.publisher.OrderCreated(ctx, orderCreatedEvent(o)); err != nil {
		s.logger.Errorw("order: publish order.created", "order", o.Number, "error", err)
	}
}

func orderCreatedEvent(o *dto.OrderDTO) contracts.OrderCreatedPayload {
	its := make([]contracts.OrderEventItem, 0, len(o.Items))
	for _, it := range o.Items {
		its = append(its, contracts.OrderEventItem{
			VariantID: it.VariantID,
			Sku:       it.Sku,
			Name:      it.Name,
			UnitPrice: it.UnitPrice,
			Quantity:  it.Quantity,
			LineTotal: it.LineTotal,
		})
	}

	return contracts.OrderCreatedPayload{
		OrderNumber: o.Number,
		UserID:      o.UserID,
		Email:       o.Email,
		Currency:    o.Currency,
		GrandTotal:  o.GrandTotal,
		Items:       its,
		CreatedAt:   o.CreatedAt,
	}
}
