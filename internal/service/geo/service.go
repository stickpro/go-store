package geo

import (
	"context"
	"fmt"
	"github.com/oschwald/geoip2-golang"
	"github.com/stickpro/go-store/internal/config"
	"github.com/stickpro/go-store/internal/constant"
	"github.com/stickpro/go-store/internal/models"
	"github.com/stickpro/go-store/internal/service/search/searchtypes"
	"github.com/stickpro/go-store/internal/storage"
	"github.com/stickpro/go-store/pkg/logger"
	utils "github.com/stickpro/go-store/pkg/util"
	"net"
)

type IGeoService interface {
	GetCityByIP(ip string) (string, error)
	Close() error
	GetAllCity(ctx context.Context) ([]*models.City, error)
	GetPopularCity(ctx context.Context) ([]*models.City, error)
	CreateCityIndex(ctx context.Context, reindex bool) error
}

type Service struct {
	cfg           *config.Config
	logger        logger.Logger
	db            *geoip2.Reader
	storage       storage.IStorage
	searchService searchtypes.ISearchService
}

const dbPath = "storage/geo/GeoLite2-City.mmdb"

func New(cfg *config.Config, logger logger.Logger, st storage.IStorage, ss searchtypes.ISearchService) *Service {
	db, err := geoip2.Open(dbPath)
	if err != nil {
		logger.Error("failed to open GeoIP database", "error", err)
	}
	return &Service{
		cfg:           cfg,
		logger:        logger,
		db:            db,
		storage:       st,
		searchService: ss,
	}
}

func (s *Service) GetCityByIP(ip string) (string, error) {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		s.logger.Error("Invalid IP address", "ip", ip)
		return "", fmt.Errorf("invalid IP address: %s", ip)
	}

	record, err := s.db.City(parsedIP)
	if err != nil {
		s.logger.Error("Failed to get city from GeoIP2", "ip", ip, "error", err)
		return "", fmt.Errorf("failed to get city: %w", err)
	}

	cityName := record.City.Names["ru"]
	if cityName == "" {
		cityName = record.City.Names["en"]
	}
	if cityName == "" {
		s.logger.Warn("City name not found for IP", "ip", ip)
		return "", fmt.Errorf("city name not found for IP: %s", ip)
	}

	return cityName, nil
}

func (s *Service) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

// GetAllCity Method only for create index not use in project
func (s *Service) GetAllCity(ctx context.Context) ([]*models.City, error) {
	cities, err := s.storage.Cities().GetAll(ctx)
	if err != nil {
		return nil, err
	}
	return cities, nil
}

func (s *Service) GetPopularCity(ctx context.Context) ([]*models.City, error) {
	cities, err := s.storage.Cities().GetCityOrderByPopulation(ctx)
	if err != nil {
		s.logger.Error("Failed to get popular cities", "error", err)
		return nil, fmt.Errorf("failed to get popular cities: %w", err)
	}
	return cities, nil
}

func (s *Service) CreateCityIndex(ctx context.Context, reindex bool) error {
	shouldCreate := reindex

	if !reindex {
		exists, err := s.searchService.CheckIndex(constant.CitiesIndex)
		if err != nil {
			s.logger.Error("Failed to check index", "error", err)
			return err
		}
		shouldCreate = !exists
	}

	if !shouldCreate {
		return nil
	}

	cities, err := s.GetAllCity(ctx)
	if err != nil {
		s.logger.Error("Failed to get all cities", "error", err)
		return err
	}

	data, err := utils.StructToMap(cities)
	if err != nil {
		s.logger.Error("Failed to convert cities to map", "error", err)
		return err
	}

	opts := searchtypes.IndexOptions{
		SearchableAttributes: []string{"city", "address", "region"},
	}

	err = s.searchService.CreateIndex(constant.CitiesIndex, data, opts)
	if err != nil {
		s.logger.Error("Failed to create city index", "error", err)
		return err
	}
	s.logger.Debug("Create new index ", constant.CitiesIndex)

	return nil
}
