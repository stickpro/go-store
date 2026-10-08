package order

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/goccy/go-json"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"

	"github.com/stickpro/go-store/internal/constant"
	"github.com/stickpro/go-store/internal/dto"
	"github.com/stickpro/go-store/internal/models"
	"github.com/stickpro/go-store/internal/storage/repository"
	"github.com/stickpro/go-store/internal/storage/repository/repository_order_edits"
	"github.com/stickpro/go-store/internal/storage/repository/repository_order_items"
	"github.com/stickpro/go-store/internal/storage/repository/repository_orders"
	"github.com/stickpro/go-store/internal/storage/repository/repository_products"
	"github.com/stickpro/go-store/pkg/dbutils/pgtypeutils"
)

// errDryRun rolls back a preview transaction once the edited order has been
// computed: the preview runs exactly the code an apply runs, minus the commit.
var errDryRun = errors.New("order: dry run")

// itemsEditable reports whether an order's lines may still change: up to
// "processing". Once shipped, the parcel is gone; cancelled/refunded orders
// are closed.
func itemsEditable(s constant.OrderStatus) bool {
	switch s {
	case constant.OrderNew, constant.OrderPending, constant.OrderPaid, constant.OrderProcessing:
		return true
	default:
		return false
	}
}

// isPaid reports whether the customer has money on the order, which caps what
// an edit may raise the total to.
func isPaid(paymentStatus string) bool {
	return paymentStatus == constant.PaymentPaid.String() ||
		paymentStatus == constant.PaymentPartiallyRefunded.String()
}

func (s *Service) EditItems(ctx context.Context, number int64, d dto.OrderItemsEditDTO) (*dto.OrderItemsEditResultDTO, error) {
	target, err := mergeEditLines(d.Lines)
	if err != nil {
		return nil, err
	}

	var result *dto.OrderItemsEditResultDTO
	txErr := repository.BeginTxFunc(ctx, s.storage.PSQLConn(), pgx.TxOptions{}, func(tx pgx.Tx) error {
		o, err := s.storage.Orders(repository.WithTx(tx)).GetByNumberForUpdate(ctx, number)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrNotFound
			}
			return fmt.Errorf("order: lock for item edit: %w", err)
		}
		if !itemsEditable(constant.OrderStatus(o.Status)) {
			return fmt.Errorf("%w: order is %s", ErrItemsLocked, o.Status)
		}
		if d.ExpectedVersion != nil && *d.ExpectedVersion != o.Version {
			return fmt.Errorf("%w: version is %d, edit was prepared for %d", ErrEditConflict, o.Version, *d.ExpectedVersion)
		}

		items, err := s.storage.OrderItems(repository.WithTx(tx)).ListByOrderID(ctx, o.ID)
		if err != nil {
			return fmt.Errorf("order: load items for edit: %w", err)
		}

		// Catalogue rows (and product locks) are only needed for lines that
		// take more stock: added ones and ones whose quantity goes up.
		rows, err := s.catalogueRowsFor(ctx, tx, items, target)
		if err != nil {
			return err
		}

		buyer, err := s.orderBuyer(ctx, o)
		if err != nil {
			return err
		}

		plan, err := planItemsEdit(items, target, rows, buyer)
		if err != nil {
			return err
		}

		subtotal := sumSubtotal(plan.final)
		grand := subtotal.Sub(o.DiscountTotal).Add(o.ShippingTotal).Add(o.TaxTotal)
		if d.ExpectedGrandTotal != nil && !d.ExpectedGrandTotal.Equal(grand) {
			return fmt.Errorf("%w: new total is %s, edit was prepared for %s", ErrEditConflict, grand, d.ExpectedGrandTotal)
		}

		refundDue := decimal.Zero
		if isPaid(o.PaymentStatus) {
			// refunded_total only counts settled refunds. One still in flight
			// was sized against the current total; changing the total under
			// it would make that amount wrong.
			pending, err := s.storage.PaymentRefunds(repository.WithTx(tx)).SumPendingByOrderID(ctx, o.ID)
			if err != nil {
				return fmt.Errorf("order: check pending refunds: %w", err)
			}
			if pending.IsPositive() {
				return fmt.Errorf("%w: a refund of %s is still pending at the acquirer", ErrItemsLocked, pending)
			}

			netPaid := o.PaidTotal.Sub(o.RefundedTotal)
			if grand.GreaterThan(netPaid) {
				return fmt.Errorf("%w: new total %s exceeds the %s paid", ErrSurchargeRequired, grand, netPaid)
			}
			refundDue = netPaid.Sub(grand)
		}

		result = &dto.OrderItemsEditResultDTO{
			PreviousGrandTotal: o.GrandTotal,
			RefundDue:          refundDue,
			Changed:            plan.changed(),
		}

		updated := o
		if plan.changed() {
			if err := s.applyStock(ctx, tx, plan, rows); err != nil {
				return err
			}
			if err := s.writeItems(ctx, tx, o.ID, plan); err != nil {
				return err
			}

			updated, err = s.storage.Orders(repository.WithTx(tx)).UpdateTotals(ctx, repository_orders.UpdateTotalsParams{
				ID:         o.ID,
				Subtotal:   subtotal,
				GrandTotal: grand,
			})
			if err != nil {
				return fmt.Errorf("order: update totals: %w", err)
			}

			if err := s.recordEdit(ctx, tx, o, updated, items, plan, d); err != nil {
				return err
			}
		}

		finalItems, err := s.storage.OrderItems(repository.WithTx(tx)).ListByOrderID(ctx, o.ID)
		if err != nil {
			return fmt.Errorf("order: reload items after edit: %w", err)
		}
		result.Order = dto.OrderDTOFromModel(updated, finalItems)

		if d.DryRun {
			// The rolled-back UPDATE bumped the version; the preview must
			// report the version the order actually has, which is what the
			// apply call sends back as expected_version.
			result.Order.Version = o.Version
			return errDryRun
		}
		return nil
	})
	if txErr != nil && !errors.Is(txErr, errDryRun) {
		return nil, txErr
	}
	return result, nil
}

func (s *Service) ListEdits(ctx context.Context, number int64) ([]*dto.OrderEditDTO, error) {
	o, err := s.storage.Orders().GetByNumber(ctx, number)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("order: get for edit history: %w", err)
	}

	rows, err := s.storage.OrderEdits().ListByOrderID(ctx, o.ID)
	if err != nil {
		return nil, fmt.Errorf("order: list edits: %w", err)
	}

	out := make([]*dto.OrderEditDTO, 0, len(rows))
	for _, r := range rows {
		before, err := decodeEditLines(r.LinesBefore)
		if err != nil {
			return nil, err
		}
		after, err := decodeEditLines(r.LinesAfter)
		if err != nil {
			return nil, err
		}
		out = append(out, dto.OrderEditDTOFromModel(r, before, after))
	}
	return out, nil
}

// mergeEditLines validates the requested line set and folds repeated variants
// into one line, keeping first-seen order.
func mergeEditLines(lines []dto.OrderLineEditDTO) ([]dto.OrderLineEditDTO, error) {
	if len(lines) == 0 {
		return nil, ErrNoLines
	}
	idx := make(map[uuid.UUID]int, len(lines))
	out := make([]dto.OrderLineEditDTO, 0, len(lines))
	for _, l := range lines {
		if l.Quantity <= 0 {
			return nil, &LineError{VariantID: l.VariantID, Reason: LineBelowMinimum, Requested: l.Quantity, Available: 1}
		}
		if i, ok := idx[l.VariantID]; ok {
			out[i].Quantity += l.Quantity
			continue
		}
		idx[l.VariantID] = len(out)
		out = append(out, l)
	}
	return out, nil
}

// catalogueRowsFor loads (and locks) the catalogue rows of every target line
// that takes stock beyond what the order already holds.
func (s *Service) catalogueRowsFor(
	ctx context.Context,
	tx pgx.Tx,
	items []*models.OrderItem,
	target []dto.OrderLineEditDTO,
) (map[uuid.UUID]*repository_products.GetOrderLinesByVariantIDsRow, error) {
	held := make(map[uuid.UUID]int64, len(items))
	for _, it := range items {
		if it.VariantID.Valid {
			held[it.VariantID.UUID] += it.Quantity
		}
	}

	need := make([]uuid.UUID, 0, len(target))
	for _, l := range target {
		if l.Quantity > held[l.VariantID] {
			need = append(need, l.VariantID)
		}
	}

	rows := make(map[uuid.UUID]*repository_products.GetOrderLinesByVariantIDsRow, len(need))
	if len(need) == 0 {
		return rows, nil
	}

	found, err := s.storage.Products(repository.WithTx(tx)).GetOrderLinesByVariantIDs(ctx, need)
	if err != nil {
		return nil, fmt.Errorf("order: lock product lines for edit: %w", err)
	}
	for _, r := range found {
		rows[r.VariantID] = r
	}
	return rows, nil
}

// orderBuyer loads the account the order belongs to, so added lines are
// priced for that buyer exactly as checkout would. Guests (and accounts
// deleted since) price as nil.
func (s *Service) orderBuyer(ctx context.Context, o *models.Order) (*models.User, error) {
	if !o.UserID.Valid {
		return nil, nil //nolint:nilnil // a guest order has no buyer account
	}
	u, err := s.storage.Users().GetByID(ctx, o.UserID.UUID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil //nolint:nilnil // account deleted since checkout
		}
		return nil, fmt.Errorf("order: load buyer: %w", err)
	}
	return u, nil
}

// itemsEditPlan is the diff between an order's current lines and the target
// line set, plus the net stock movement per product it implies.
type itemsEditPlan struct {
	update []itemQuantityChange
	remove []*models.OrderItem
	add    []checkoutLine
	// final is the full line set after the edit, for totals.
	final []checkoutLine
	// stockDelta is units leaving stock per product: positive takes stock,
	// negative returns it.
	stockDelta map[uuid.UUID]int64
}

type itemQuantityChange struct {
	item     *models.OrderItem
	quantity int64
}

func (p *itemsEditPlan) changed() bool {
	return len(p.update) > 0 || len(p.remove) > 0 || len(p.add) > 0
}

// planItemsEdit works out what an edit does without touching the database.
// Kept lines keep their sold-at snapshot (price, name, parcel); only added
// lines read the catalogue. Lines that take more stock must be orderable now
// (enabled, at least the minimum quantity); lines that shrink or go away are
// never blocked by the catalogue.
func planItemsEdit(
	items []*models.OrderItem,
	target []dto.OrderLineEditDTO,
	rows map[uuid.UUID]*repository_products.GetOrderLinesByVariantIDsRow,
	buyer *models.User,
) (*itemsEditPlan, error) {
	current := make(map[uuid.UUID]*models.OrderItem, len(items))
	for _, it := range items {
		if !it.VariantID.Valid {
			return nil, fmt.Errorf("%w: line %q has no variant", ErrItemsLocked, it.Name)
		}
		if _, dup := current[it.VariantID.UUID]; dup {
			return nil, fmt.Errorf("%w: variant %s appears on more than one line", ErrItemsLocked, it.VariantID.UUID)
		}
		current[it.VariantID.UUID] = it
	}

	plan := &itemsEditPlan{stockDelta: make(map[uuid.UUID]int64)}
	inTarget := make(map[uuid.UUID]bool, len(target))

	for _, l := range target {
		inTarget[l.VariantID] = true
		it, kept := current[l.VariantID]

		if kept && l.Quantity <= it.Quantity {
			if l.Quantity < it.Quantity {
				plan.update = append(plan.update, itemQuantityChange{item: it, quantity: l.Quantity})
				if it.ProductID.Valid {
					plan.stockDelta[it.ProductID.UUID] -= it.Quantity - l.Quantity
				}
			}
			plan.final = append(plan.final, lineFromItem(it, l.Quantity))
			continue
		}

		// Added, or its quantity goes up: it takes stock, so it must be
		// orderable today.
		r, ok := rows[l.VariantID]
		if !ok || !r.ProductEnabled || !r.VariantEnabled {
			return nil, &LineError{VariantID: l.VariantID, Reason: LineUnavailable}
		}
		if l.Quantity < r.Minimum {
			return nil, &LineError{VariantID: l.VariantID, Reason: LineBelowMinimum, Requested: l.Quantity, Available: r.Minimum}
		}

		if kept {
			plan.update = append(plan.update, itemQuantityChange{item: it, quantity: l.Quantity})
			plan.stockDelta[r.ProductID] += l.Quantity - it.Quantity
			plan.final = append(plan.final, lineFromItem(it, l.Quantity))
			continue
		}

		line := checkoutLine{
			productID: r.ProductID,
			variantID: r.VariantID,
			sku:       pgtypeutils.DecodeText(r.Sku),
			name:      r.Name,
			slug:      strPtr(r.Slug),
			imagePath: pgtypeutils.DecodeText(r.ImagePath),
			unitPrice: resolveUnitPrice(buyer, linePricing{r.PriceRetail, r.PriceBusiness, r.PriceWholesale}),
			quantity:  l.Quantity,
			weightKG:  r.Weight,
			lengthCM:  r.Length,
			widthCM:   r.Width,
			heightCM:  r.Height,
		}
		plan.add = append(plan.add, line)
		plan.stockDelta[r.ProductID] += l.Quantity
		plan.final = append(plan.final, line)
	}

	for _, it := range items {
		if inTarget[it.VariantID.UUID] {
			continue
		}
		plan.remove = append(plan.remove, it)
		if it.ProductID.Valid {
			plan.stockDelta[it.ProductID.UUID] -= it.Quantity
		}
	}

	return plan, nil
}

// lineFromItem is a persisted line with a new quantity; everything else stays
// as sold.
func lineFromItem(it *models.OrderItem, quantity int64) checkoutLine {
	return checkoutLine{
		productID: it.ProductID.UUID,
		variantID: it.VariantID.UUID,
		sku:       pgtypeutils.DecodeText(it.Sku),
		name:      it.Name,
		slug:      pgtypeutils.DecodeText(it.Slug),
		imagePath: pgtypeutils.DecodeText(it.ImagePath),
		unitPrice: it.UnitPrice,
		quantity:  quantity,
		weightKG:  it.WeightKg,
		lengthCM:  it.LengthCm,
		widthCM:   it.WidthCm,
		heightCM:  it.HeightCm,
	}
}

// applyStock moves the net per-product stock delta. Products that take stock
// were locked by catalogueRowsFor; the guarded decrement still refuses to
// oversell. Returned stock follows RestockOrderItems' rule: only products
// that track stock.
func (s *Service) applyStock(
	ctx context.Context,
	tx pgx.Tx,
	plan *itemsEditPlan,
	rows map[uuid.UUID]*repository_products.GetOrderLinesByVariantIDsRow,
) error {
	byProduct := make(map[uuid.UUID]*repository_products.GetOrderLinesByVariantIDsRow, len(rows))
	for _, r := range rows {
		byProduct[r.ProductID] = r
	}

	// Same lock order as checkout (by product id), so concurrent edits and
	// checkouts touching the same products can't deadlock.
	productIDs := make([]uuid.UUID, 0, len(plan.stockDelta))
	for id := range plan.stockDelta {
		productIDs = append(productIDs, id)
	}
	slices.SortFunc(productIDs, func(a, b uuid.UUID) int { return bytes.Compare(a[:], b[:]) })

	for _, productID := range productIDs {
		delta := plan.stockDelta[productID]
		switch {
		case delta > 0:
			r := byProduct[productID]
			if r == nil || !r.Subtract {
				continue
			}
			n, err := s.storage.Products(repository.WithTx(tx)).DecrementProductStock(ctx, repository_products.DecrementProductStockParams{
				ID:       productID,
				Quantity: delta,
			})
			if err != nil {
				return fmt.Errorf("order: decrement stock on edit: %w", err)
			}
			if n != 1 {
				return &LineError{VariantID: r.VariantID, Reason: LineInsufficientStock, Requested: delta, Available: r.StockQuantity}
			}
		case delta < 0:
			if err := s.storage.Products(repository.WithTx(tx)).RestockTrackedProduct(ctx, repository_products.RestockTrackedProductParams{
				ID:       productID,
				Quantity: -delta,
			}); err != nil {
				return fmt.Errorf("order: restock on edit: %w", err)
			}
		}
	}
	return nil
}

func (s *Service) writeItems(ctx context.Context, tx pgx.Tx, orderID uuid.UUID, plan *itemsEditPlan) error {
	q := s.storage.OrderItems(repository.WithTx(tx))

	for _, it := range plan.remove {
		if err := q.DeleteByID(ctx, it.ID); err != nil {
			return fmt.Errorf("order: remove line: %w", err)
		}
	}
	for _, c := range plan.update {
		if _, err := q.UpdateQuantity(ctx, repository_order_items.UpdateQuantityParams{
			ID:        c.item.ID,
			Quantity:  c.quantity,
			LineTotal: c.item.UnitPrice.Mul(decimal.NewFromInt(c.quantity)),
		}); err != nil {
			return fmt.Errorf("order: change line quantity: %w", err)
		}
	}
	for _, l := range plan.add {
		if _, err := q.Create(ctx, repository_order_items.CreateParams{
			OrderID:   orderID,
			ProductID: uuid.NullUUID{UUID: l.productID, Valid: true},
			VariantID: uuid.NullUUID{UUID: l.variantID, Valid: true},
			Sku:       pgtypeutils.EncodeText(l.sku),
			Name:      l.name,
			Slug:      pgtypeutils.EncodeText(l.slug),
			ImagePath: pgtypeutils.EncodeText(l.imagePath),
			UnitPrice: l.unitPrice,
			Quantity:  l.quantity,
			LineTotal: l.lineTotal(),
			WeightKg:  l.weightKG,
			LengthCm:  l.lengthCM,
			WidthCm:   l.widthCM,
			HeightCm:  l.heightCM,
		}); err != nil {
			return fmt.Errorf("order: add line: %w", err)
		}
	}
	return nil
}

// editLineSnapshot is how a line is stored in order_edits.lines_before/after.
type editLineSnapshot struct {
	VariantID *uuid.UUID      `json:"variant_id"`
	Sku       *string         `json:"sku"`
	Name      string          `json:"name"`
	UnitPrice decimal.Decimal `json:"unit_price"`
	Quantity  int64           `json:"quantity"`
	LineTotal decimal.Decimal `json:"line_total"`
}

func (s *Service) recordEdit(
	ctx context.Context,
	tx pgx.Tx,
	before, after *models.Order,
	items []*models.OrderItem,
	plan *itemsEditPlan,
	d dto.OrderItemsEditDTO,
) error {
	beforeLines := make([]editLineSnapshot, 0, len(items))
	for _, it := range items {
		var vid *uuid.UUID
		if it.VariantID.Valid {
			vid = &it.VariantID.UUID
		}
		beforeLines = append(beforeLines, editLineSnapshot{
			VariantID: vid,
			Sku:       pgtypeutils.DecodeText(it.Sku),
			Name:      it.Name,
			UnitPrice: it.UnitPrice,
			Quantity:  it.Quantity,
			LineTotal: it.LineTotal,
		})
	}
	afterLines := make([]editLineSnapshot, 0, len(plan.final))
	for _, l := range plan.final {
		vid := l.variantID
		afterLines = append(afterLines, editLineSnapshot{
			VariantID: &vid,
			Sku:       l.sku,
			Name:      l.name,
			UnitPrice: l.unitPrice,
			Quantity:  l.quantity,
			LineTotal: l.lineTotal(),
		})
	}

	beforeJSON, err := json.Marshal(beforeLines)
	if err != nil {
		return fmt.Errorf("order: encode edit snapshot: %w", err)
	}
	afterJSON, err := json.Marshal(afterLines)
	if err != nil {
		return fmt.Errorf("order: encode edit snapshot: %w", err)
	}

	if _, err := s.storage.OrderEdits(repository.WithTx(tx)).Create(ctx, repository_order_edits.CreateParams{
		OrderID:          before.ID,
		Actor:            d.Actor,
		Comment:          pgtypeutils.EncodeText(d.Comment),
		LinesBefore:      beforeJSON,
		LinesAfter:       afterJSON,
		GrandTotalBefore: before.GrandTotal,
		GrandTotalAfter:  after.GrandTotal,
	}); err != nil {
		return fmt.Errorf("order: record edit: %w", err)
	}
	return nil
}

func decodeEditLines(raw []byte) ([]dto.OrderEditLineDTO, error) {
	var snaps []editLineSnapshot
	if err := json.Unmarshal(raw, &snaps); err != nil {
		return nil, fmt.Errorf("order: decode edit snapshot: %w", err)
	}
	out := make([]dto.OrderEditLineDTO, 0, len(snaps))
	for _, l := range snaps {
		out = append(out, dto.OrderEditLineDTO(l))
	}
	return out, nil
}
