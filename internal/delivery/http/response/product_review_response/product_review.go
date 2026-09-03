package product_review_response

import (
	"time"

	"github.com/google/uuid"
	"github.com/stickpro/go-store/internal/models"
	"github.com/stickpro/go-store/internal/storage/base"
	"github.com/stickpro/go-store/pkg/dbutils/pgtypeutils"
)

type ProductReviewResponse struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	Rating    int16     `json:"rating"`
	Title     *string   `json:"title"`
	Body      *string   `json:"body"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
} //	@name	ProductReviewResponse

func NewFromModel(productReview *models.ProductReview) *ProductReviewResponse {
	return &ProductReviewResponse{
		ID:        productReview.ID,
		UserID:    productReview.UserID,
		Rating:    productReview.Rating,
		Title:     pgtypeutils.DecodeText(productReview.Title),
		Body:      pgtypeutils.DecodeText(productReview.Body),
		Status:    productReview.Status,
		CreatedAt: productReview.CreatedAt.Time,
		UpdatedAt: productReview.UpdatedAt.Time,
	}
}

// NewPaginated maps a DB-backed page of reviews to the response contract.
func NewPaginated(
	data *base.FindResponseWithFullPagination[*models.ProductReview],
) *base.FindResponseWithFullPagination[*ProductReviewResponse] {
	items := make([]*ProductReviewResponse, 0, len(data.Items))
	for _, r := range data.Items {
		items = append(items, NewFromModel(r))
	}
	return &base.FindResponseWithFullPagination[*ProductReviewResponse]{
		Items:      items,
		Pagination: data.Pagination,
	}
}
