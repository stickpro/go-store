package order_response

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/stickpro/go-store/internal/dto"
	"github.com/stickpro/go-store/internal/storage/base"
)

// AdminOrderResponse is OrderResponse plus the fields only the admin panel
// needs: the owning account (absent for guest orders), what was paid, and the
// version to send back with an item edit.
type AdminOrderResponse struct {
	OrderResponse
	UserID    *uuid.UUID      `json:"user_id"`
	PaidTotal decimal.Decimal `json:"paid_total"`
	Version   int64           `json:"version"`
} //	@name	AdminOrderResponse

func NewAdminFromDTO(d *dto.OrderDTO) *AdminOrderResponse {
	return &AdminOrderResponse{
		OrderResponse: *NewFromDTO(d),
		UserID:        d.UserID,
		PaidTotal:     d.PaidTotal,
		Version:       d.Version,
	}
}

// NewAdminPaginated maps a page of order DTOs to the admin response contract.
func NewAdminPaginated(
	data *base.FindResponseWithFullPagination[*dto.OrderDTO],
) *base.FindResponseWithFullPagination[*AdminOrderResponse] {
	items := make([]*AdminOrderResponse, 0, len(data.Items))
	for _, o := range data.Items {
		items = append(items, NewAdminFromDTO(o))
	}
	return &base.FindResponseWithFullPagination[*AdminOrderResponse]{
		Items:      items,
		Pagination: data.Pagination,
	}
}
