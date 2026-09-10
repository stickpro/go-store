package order

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
)

var (
	// ErrCartEmpty — checkout was called with nothing in the cart.
	ErrCartEmpty = errors.New("cart is empty")
	// ErrPriceChanged — CreateOrderDTO.ExpectedTotal no longer matches the
	// server-computed grand total (price or availability changed since the
	// customer reviewed the cart).
	ErrPriceChanged = errors.New("order total changed, review the cart")
	// ErrNotFound — no order matches the lookup.
	ErrNotFound = errors.New("order not found")
	// ErrForbidden — the order exists but does not belong to the requester.
	ErrForbidden = errors.New("order does not belong to the requester")
	// ErrInvalidTransition — the requested status change is not allowed from the
	// current status.
	ErrInvalidTransition = errors.New("invalid order status transition")
	// ErrDetailsLocked — the order is too far along its lifecycle to edit its
	// contact / shipping / payment details (only "new" and "pending" orders are
	// editable).
	ErrDetailsLocked = errors.New("order details can no longer be edited")
	// ErrShippingAddressRequired — a "new" order cannot be confirmed into
	// "pending" without a delivery address.
	ErrShippingAddressRequired = errors.New("shipping address is required to confirm the order")
)

// LineError points at a specific cart line that blocked checkout.
type LineError struct {
	VariantID uuid.UUID
	Reason    LineErrorReason
	Requested int64
	Available int64
}

type LineErrorReason string

const (
	LineUnavailable       LineErrorReason = "unavailable"
	LineBelowMinimum      LineErrorReason = "below_minimum"
	LineInsufficientStock LineErrorReason = "insufficient_stock"
)

func (e *LineError) Error() string {
	switch e.Reason {
	case LineBelowMinimum:
		return fmt.Sprintf("variant %s: quantity %d is below the minimum of %d", e.VariantID, e.Requested, e.Available)
	case LineInsufficientStock:
		return fmt.Sprintf("variant %s: requested %d but only %d in stock", e.VariantID, e.Requested, e.Available)
	default:
		return fmt.Sprintf("variant %s is not available for order", e.VariantID)
	}
}
