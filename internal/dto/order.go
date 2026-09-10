package dto

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shopspring/decimal"

	"github.com/stickpro/go-store/internal/delivery/http/request/order_request"
	"github.com/stickpro/go-store/internal/models"
	"github.com/stickpro/go-store/pkg/dbutils/pgtypeutils"
)

// CreateOrderDTO is the checkout input. The cart itself is not passed here —
// the order service re-reads it (raw) and re-prices every line inside the
// creating transaction.
type CreateOrderDTO struct {
	Owner Owner
	// User is the authenticated account, or nil for a guest checkout.
	User *models.User

	Email string
	Phone *string

	ShipCityID     *uuid.UUID
	ShipCityName   string
	ShipAddress    string
	ShipPostcode   *string
	ShipRecipient  string
	ShippingMethod *string

	// Delivery choice. DeliveryMethodCode (from GET /v1/delivery/methods) is
	// resolved to a carrier + tariff by the order service; ShipProvider +
	// ShipTariffCode are the raw fallback. With neither, flat/free shipping.
	DeliveryMethodCode *string
	ShipProvider       *string
	ShipTariffCode     *string
	ShipPointCode      *string

	PaymentMethod string
	Comment       *string

	// IdempotencyKey, when set, makes repeated checkout calls return the first
	// created order instead of creating duplicates.
	IdempotencyKey *string
	// ExpectedTotal, when set, must equal the server-computed grand total or the
	// order is rejected with ErrPriceChanged.
	ExpectedTotal *decimal.Decimal

	// Quick marks a one-click "quick order": shipping is not resolved, the order
	// starts in status "new" with source "quick" and a grand total of the item
	// subtotal only. Set via CreateQuickOrderDTO, never from the checkout request.
	Quick bool
}

// CreateQuickOrderDTO is the "quick order" input: the customer has a cart and
// leaves only a name + phone (optionally an email / comment). Address, delivery
// and payment are collected later by a manager, so the resulting order starts in
// status "new" with source "quick".
type CreateQuickOrderDTO struct {
	Owner Owner
	// User is the authenticated account, or nil for a guest.
	User *models.User

	Name    string
	Phone   string
	Email   *string
	Comment *string

	// IdempotencyKey, when set, makes repeated calls return the first created
	// order instead of creating duplicates.
	IdempotencyKey *string
}

// ToCreateOrderDTO folds the quick-order input into the shared checkout DTO with
// Quick set. The account email wins over any email in the request; a guest with
// no email produces an order with an empty email (phone is the contact then).
func (d CreateQuickOrderDTO) ToCreateOrderDTO() CreateOrderDTO {
	email := ""
	switch {
	case d.User != nil:
		email = d.User.Email
	case d.Email != nil:
		email = *d.Email
	}
	phone := d.Phone

	return CreateOrderDTO{
		Owner:          d.Owner,
		User:           d.User,
		Email:          email,
		Phone:          &phone,
		ShipRecipient:  d.Name,
		Comment:        d.Comment,
		IdempotencyKey: d.IdempotencyKey,
		Quick:          true,
	}
}

// RequestToCreateQuickOrderDTO builds the quick-order input from the HTTP request
// plus the resolved caller. user is nil for a guest.
func RequestToCreateQuickOrderDTO(
	req *order_request.CreateQuickOrderRequest,
	owner Owner,
	user *models.User,
) CreateQuickOrderDTO {
	return CreateQuickOrderDTO{
		Owner:   owner,
		User:    user,
		Name:    req.Name,
		Phone:   req.Phone,
		Email:   req.Email,
		Comment: req.Comment,
	}
}

// ErrEmailRequired is returned by RequestToCreateOrderDTO when a guest checkout
// omits the contact email.
var ErrEmailRequired = errors.New("email is required for guest checkout")

// RequestToCreateOrderDTO builds the checkout input from the HTTP request plus
// the resolved caller. user is nil for guest checkout.
func RequestToCreateOrderDTO(req *order_request.CreateOrderRequest, owner Owner, user *models.User) (CreateOrderDTO, error) {
	email := req.Email
	if user != nil {
		email = user.Email
	}
	if email == "" {
		return CreateOrderDTO{}, ErrEmailRequired
	}

	var expected *decimal.Decimal
	if req.ExpectedTotal != nil && *req.ExpectedTotal != "" {
		v, err := decimal.NewFromString(*req.ExpectedTotal)
		if err != nil {
			return CreateOrderDTO{}, errors.New("expected_total must be a decimal number")
		}
		expected = &v
	}

	return CreateOrderDTO{
		Owner:              owner,
		User:               user,
		Email:              email,
		Phone:              req.Phone,
		ShipCityID:         req.ShipCityID,
		ShipCityName:       req.ShipCityName,
		ShipAddress:        req.ShipAddress,
		ShipPostcode:       req.ShipPostcode,
		ShipRecipient:      req.ShipRecipient,
		ShippingMethod:     req.ShippingMethod,
		DeliveryMethodCode: req.DeliveryMethodCode,
		ShipProvider:       req.ShipProvider,
		ShipTariffCode:     req.ShipTariffCode,
		ShipPointCode:      req.ShipPointCode,
		PaymentMethod:      req.PaymentMethod,
		Comment:            req.Comment,
		ExpectedTotal:      expected,
	}, nil
}

// ShippingSelection groups the delivery-choice fields of the checkout DTO.
func (d CreateOrderDTO) ShippingSelection() ShippingSelection {
	return ShippingSelection{
		MethodCode: d.DeliveryMethodCode,
		Provider:   d.ShipProvider,
		TariffCode: d.ShipTariffCode,
		PointCode:  d.ShipPointCode,
		Postcode:   d.ShipPostcode,
		Method:     d.ShippingMethod,
	}
}

// CheckoutPreviewDTO asks the order service for the server-computed cart total
// under a delivery choice, without creating an order. The cart is the caller's.
type CheckoutPreviewDTO struct {
	Owner    Owner
	User     *models.User
	Shipping ShippingSelection
}

// CheckoutPreviewResultDTO is the authoritative money breakdown for a cart +
// delivery choice. The frontend renders GrandTotal directly instead of adding
// shipping to the cart total itself.
type CheckoutPreviewResultDTO struct {
	Currency      string
	ItemCount     int
	Subtotal      decimal.Decimal
	DiscountTotal decimal.Decimal
	ShippingTotal decimal.Decimal
	TaxTotal      decimal.Decimal
	GrandTotal    decimal.Decimal
	Shipping      OrderShippingDTO
}

// RequestToCheckoutPreviewDTO builds the preview input from the HTTP request.
func RequestToCheckoutPreviewDTO(req *order_request.CheckoutPreviewRequest, owner Owner, user *models.User) CheckoutPreviewDTO {
	return CheckoutPreviewDTO{
		Owner: owner,
		User:  user,
		Shipping: ShippingSelection{
			MethodCode: req.DeliveryMethodCode,
			Provider:   req.ShipProvider,
			TariffCode: req.ShipTariffCode,
			PointCode:  req.ShipPointCode,
			Postcode:   req.ShipPostcode,
		},
	}
}

// RequestToListOrdersDTO maps the paging query into the shared GetDTO.
func RequestToListOrdersDTO(req *order_request.ListOrdersRequest) GetDTO {
	return GetDTO{Page: req.Page, PageSize: req.PageSize}
}

// AdminOrderFilter is the optional filter set for the admin order list.
// A nil field means "don't filter on it".
type AdminOrderFilter struct {
	Page          *uint64
	PageSize      *uint64
	Status        *string
	PaymentStatus *string
	Source        *string
	UserID        *uuid.UUID
	CreatedFrom   *time.Time
	CreatedTo     *time.Time
}

// RequestToAdminOrderFilter maps the admin list query into AdminOrderFilter.
// Fails if created_from/created_to aren't valid RFC3339 timestamps.
func RequestToAdminOrderFilter(req *order_request.AdminListOrdersRequest) (AdminOrderFilter, error) {
	from, err := parseRFC3339Ptr(req.CreatedFrom)
	if err != nil {
		return AdminOrderFilter{}, fmt.Errorf("created_from: %w", err)
	}
	to, err := parseRFC3339Ptr(req.CreatedTo)
	if err != nil {
		return AdminOrderFilter{}, fmt.Errorf("created_to: %w", err)
	}

	return AdminOrderFilter{
		Page:          req.Page,
		PageSize:      req.PageSize,
		Status:        req.Status,
		PaymentStatus: req.PaymentStatus,
		Source:        req.Source,
		UserID:        req.UserID,
		CreatedFrom:   from,
		CreatedTo:     to,
	}, nil
}

func parseRFC3339Ptr(v *string) (*time.Time, error) {
	if v == nil || *v == "" {
		return nil, nil
	}
	t, err := time.Parse(time.RFC3339, *v)
	if err != nil {
		return nil, errors.New("must be an RFC3339 timestamp")
	}
	return &t, nil
}

// OrderStatusUpdateDTO drives an admin-initiated order status transition.
// Cancelling (Status == constant.OrderCancelled) is handled by the Cancel
// method instead, so cancelled_at/restock/notification stay in one place.
type OrderStatusUpdateDTO struct {
	Status  string
	Actor   string
	Comment *string
	// PaymentMethod is only applied when Status is "paid".
	PaymentMethod *string
}

// OrderDetailsUpdateDTO is the admin order-edit input. Every field is optional —
// a nil pointer leaves that column as-is. Item lines and their prices are never
// touched. When any Shipping selection field is set the carrier is re-quoted and
// shipping_total / grand_total recomputed from the stored subtotal; otherwise the
// persisted shipping stands. Editing a "new" order (quick order) transitions it
// to "pending"; that requires a shipping address.
type OrderDetailsUpdateDTO struct {
	Actor string

	Email         *string
	Phone         *string
	ShipCityID    *uuid.UUID
	ShipCityName  *string
	ShipAddress   *string
	ShipPostcode  *string
	ShipRecipient *string
	PaymentMethod *string
	Comment       *string

	Shipping ShippingSelection
}

// HasShippingSelection reports whether the edit carries a delivery choice that
// should trigger a re-quote.
func (d OrderDetailsUpdateDTO) HasShippingSelection() bool {
	s := d.Shipping
	return nonEmpty(s.MethodCode) || nonEmpty(s.Provider) || nonEmpty(s.TariffCode) || nonEmpty(s.PointCode)
}

func nonEmpty(s *string) bool { return s != nil && *s != "" }

// RequestToOrderDetailsUpdateDTO maps the admin edit request into the DTO.
func RequestToOrderDetailsUpdateDTO(req *order_request.AdminUpdateOrderRequest, actor string) OrderDetailsUpdateDTO {
	return OrderDetailsUpdateDTO{
		Actor:         actor,
		Email:         req.Email,
		Phone:         req.Phone,
		ShipCityID:    req.ShipCityID,
		ShipCityName:  req.ShipCityName,
		ShipAddress:   req.ShipAddress,
		ShipPostcode:  req.ShipPostcode,
		ShipRecipient: req.ShipRecipient,
		PaymentMethod: req.PaymentMethod,
		Comment:       req.Comment,
		Shipping: ShippingSelection{
			MethodCode: req.DeliveryMethodCode,
			Provider:   req.ShipProvider,
			TariffCode: req.ShipTariffCode,
			PointCode:  req.ShipPointCode,
			Postcode:   req.ShipPostcode,
		},
	}
}

// ShippingSelection is the customer's delivery choice, shared by the checkout
// preview and the checkout itself.
//
// MethodCode (from GET /v1/delivery/methods) is the preferred input — the order
// service resolves it to a carrier + tariff. Provider + TariffCode are the raw
// fallback. With neither, the configured flat/free shipping applies.
type ShippingSelection struct {
	MethodCode *string
	Provider   *string
	TariffCode *string
	PointCode  *string
	Postcode   *string
	Method     *string
}

type OrderShippingDTO struct {
	CityID    *uuid.UUID
	CityName  string
	Address   string
	Postcode  *string
	Recipient string
	Method    *string

	// Carrier snapshot: what was picked and quoted at checkout.
	Provider   *string
	TariffCode *string
	PointCode  *string
	MinDays    *int32
	MaxDays    *int32
}

type OrderItemDTO struct {
	ProductID *uuid.UUID
	VariantID *uuid.UUID
	Sku       *string
	Name      string
	Slug      *string
	ImagePath *string
	UnitPrice decimal.Decimal
	Quantity  int64
	LineTotal decimal.Decimal
}

type OrderDTO struct {
	ID            uuid.UUID
	Number        int64
	UserID        *uuid.UUID
	Status        string
	Source        string
	PaymentStatus string
	PaymentMethod *string
	Currency      string

	Email    string
	Phone    *string
	Shipping OrderShippingDTO

	Items         []OrderItemDTO
	Subtotal      decimal.Decimal
	DiscountTotal decimal.Decimal
	ShippingTotal decimal.Decimal
	TaxTotal      decimal.Decimal
	GrandTotal    decimal.Decimal

	Comment     *string
	CreatedAt   time.Time
	PaidAt      *time.Time
	CancelledAt *time.Time
}

// OrderDTOFromModel maps a persisted order plus its line rows into the domain DTO.
func OrderDTOFromModel(o *models.Order, items []*models.OrderItem) *OrderDTO {
	out := &OrderDTO{
		ID:            o.ID,
		Number:        o.OrderNumber,
		UserID:        nullUUIDPtr(o.UserID),
		Status:        o.Status,
		Source:        o.Source,
		PaymentStatus: o.PaymentStatus,
		PaymentMethod: pgtypeutils.DecodeText(o.PaymentMethod),
		Currency:      o.Currency,
		Email:         o.Email,
		Phone:         pgtypeutils.DecodeText(o.Phone),
		Shipping: OrderShippingDTO{
			CityID:     nullUUIDPtr(o.ShipCityID),
			CityName:   o.ShipCityName,
			Address:    o.ShipAddress,
			Postcode:   pgtypeutils.DecodeText(o.ShipPostcode),
			Recipient:  o.ShipRecipient,
			Method:     pgtypeutils.DecodeText(o.ShippingMethod),
			Provider:   pgtypeutils.DecodeText(o.ShipProvider),
			TariffCode: pgtypeutils.DecodeText(o.ShipTariffCode),
			PointCode:  pgtypeutils.DecodeText(o.ShipPointCode),
			MinDays:    pgtypeutils.DecodeInt4(o.ShipMinDays),
			MaxDays:    pgtypeutils.DecodeInt4(o.ShipMaxDays),
		},
		Subtotal:      o.Subtotal,
		DiscountTotal: o.DiscountTotal,
		ShippingTotal: o.ShippingTotal,
		TaxTotal:      o.TaxTotal,
		GrandTotal:    o.GrandTotal,
		Comment:       pgtypeutils.DecodeText(o.Comment),
		CreatedAt:     o.CreatedAt.Time,
		PaidAt:        timestampPtr(o.PaidAt),
		CancelledAt:   timestampPtr(o.CancelledAt),
	}

	out.Items = make([]OrderItemDTO, 0, len(items))
	for _, it := range items {
		out.Items = append(out.Items, OrderItemDTO{
			ProductID: nullUUIDPtr(it.ProductID),
			VariantID: nullUUIDPtr(it.VariantID),
			Sku:       pgtypeutils.DecodeText(it.Sku),
			Name:      it.Name,
			Slug:      pgtypeutils.DecodeText(it.Slug),
			ImagePath: pgtypeutils.DecodeText(it.ImagePath),
			UnitPrice: it.UnitPrice,
			Quantity:  it.Quantity,
			LineTotal: it.LineTotal,
		})
	}
	return out
}

func nullUUIDPtr(v uuid.NullUUID) *uuid.UUID {
	if !v.Valid {
		return nil
	}
	id := v.UUID
	return &id
}

func timestampPtr(v pgtype.Timestamp) *time.Time {
	if !v.Valid {
		return nil
	}
	t := v.Time
	return &t
}
