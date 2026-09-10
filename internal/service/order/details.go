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
	"github.com/stickpro/go-store/internal/models"
	"github.com/stickpro/go-store/internal/storage/repository"
	"github.com/stickpro/go-store/internal/storage/repository/repository_order_status_history"
	"github.com/stickpro/go-store/internal/storage/repository/repository_orders"
	"github.com/stickpro/go-store/pkg/dbutils/pgtypeutils"
)

// UpdateDetails is the admin order edit: contact, shipping address, carrier,
// payment method and comment. Item lines and their prices are never touched.
//
// Only "new" and "pending" orders are editable (ErrDetailsLocked otherwise).
// When the edit carries a delivery choice the carrier is re-quoted from the
// line-level parcel snapshot and shipping_total / grand_total are recomputed
// from the stored subtotal. Editing a "new" order (a quick order awaiting a
// manager) confirms it into "pending"; that needs a shipping address.
func (s *Service) UpdateDetails(ctx context.Context, number int64, d dto.OrderDetailsUpdateDTO) (*dto.OrderDTO, error) {
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
			return fmt.Errorf("order: lock for edit: %w", err)
		}

		from := constant.OrderStatus(o.Status)
		if from != constant.OrderNew && from != constant.OrderPending {
			return ErrDetailsLocked
		}

		items, err = s.storage.OrderItems(repository.WithTx(tx)).ListByOrderID(ctx, o.ID)
		if err != nil {
			return fmt.Errorf("order: load items for edit: %w", err)
		}

		params := detailsParamsFrom(o, d)

		shippingTotal := o.ShippingTotal
		if d.HasShippingSelection() {
			sel := d.Shipping
			if sel.Postcode == nil {
				sel.Postcode = pgtypeutils.DecodeText(params.ShipPostcode)
			}
			sc, err := s.resolveShipping(ctx, sel, checkoutLinesFromItems(items))
			if err != nil {
				return err
			}
			shippingTotal = sc.cost
			params.ShippingMethod = pgtypeutils.EncodeText(firstNonNil(sc.method, pgtypeutils.DecodeText(o.ShippingMethod)))
			params.ShipProvider = pgtypeutils.EncodeText(sc.provider)
			params.ShipTariffCode = pgtypeutils.EncodeText(sc.tariff)
			params.ShipPointCode = pgtypeutils.EncodeText(sc.point)
			params.ShipMinDays = pgtypeutils.EncodeInt4(sc.minDays)
			params.ShipMaxDays = pgtypeutils.EncodeInt4(sc.maxDays)
		}
		params.ShippingTotal = shippingTotal
		params.GrandTotal = o.Subtotal.Sub(o.DiscountTotal).Add(shippingTotal).Add(o.TaxTotal)

		to := from
		if from == constant.OrderNew {
			to = constant.OrderPending
			if params.ShipAddress == "" || params.ShipCityName == "" || params.ShipRecipient == "" {
				return ErrShippingAddressRequired
			}
		}
		params.Status = to.String()

		updated, err = s.storage.Orders(repository.WithTx(tx)).UpdateDetails(ctx, params)
		if err != nil {
			return fmt.Errorf("order: update details: %w", err)
		}

		if to != from {
			if _, err := s.storage.OrderStatusHistory(repository.WithTx(tx)).Create(ctx, repository_order_status_history.CreateParams{
				OrderID:    o.ID,
				FromStatus: pgtype.Text{String: o.Status, Valid: true},
				ToStatus:   to.String(),
				Actor:      d.Actor,
			}); err != nil {
				return fmt.Errorf("order: edit history: %w", err)
			}
		}
		return nil
	})
	if txErr != nil {
		return nil, txErr
	}

	return dto.OrderDTOFromModel(updated, items), nil
}

// detailsParamsFrom seeds the update params with the order's current values and
// overlays the non-nil fields of the edit DTO. Shipping cost / carrier / status
// are set by the caller after the re-quote.
func detailsParamsFrom(o *models.Order, d dto.OrderDetailsUpdateDTO) repository_orders.UpdateDetailsParams {
	p := repository_orders.UpdateDetailsParams{
		ID:             o.ID,
		Status:         o.Status,
		Email:          o.Email,
		Phone:          o.Phone,
		ShipCityID:     o.ShipCityID,
		ShipCityName:   o.ShipCityName,
		ShipAddress:    o.ShipAddress,
		ShipPostcode:   o.ShipPostcode,
		ShipRecipient:  o.ShipRecipient,
		ShippingMethod: o.ShippingMethod,
		ShipProvider:   o.ShipProvider,
		ShipTariffCode: o.ShipTariffCode,
		ShipPointCode:  o.ShipPointCode,
		ShipMinDays:    o.ShipMinDays,
		ShipMaxDays:    o.ShipMaxDays,
		PaymentMethod:  o.PaymentMethod,
		Comment:        o.Comment,
		ShippingTotal:  o.ShippingTotal,
		GrandTotal:     o.GrandTotal,
	}

	if d.Email != nil {
		p.Email = *d.Email
	}
	if d.Phone != nil {
		p.Phone = pgtypeutils.EncodeText(d.Phone)
	}
	if d.ShipCityID != nil {
		p.ShipCityID = uuid.NullUUID{UUID: *d.ShipCityID, Valid: true}
	}
	if d.ShipCityName != nil {
		p.ShipCityName = *d.ShipCityName
	}
	if d.ShipAddress != nil {
		p.ShipAddress = *d.ShipAddress
	}
	if d.ShipPostcode != nil {
		p.ShipPostcode = pgtypeutils.EncodeText(d.ShipPostcode)
	}
	if d.ShipRecipient != nil {
		p.ShipRecipient = *d.ShipRecipient
	}
	if d.PaymentMethod != nil {
		p.PaymentMethod = pgtypeutils.EncodeText(d.PaymentMethod)
	}
	if d.Comment != nil {
		p.Comment = pgtypeutils.EncodeText(d.Comment)
	}
	return p
}

// checkoutLinesFromItems rebuilds priced lines from persisted order items using
// the parcel snapshot stored on each line, so shipping can be re-quoted without
// touching the (possibly gone) catalogue product.
func checkoutLinesFromItems(items []*models.OrderItem) []checkoutLine {
	lines := make([]checkoutLine, 0, len(items))
	for _, it := range items {
		lines = append(lines, checkoutLine{
			productID: it.ProductID.UUID,
			variantID: it.VariantID.UUID,
			unitPrice: it.UnitPrice,
			quantity:  it.Quantity,
			weightKG:  it.WeightKg,
			lengthCM:  it.LengthCm,
			widthCM:   it.WidthCm,
			heightCM:  it.HeightCm,
		})
	}
	return lines
}
