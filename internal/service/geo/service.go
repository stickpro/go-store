package geo

import (
	"context"
	"fmt"
	"github.com/oschwald/geoip2-golang"
	"github.com/stickpro/go-store/internal/config"
	"github.com/stickpro/go-store/internal/models"
	"github.com/stickpro/go-store/internal/storage"
	"github.com/stickpro/go-store/pkg/logger"
	"net"
)

type IGeoService interface {
	GetCityByIP(ip string) (string, error)
	Close() error
	GetAllCity(ctx context.Context) ([]*models.City, error)
}

type Service struct {
	cfg     *config.Config
	logger  logger.Logger
	db      *geoip2.Reader
	storage storage.IStorage
}

const dbPath = "storage/geo/GeoLite2-City.mmdb"

func New(cfg *config.Config, logger logger.Logger, st storage.IStorage) *Service {
	db, err := geoip2.Open(dbPath)
	if err != nil {
		logger.Error("failed to open GeoIP database", "error", err)
	}
	return &Service{
		cfg:     cfg,
		logger:  logger,
		db:      db,
		storage: st,
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
