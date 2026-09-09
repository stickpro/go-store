package service

import (
	"github.com/stickpro/go-store/internal/config"
	"github.com/stickpro/go-store/internal/service/attribute"
	"github.com/stickpro/go-store/internal/service/auth"
	"github.com/stickpro/go-store/internal/service/cart"
	"github.com/stickpro/go-store/internal/service/category"
	"github.com/stickpro/go-store/internal/service/cdek"
	"github.com/stickpro/go-store/internal/service/collections"
	"github.com/stickpro/go-store/internal/service/dashboard"
	"github.com/stickpro/go-store/internal/service/geo"
	"github.com/stickpro/go-store/internal/service/mail"
	"github.com/stickpro/go-store/internal/service/manufacturer"
	"github.com/stickpro/go-store/internal/service/media"
	"github.com/stickpro/go-store/internal/service/order"
	"github.com/stickpro/go-store/internal/service/pochta"
	"github.com/stickpro/go-store/internal/service/product"
	"github.com/stickpro/go-store/internal/service/review"
	"github.com/stickpro/go-store/internal/service/search"
	"github.com/stickpro/go-store/internal/service/search/searchtypes"
	"github.com/stickpro/go-store/internal/service/shipping"
	"github.com/stickpro/go-store/internal/service/user"
	"github.com/stickpro/go-store/internal/service/viewed"
	"github.com/stickpro/go-store/internal/service/yandexdelivery"
	"github.com/stickpro/go-store/internal/storage"
	"github.com/stickpro/go-store/pkg/logger"
	"github.com/stickpro/go-store/pkg/queue"
)

type Services struct {
	UserService          user.IUserService
	AuthService          auth.IAuthService
	CategoryService      category.ICategoryService
	ProductService       product.IProductService
	ProductReviewService review.IProductReviewService
	CollectionService    collections.ICollectionsService
	MediaService         media.IMediaService
	SearchService        searchtypes.ISearchService
	ManufacturerService  manufacturer.IManufacturerService
	AttributeService     attribute.IAttributeService
	GeoService           geo.IGeoService
	CartService          cart.ICartService
	ViewedService        viewed.IViewedService
	MailService          mail.IMailService
	OrderService         order.IOrderService
	DashboardService     dashboard.IDashboardService

	// Shipping is every cached delivery-points integration (CDEK, Yandex
	// Delivery, Russian Post). They share the /v1/delivery handler and the
	// app/tickers refresh loop.
	Shipping *shipping.Registry
}

func InitService(
	conf *config.Config,
	logger logger.Logger,
	storage storage.IStorage,
	q queue.IQueue,
) (*Services, error) {
	mailService, err := mail.New(conf, logger, q)
	if err != nil {
		return nil, err
	}

	userService := user.New(conf, logger, storage)
	authService := auth.New(conf, logger, storage, userService, mailService, storage.KeyValue())
	searchService, err := search.New(conf)
	if err != nil {
		return nil, err
	}

	categoryService := category.New(conf, logger, storage)
	productService := product.New(conf, logger, storage, searchService)
	productReviewService := review.New(conf, logger, storage, productService)
	collectionServer := collections.New(conf, logger, storage)
	mediaService := media.New(conf, logger, storage)
	manufacturerService := manufacturer.New(conf, logger, storage)
	attributeService := attribute.New(conf, logger, storage, searchService)

	geoService := geo.New(conf, logger, storage, searchService)

	cartService := cart.New(conf, logger, storage, storage.KeyValue())
	viewedService := viewed.New(conf, logger, storage, storage.KeyValue())

	cdekService := cdek.New(conf, logger, storage.KeyValue())
	yandexDeliveryService := yandexdelivery.New(conf, logger, storage.KeyValue())
	pochtaService := pochta.New(conf, logger, storage.KeyValue())

	shippingRegistry := shipping.NewRegistry(
		storage.KeyValue(),
		conf.Shipping.QuoteCacheTTL,
		shipping.ParcelDefaultsFromConfig(conf.Shipping),
		conf.Shipping.ResolvedMethods(),
		cdekService, yandexDeliveryService, pochtaService,
	)

	orderService, err := order.New(conf, logger, storage, cartService, userService, mailService, shippingRegistry, nil)
	if err != nil {
		return nil, err
	}

	dashboardService, err := dashboard.New(conf, logger, storage)
	if err != nil {
		return nil, err
	}

	return &Services{
		UserService:          userService,
		AuthService:          authService,
		CategoryService:      categoryService,
		ProductService:       productService,
		ProductReviewService: productReviewService,
		CollectionService:    collectionServer,
		MediaService:         mediaService,
		SearchService:        searchService,
		ManufacturerService:  manufacturerService,
		AttributeService:     attributeService,
		GeoService:           geoService,
		CartService:          cartService,
		ViewedService:        viewedService,
		MailService:          mailService,
		OrderService:         orderService,
		DashboardService:     dashboardService,

		Shipping: shippingRegistry,
	}, nil
}

func (s *Services) Close() error {
	s.SearchService.Close()
	if err := s.GeoService.Close(); err != nil {
		return err
	}
	return nil
}
