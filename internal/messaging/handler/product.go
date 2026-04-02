package handler

import (
	"context"

	"github.com/goccy/go-json"
	"github.com/google/uuid"
	"github.com/stickpro/go-store/internal/dto"
	"github.com/stickpro/go-store/internal/messaging/contracts"
	"github.com/stickpro/go-store/internal/messaging/tasks"
	"github.com/stickpro/go-store/internal/service/attribute"
	"github.com/stickpro/go-store/internal/service/product"
	"github.com/stickpro/go-store/internal/storage/repository"
	"github.com/stickpro/go-store/pkg/logger"
	"github.com/stickpro/go-store/pkg/queue"
)

type ProductHandler struct {
	svc     product.IProductService
	attrSvc attribute.IAttributeService
	queue   queue.IQueue
	logger  logger.Logger
}

func NewProductHandler(svc product.IProductService, attrSvc attribute.IAttributeService, q queue.IQueue, log logger.Logger) *ProductHandler {
	return &ProductHandler{svc: svc, attrSvc: attrSvc, queue: q, logger: log}
}

func (h *ProductHandler) HandleProduct(ctx context.Context, p contracts.ProductPayload) error {
	upsertDTO := dto.ProductUpsertDTO{
		ExternalID:     p.ExternalID,
		Model:          p.Model,
		Sku:            p.Sku,
		PriceRetail:    p.PriceRetail,
		PriceBusiness:  p.PriceBusiness,
		PriceWholesale: p.PriceWholesale,
		Quantity:       p.Quantity,
		StockStatus:    p.StockStatus,
		IsEnable:       p.IsEnable,
	}

	items := make([]dto.AttributeKafkaItem, len(p.Attributes))
	for i, a := range p.Attributes {
		items[i] = dto.AttributeKafkaItem{
			Name:  a.Name,
			Slug:  a.Slug,
			Type:  a.Type,
			Unit:  a.Unit,
			Value: a.Value,
		}
	}

	var productID uuid.UUID
	err := h.attrSvc.RunInTx(ctx, func(opts ...repository.Option) error {
		prod, txErr := h.svc.UpsertProductByExternalID(ctx, p.ExternalID, upsertDTO, opts...)
		if txErr != nil {
			return txErr
		}
		productID = prod.ID
		return h.attrSvc.SyncAttributesFromKafka(ctx, productID, items, opts...)
	})
	if err != nil {
		return err
	}

	if len(p.Images) == 0 && p.ImageMain == nil {
		return nil
	}

	task := tasks.ImageSyncTask{
		ProductID: productID,
		ImageMain: p.ImageMain,
		Images:    p.Images,
	}
	payload, err := json.Marshal(task)
	if err != nil {
		return err
	}
	return h.queue.Push(ctx, tasks.ImageSyncQueue, payload)
}
