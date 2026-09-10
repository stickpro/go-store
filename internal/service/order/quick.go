package order

import (
	"context"

	"github.com/stickpro/go-store/internal/constant"
	"github.com/stickpro/go-store/internal/dto"
)

// CreateQuickOrder turns the caller's cart into a "quick order": the customer
// supplied only a name + phone, so the order is persisted with source "quick" in
// status "new", stock is decremented and the cart cleared exactly as in a normal
// checkout, but shipping is left unresolved and the grand total is the item
// subtotal only. A manager collects address / delivery / payment afterwards and
// transitions the order to "pending".
//
// It shares CreateOrder's transaction (idempotency, row locks, stock, history,
// order.created event) via CreateQuickOrderDTO.ToCreateOrderDTO.
func (s *Service) CreateQuickOrder(ctx context.Context, d dto.CreateQuickOrderDTO) (*dto.OrderDTO, error) {
	return s.CreateOrder(ctx, d.ToCreateOrderDTO())
}

// initialStatus is the status a freshly created order starts in: "new" for a
// quick order (awaiting a manager), "pending" for a normal checkout.
func initialStatus(d dto.CreateOrderDTO) constant.OrderStatus {
	if d.Quick {
		return constant.OrderNew
	}
	return constant.OrderPending
}

// orderSource is the acquisition channel stored on the order.
func orderSource(d dto.CreateOrderDTO) string {
	if d.Quick {
		return constant.OrderSourceQuick
	}
	return constant.OrderSourceCheckout
}
