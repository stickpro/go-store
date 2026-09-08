package product_review_response

import (
	"time"

	"github.com/google/uuid"

	"github.com/stickpro/go-store/internal/models"
	"github.com/stickpro/go-store/internal/storage/base"
	"github.com/stickpro/go-store/pkg/dbutils/pgtypeutils"
)

// AdminProductReviewResponse is ProductReviewResponse plus the fields the
// moderation panel needs: which variant the review is on, the linked order (if
// any) and the soft-delete timestamp.
type AdminProductReviewResponse struct {
	ProductReviewResponse
	VariantID uuid.UUID  `json:"variant_id"`
	OrderID   *uuid.UUID `json:"order_id"`
	DeletedAt *time.Time `json:"deleted_at"`
} //	@name	AdminProductReviewResponse

func NewAdminFromModel(r *models.ProductReview) *AdminProductReviewResponse {
	var orderID *uuid.UUID
	if r.OrderID.Valid {
		orderID = &r.OrderID.UUID
	}
	return &AdminProductReviewResponse{
		ProductReviewResponse: *NewFromModel(r),
		VariantID:             r.VariantID,
		OrderID:               orderID,
		DeletedAt:             pgtypeutils.DecodeTimePtr(r.DeletedAt),
	}
}

// NewAdminPaginated maps a page of reviews to the admin response contract.
func NewAdminPaginated(
	data *base.FindResponseWithFullPagination[*models.ProductReview],
) *base.FindResponseWithFullPagination[*AdminProductReviewResponse] {
	items := make([]*AdminProductReviewResponse, 0, len(data.Items))
	for _, r := range data.Items {
		items = append(items, NewAdminFromModel(r))
	}
	return &base.FindResponseWithFullPagination[*AdminProductReviewResponse]{
		Items:      items,
		Pagination: data.Pagination,
	}
}
