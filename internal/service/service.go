package service

import (
	"github.com/stickpro/go-store/internal/config"
	"github.com/stickpro/go-store/internal/service/auth"
	"github.com/stickpro/go-store/internal/service/category"
	"github.com/stickpro/go-store/internal/service/media"
	"github.com/stickpro/go-store/internal/service/product"
	"github.com/stickpro/go-store/internal/service/user"
	"github.com/stickpro/go-store/internal/storage"
	"github.com/stickpro/go-store/pkg/logger"
)

type Services struct {
	UserService     user.IUser
	AuthService     auth.IAuth
	CategoryService category.ICategory
	ProductService  product.IProduct
	MediaService    media.IMediaService
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
	return &Services{
		UserService:     userService,
		AuthService:     authService,
		CategoryService: categoryService,
		ProductService:  productService,
		MediaService:    mediaService,
	}, nil
}
