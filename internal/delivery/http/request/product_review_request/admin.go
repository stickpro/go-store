package product_review_request

// AdminListProductReviewsRequest drives the admin moderation list. Every filter
// is optional; unlike the storefront listing it is not restricted to APPROVED
// reviews and can include soft-deleted rows.
type AdminListProductReviewsRequest struct {
	Page         *uint64 `json:"page" query:"page"`
	PageSize     *uint64 `json:"page_size" query:"page_size"`
	Status       *string `json:"status,omitempty" query:"status,omitempty" validate:"omitempty,oneof=PENDING APPROVED REJECTED"`
	WithDeleted  bool    `json:"with_deleted" query:"with_deleted"`
	SortByRating *string `json:"sort_by_rating,omitempty" query:"sort_by_rating,omitempty" validate:"omitempty,oneof=asc desc"`
} //	@name	AdminListProductReviewsRequest

// UpdateProductReviewStatusRequest carries the moderation decision for a single
// review.
type UpdateProductReviewStatusRequest struct {
	Status string `json:"status" validate:"required,oneof=PENDING APPROVED REJECTED"`
} //	@name	UpdateProductReviewStatusRequest
