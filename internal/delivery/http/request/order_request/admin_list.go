package order_request

import "github.com/google/uuid"

// AdminListOrdersRequest is the admin order list filter set. Every filter is
// optional; an empty request lists every order, newest first.
type AdminListOrdersRequest struct {
	Page          *uint64    `json:"page" query:"page"`
	PageSize      *uint64    `json:"page_size" query:"page_size"`
	Status        *string    `json:"status" query:"status" validate:"omitempty,oneof=pending paid processing shipped delivered cancelled refunded"`
	PaymentStatus *string    `json:"payment_status" query:"payment_status" validate:"omitempty,oneof=unpaid paid refunded failed"`
	UserID        *uuid.UUID `json:"user_id" query:"user_id"`
	// CreatedFrom, CreatedTo bound created_at, inclusive. RFC3339, e.g.
	// "2025-01-01T00:00:00Z".
	CreatedFrom *string `json:"created_from" query:"created_from"`
	CreatedTo   *string `json:"created_to" query:"created_to"`
} //	@name	AdminListOrdersRequest
