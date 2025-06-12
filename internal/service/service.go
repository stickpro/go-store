package service

import (
	"github.com/stickpro/go-store/internal/config"
	"github.com/stickpro/go-store/internal/service/attribute"
	"github.com/stickpro/go-store/internal/service/auth"
	"github.com/stickpro/go-store/internal/service/category"
	"github.com/stickpro/go-store/internal/service/geo"
	"github.com/stickpro/go-store/internal/service/manufacturer"
	"github.com/stickpro/go-store/internal/service/media"
	"github.com/stickpro/go-store/internal/service/product"
	"github.com/stickpro/go-store/internal/service/search"
	"github.com/stickpro/go-store/internal/service/search/searchtypes"
	"github.com/stickpro/go-store/internal/service/user"
	"github.com/stickpro/go-store/internal/storage"
	"github.com/stickpro/go-store/pkg/logger"
)

type Services struct {
	UserService         user.IUserService
	AuthService         auth.IAuthService
	CategoryService     category.ICategoryService
	ProductService      product.IProductService
	MediaService        media.IMediaService
	SearchService       searchtypes.ISearchService
	ManufacturerService manufacturer.IManufacturerService
	AttributeService    attribute.IAttributeService
	GeoService          geo.IGeoService
}

func InitService(
	conf *config.Config,
	logger logger.Logger,
	storage storage.IStorage,
) (*Services, error) {
	userService := user.New(conf, logger, storage)
	authService := auth.New(conf, logger, storage, userService)

	categoryService := category.New(conf, logger, storage)
	productService := product.New(conf, logger, storage)
	mediaService := media.New(conf, logger, storage)
	manufacturerService := manufacturer.New(conf, logger, storage)
	attributeService := attribute.New(conf, logger, storage)

	searchService, _ := search.New(conf)

	geoService := geo.New(conf, logger, storage, searchService)

	return &Services{
		UserService:         userService,
		AuthService:         authService,
		CategoryService:     categoryService,
		ProductService:      productService,
		MediaService:        mediaService,
		SearchService:       searchService,
		ManufacturerService: manufacturerService,
		AttributeService:    attributeService,
		GeoService:          geoService,
	}, nil
}

func (s *Services) Close() error {
	if err := s.GeoService.Close(); err != nil {
		return err
	}
	return nil
}
