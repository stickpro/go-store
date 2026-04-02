package viewed

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/goccy/go-json"
	"github.com/google/uuid"
	"github.com/stickpro/go-store/internal/dto"
	"github.com/stickpro/go-store/internal/storage"
	"github.com/stickpro/go-store/pkg/key_value"
	"github.com/stickpro/go-store/pkg/logger"
)

const (
	maxViewedItems   = 10
	viewedUserTTL    = 30 * 24 * time.Hour
	viewedSessionTTL = 7 * 24 * time.Hour
)

type IViewedService interface {
	Track(ctx context.Context, owner dto.Owner, variantID uuid.UUID) error
	GetViewed(ctx context.Context, owner dto.Owner) (*dto.ViewedDTO, error)
}

type Service struct {
	logger  logger.Logger
	storage storage.IStorage
	kv      key_value.IKeyValue
}

func New(l logger.Logger, st storage.IStorage, kv key_value.IKeyValue) *Service {
	return &Service{logger: l, storage: st, kv: kv}
}

// Track adds variantID to the front of the viewed list, deduplicates, and trims to maxViewedItems.
func (s Service) Track(ctx context.Context, owner dto.Owner, variantID uuid.UUID) error {
	ids, err := s.loadIDs(ctx, owner)
	if err != nil {
		return err
	}

	for i, id := range ids {
		if id == variantID {
			ids = append(ids[:i], ids[i+1:]...)
			break
		}
	}
	ids = append([]uuid.UUID{variantID}, ids...)

	if len(ids) > maxViewedItems {
		ids = ids[:maxViewedItems]
	}

	return s.saveIDs(ctx, owner, ids)
}

func (s Service) GetViewed(ctx context.Context, owner dto.Owner) (*dto.ViewedDTO, error) {
	ids, err := s.loadIDs(ctx, owner)
	if err != nil {
		return nil, err
	}
	if len(ids) == 0 {
		return &dto.ViewedDTO{}, nil
	}

	return s.enrich(ctx, ids)
}

func (s Service) loadIDs(ctx context.Context, owner dto.Owner) ([]uuid.UUID, error) {
	data, err := s.kv.Get(ctx, viewedKey(owner))
	if err != nil {
		if errors.Is(err, key_value.ErrEntryNotFound) {
			return []uuid.UUID{}, nil
		}
		return nil, err
	}

	var ids []uuid.UUID
	if err = json.Unmarshal(data.Bytes(), &ids); err != nil {
		return nil, fmt.Errorf("unmarshal viewed ids: %w", err)
	}
	return ids, nil
}

func (s Service) saveIDs(ctx context.Context, owner dto.Owner, ids []uuid.UUID) error {
	data, err := json.Marshal(ids)
	if err != nil {
		return fmt.Errorf("marshal viewed ids: %w", err)
	}
	return s.kv.Set(ctx, viewedKey(owner), string(data), viewedTTL(owner))
}

func (s Service) enrich(ctx context.Context, ids []uuid.UUID) (*dto.ViewedDTO, error) {
	rows, err := s.storage.Products().GetCartItemsByVariantIDs(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("get viewed products: %w", err)
	}

	rowByVariant := make(map[uuid.UUID]*dto.ViewedItemDTO, len(rows))
	for _, r := range rows {
		imageURL := ""
		if r.Image.Valid {
			imageURL = r.Image.String
		}
		rowByVariant[r.VariantID] = &dto.ViewedItemDTO{
			ProductID: r.ProductID,
			VariantID: r.VariantID,
			Name:      r.Name,
			Slug:      r.Slug,
			ImageURL:  imageURL,
			Price:     r.PriceRetail, // todo get price by user group retail/business/wholesale, not just retail.
		}
	}

	result := &dto.ViewedDTO{Items: make([]dto.ViewedItemDTO, 0, len(ids))}
	for _, id := range ids {
		item, ok := rowByVariant[id]
		if !ok {
			continue
		}
		result.Items = append(result.Items, *item)
	}

	return result, nil
}

func viewedKey(owner dto.Owner) string {
	if owner.UserID != nil {
		return "viewed:user:" + owner.UserID.String()
	}
	return "viewed:session:" + owner.SessionID.String()
}

func viewedTTL(owner dto.Owner) time.Duration {
	if owner.UserID != nil {
		return viewedUserTTL
	}
	return viewedSessionTTL
}
