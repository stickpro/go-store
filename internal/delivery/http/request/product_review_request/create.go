package product_review_request

import "github.com/google/uuid"

type CreateProductReviewRequest struct {
	ProductID uuid.UUID `json:"product_id"`
	Rating    int16     `json:"rating"`
	Title     string    `json:"title"`
	Body      string    `json:"body"`
}
