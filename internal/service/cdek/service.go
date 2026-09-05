package cdek

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/goccy/go-json"
	"github.com/stickpro/go-store/internal/config"
	"github.com/stickpro/go-store/internal/constant"
	"github.com/stickpro/go-store/internal/dto"
	"github.com/stickpro/go-store/pkg/key_value"
	"github.com/stickpro/go-store/pkg/logger"
)

// ICDEKService exposes the cached CDEK delivery points list. The list is
// fetched from the CDEK API and kept in the key/value store; callers never hit
// the CDEK API directly.
type ICDEKService interface {
	// ListDeliveryPoints returns the cached delivery points, optionally narrowed
	// by filter. Returns an empty slice (not an error) if the cache hasn't been
	// populated yet or the integration is disabled.
	ListDeliveryPoints(ctx context.Context, filter dto.CDEKDeliveryPointsFilter) ([]dto.CDEKDeliveryPointDTO, error)
	// RefreshDeliveryPoints fetches the full delivery points list from the CDEK
	// API and overwrites the cache.
	RefreshDeliveryPoints(ctx context.Context) error
	// RunCacheRefresher blocks, keeping the delivery points cache warm: it
	// populates the cache immediately if empty, then refreshes it on
	// cfg.CDEK.DeliveryPointsRefreshEvery until ctx is cancelled. No-op if the
	// integration is disabled.
	RunCacheRefresher(ctx context.Context)
}

type Service struct {
	cfg    *config.Config
	logger logger.Logger
	kv     key_value.IKeyValue
	client *client
}

func New(cfg *config.Config, l logger.Logger, kv key_value.IKeyValue) *Service {
	return &Service{
		cfg:    cfg,
		logger: l,
		kv:     kv,
		client: newClient(cfg.CDEK, l),
	}
}

func (s *Service) ListDeliveryPoints(ctx context.Context, filter dto.CDEKDeliveryPointsFilter) ([]dto.CDEKDeliveryPointDTO, error) {
	if !s.cfg.CDEK.Enabled {
		return []dto.CDEKDeliveryPointDTO{}, nil
	}

	points, err := s.loadCache(ctx)
	if err != nil {
		return nil, fmt.Errorf("load cached delivery points: %w", err)
	}

	return applyFilter(points, filter), nil
}

func (s *Service) RefreshDeliveryPoints(ctx context.Context) error {
	points, err := s.client.searchAllDeliveryPoints(ctx, s.cfg.CDEK.DeliveryPointsCountryCode)
	if err != nil {
		return fmt.Errorf("search cdek delivery points: %w", err)
	}

	data, err := json.Marshal(points)
	if err != nil {
		return fmt.Errorf("marshal delivery points: %w", err)
	}

	if err := s.kv.Set(ctx, constant.CacheKeyCDEKDeliveryPoints, string(data), s.cfg.CDEK.DeliveryPointsCacheTTL); err != nil {
		return fmt.Errorf("cache delivery points: %w", err)
	}

	s.logger.Infow("cdek: delivery points cache refreshed", "count", len(points))
	return nil
}

func (s *Service) RunCacheRefresher(ctx context.Context) {
	if !s.cfg.CDEK.Enabled {
		return
	}

	if cached, err := s.loadCache(ctx); err != nil {
		s.logger.Errorw("cdek: read delivery points cache", "error", err)
	} else if cached == nil {
		if err := s.RefreshDeliveryPoints(ctx); err != nil {
			s.logger.Errorw("cdek: initial delivery points refresh failed", "error", err)
		}
	}

	interval := s.cfg.CDEK.DeliveryPointsRefreshEvery
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
				s.logger.Errorw("cdek: delivery points refresh failed", "error", err)
			}
		}
	}
}

// loadCache returns nil (no error) if the cache hasn't been populated yet.
func (s *Service) loadCache(ctx context.Context) ([]dto.CDEKDeliveryPointDTO, error) {
	data, err := s.kv.Get(ctx, constant.CacheKeyCDEKDeliveryPoints)
	if err != nil {
		if errors.Is(err, key_value.ErrEntryNotFound) {
			return nil, nil
		}
		return nil, err
	}

	var points []dto.CDEKDeliveryPointDTO
	if err := json.Unmarshal(data.Bytes(), &points); err != nil {
		return nil, fmt.Errorf("unmarshal cached delivery points: %w", err)
	}

	return points, nil
}

func applyFilter(points []dto.CDEKDeliveryPointDTO, filter dto.CDEKDeliveryPointsFilter) []dto.CDEKDeliveryPointDTO {
	if filter.CityCode == nil && filter.PostalCode == "" && filter.Type == "" && filter.BBox == nil {
		return points
	}

	result := make([]dto.CDEKDeliveryPointDTO, 0, len(points))
	for _, p := range points {
		if filter.CityCode != nil && p.CityCode != *filter.CityCode {
			continue
		}
		if filter.PostalCode != "" && p.PostalCode != filter.PostalCode {
			continue
		}
		if filter.Type != "" && p.Type != filter.Type {
			continue
		}
		if filter.BBox != nil && !inBBox(p, *filter.BBox) {
			continue
		}
		result = append(result, p)
	}

	return result
}

// inBBox reports whether p falls inside b, bounds inclusive.
func inBBox(p dto.CDEKDeliveryPointDTO, b dto.CDEKDeliveryPointsBBox) bool {
	return p.Latitude >= b.MinLat && p.Latitude <= b.MaxLat &&
		p.Longitude >= b.MinLon && p.Longitude <= b.MaxLon
}
