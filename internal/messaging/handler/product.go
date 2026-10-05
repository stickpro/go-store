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
		Name:           p.Name,
		Model:          p.Model,
		Sku:            p.Sku,
		PriceRetail:    p.PriceRetail,
		PriceBusiness:  p.PriceBusiness,
		PriceWholesale: p.PriceWholesale,
		Quantity:       p.Quantity,
		StockStatus:    p.StockStatus,
		IsEnable:       p.IsEnable,
		Weight:         p.Weight.Decimal,
		Length:         p.Length.Decimal,
		Width:          p.Width.Decimal,
		Height:         p.Height.Decimal,
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

	// Best effort: the Postgres write already committed, so a search-side
	// hiccup here must not fail the handler — that would redeliver the same
	// Kafka message forever instead of just leaving the index briefly stale.
	if err := h.svc.IndexProductVariants(ctx, productID); err != nil {
		h.logger.Errorw("kafka: reindex product variants", "product_id", productID, "error", err)
	}
	// A brand-new attribute slug (attrSvc.RunInTx above) defaults to
	// filterable; push it to the index's settings so category filters can
	// actually use it as a facet right away, without a full reindex.
	if err := h.attrSvc.RefreshFilterableAttributes(ctx); err != nil {
		h.logger.Errorw("kafka: refresh filterable attributes", "product_id", productID, "error", err)
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
