package product_review_request

import "github.com/google/uuid"

type GetProductReviewsWithPagination struct {
	ProductID *uuid.UUID `json:"product_id,omitempty" query:"product_id,omitempty"`
	Page      *uint32    `json:"page" query:"page"`
	PageSize  *uint32    `json:"page_size" query:"page_size"`
} // @name GetProductReviewsWithPagination
