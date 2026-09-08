package product_review_request

import "github.com/google/uuid"

// AdminListProductReviewsRequest is the admin review-moderation filter set.
// Every filter is optional; unlike the storefront listing it is not restricted
// to APPROVED reviews and can include soft-deleted rows.
type AdminListProductReviewsRequest struct {
	Page         *uint64    `json:"page" query:"page"`
	PageSize     *uint64    `json:"page_size" query:"page_size"`
	Status       *string    `json:"status" query:"status" validate:"omitempty,oneof=PENDING APPROVED REJECTED"`
	VariantID    *uuid.UUID `json:"variant_id" query:"variant_id"`
	UserID       *uuid.UUID `json:"user_id" query:"user_id"`
	WithDeleted  bool       `json:"with_deleted" query:"with_deleted"`
	SortByRating *string    `json:"sort_by_rating" query:"sort_by_rating" validate:"omitempty,oneof=asc desc"`
} //	@name	AdminListProductReviewsRequest

// UpdateProductReviewStatusRequest moderates a review: APPROVED publishes it,
// REJECTED hides it, PENDING sends it back to the queue.
type UpdateProductReviewStatusRequest struct {
	Status string `json:"status" validate:"required,oneof=PENDING APPROVED REJECTED"`
} //	@name	UpdateProductReviewStatusRequest
