package yandexdelivery

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/goccy/go-json"
	"github.com/stickpro/go-store/internal/config"
	"github.com/stickpro/go-store/internal/constant"
	"github.com/stickpro/go-store/internal/dto"
	"github.com/stickpro/go-store/pkg/key_value"
	"github.com/stickpro/go-store/pkg/logger"
)

// IYandexDeliveryService exposes the Yandex Delivery pickup points list. The
// list is fetched from the Yandex Delivery API and kept in process memory
// (the full list is tens of MB, so it is not re-read from the key/value store
// per request); the key/value store only holds a copy for a warm start after a
// restart. Callers never hit the Yandex API directly.
type IYandexDeliveryService interface {
	// ListDeliveryPoints returns the in-memory delivery points, optionally
	// narrowed by filter. Returns an empty slice (not an error) if the list
	// hasn't been loaded yet or the integration is disabled.
	ListDeliveryPoints(ctx context.Context, filter dto.YandexDeliveryPointsFilter) ([]dto.YandexDeliveryPointDTO, error)
	// RefreshDeliveryPoints fetches the full pickup points list from the Yandex
	// Delivery API, swaps it into memory and persists a copy to the key/value
	// store.
	RefreshDeliveryPoints(ctx context.Context) error
	// RunCacheRefresher blocks, keeping the in-memory list warm: it seeds memory
	// from the persisted copy (or the API if there is none), then refreshes it
	// on cfg.YandexDelivery.DeliveryPointsRefreshEvery until ctx is cancelled.
	// No-op if the integration is disabled.
	RunCacheRefresher(ctx context.Context)
}

type Service struct {
	cfg    *config.Config
	logger logger.Logger
	kv     key_value.IKeyValue
	client *client

	mu     sync.RWMutex
	points []dto.YandexDeliveryPointDTO
	loaded bool
}

func New(cfg *config.Config, l logger.Logger, kv key_value.IKeyValue) *Service {
	return &Service{
		cfg:    cfg,
		logger: l,
		kv:     kv,
		client: newClient(cfg.YandexDelivery, l),
	}
}

func (s *Service) ListDeliveryPoints(_ context.Context, filter dto.YandexDeliveryPointsFilter) ([]dto.YandexDeliveryPointDTO, error) {
	if !s.cfg.YandexDelivery.Enabled {
		return []dto.YandexDeliveryPointDTO{}, nil
	}

	s.mu.RLock()
	points := s.points // the slice is only ever replaced wholesale, never mutated
	s.mu.RUnlock()

	return applyFilter(points, filter), nil
}

func (s *Service) RefreshDeliveryPoints(ctx context.Context) error {
	points, err := s.client.listAllDeliveryPoints(ctx)
	if err != nil {
		return fmt.Errorf("list yandex delivery pickup points: %w", err)
	}

	s.store(points)

	if err := s.persist(ctx, points); err != nil {
		// Memory is the source of truth; a failed persist only costs a warm start.
		s.logger.Errorw("yandex delivery: persist delivery points cache", "error", err)
	}

	s.logger.Infow("yandex delivery: delivery points refreshed", "count", len(points))
	return nil
}

func (s *Service) RunCacheRefresher(ctx context.Context) {
	if !s.cfg.YandexDelivery.Enabled {
		return
	}

	if points, err := s.loadPersisted(ctx); err != nil {
		s.logger.Errorw("yandex delivery: read persisted delivery points", "error", err)
	} else if points != nil {
		s.store(points)
		s.logger.Infow("yandex delivery: delivery points loaded from cache", "count", len(points))
	}

	if !s.isLoaded() {
		if err := s.RefreshDeliveryPoints(ctx); err != nil {
			s.logger.Errorw("yandex delivery: initial delivery points refresh failed", "error", err)
		}
	}

	interval := s.cfg.YandexDelivery.DeliveryPointsRefreshEvery
	if interval <= 0 {
		interval = 24 * time.Hour
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := s.RefreshDeliveryPoints(ctx); err != nil {
				s.logger.Errorw("yandex delivery: delivery points refresh failed", "error", err)
			}
		}
	}
}

func (s *Service) store(points []dto.YandexDeliveryPointDTO) {
	s.mu.Lock()
	s.points = points
	s.loaded = true
	s.mu.Unlock()
}

func (s *Service) isLoaded() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.loaded
}

func (s *Service) persist(ctx context.Context, points []dto.YandexDeliveryPointDTO) error {
	data, err := json.Marshal(points)
	if err != nil {
		return fmt.Errorf("marshal delivery points: %w", err)
	}
	if err := s.kv.Set(ctx, constant.CacheKeyYandexDeliveryPoints, string(data), s.cfg.YandexDelivery.DeliveryPointsCacheTTL); err != nil {
		return fmt.Errorf("cache delivery points: %w", err)
	}
	return nil
}

// loadPersisted returns nil (no error) if there is no persisted copy yet.
func (s *Service) loadPersisted(ctx context.Context) ([]dto.YandexDeliveryPointDTO, error) {
	data, err := s.kv.Get(ctx, constant.CacheKeyYandexDeliveryPoints)
	if err != nil {
		if errors.Is(err, key_value.ErrEntryNotFound) {
			return nil, nil
		}
		return nil, err
	}

	var points []dto.YandexDeliveryPointDTO
	if err := json.Unmarshal(data.Bytes(), &points); err != nil {
		return nil, fmt.Errorf("unmarshal persisted delivery points: %w", err)
	}

	return points, nil
}

func applyFilter(points []dto.YandexDeliveryPointDTO, filter dto.YandexDeliveryPointsFilter) []dto.YandexDeliveryPointDTO {
	if filter.GeoID == nil && filter.Locality == "" && filter.Type == "" && filter.BBox == nil {
		return points
	}

	result := make([]dto.YandexDeliveryPointDTO, 0, len(points))
	for _, p := range points {
		if filter.GeoID != nil && p.GeoID != *filter.GeoID {
			continue
		}
		if filter.Locality != "" && p.Locality != filter.Locality {
			continue
		}
		if filter.Type != "" && p.Type != filter.Type {
			continue
		}
		if filter.BBox != nil && !filter.BBox.Contains(p.Latitude, p.Longitude) {
			continue
		}
		result = append(result, p)
	}

	return result
}
