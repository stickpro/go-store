package handler

import (
	"context"

	"github.com/stickpro/go-store/internal/dto"
	"github.com/stickpro/go-store/internal/messaging/contracts"
	"github.com/stickpro/go-store/internal/service/product"
	"github.com/stickpro/go-store/pkg/logger"
)

type ProductHandler struct {
	svc    product.IProductService
	logger logger.Logger
}

func NewProductHandler(svc product.IProductService, log logger.Logger) *ProductHandler {
	return &ProductHandler{svc: svc, logger: log}
}

func (h *ProductHandler) HandleProduct(ctx context.Context, p contracts.ProductPayload) error {
	_, err := h.svc.UpsertProductByExternalID(ctx, p.ExternalID, dto.ProductUpsertDTO{
		ExternalID:  p.ExternalID,
		Model:       p.Model,
		Sku:         p.Sku,
		Price:       p.Price,
		Quantity:    p.Quantity,
		StockStatus: p.StockStatus,
		IsEnable:    p.IsEnable,
	})
	return err
}
