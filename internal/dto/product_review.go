package dto

import (
	"github.com/google/uuid"
	"github.com/stickpro/go-store/internal/constant"
	"github.com/stickpro/go-store/internal/delivery/http/request/product_review_request"
)

type GetProductReviewsDTO struct {
	Page         *uint32 `json:"page" query:"page"`
	PageSize     *uint32 `json:"page_size" query:"page_size"`
	WithDeleted  bool    `json:"with_deleted" query:"with_deleted"`
	SortByRating *string `json:"sort_by_rating,omitempty" query:"sort_by_rating,omitempty"`
}

type CreateProductReviewDTO struct {
	ProductID uuid.UUID `json:"product_id"`
	UserID    uuid.UUID `json:"user_id"`
	OrderID   uuid.UUID `json:"order_id"`
	Rating    int16     `json:"rating"`
	Title     string    `json:"title"`
	Body      string    `json:"body"`
}

func RequestToCreateProductReviewDTO(req *product_review_request.CreateProductReviewRequest, usrID uuid.UUID) CreateProductReviewDTO {
	return CreateProductReviewDTO{
		ProductID: req.ProductID,
		UserID:    usrID,
		Rating:    req.Rating,
		Title:     req.Title,
		Body:      req.Body,
	}
}

func RequestToGetProductReviewDTO(req *product_review_request.GetProductReviewsWithPagination) GetProductReviewsDTO {
	return GetProductReviewsDTO{
		Page:         req.Page,
		PageSize:     req.PageSize,
		WithDeleted:  false,
		SortByRating: req.SortByRating,
	}
}

type UpdateProductReviewStatusDTO struct {
	ID     uuid.UUID                    `json:"id"`
	Status constant.ProductReviewStatus `json:"status"`
}
