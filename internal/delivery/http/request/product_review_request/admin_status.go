package product_review_request

// UpdateProductReviewStatusRequest moderates a review: APPROVED publishes it,
// REJECTED hides it, PENDING sends it back to the queue.
type UpdateProductReviewStatusRequest struct {
	Status string `json:"status" validate:"required,oneof=PENDING APPROVED REJECTED"`
} //	@name	UpdateProductReviewStatusRequest
