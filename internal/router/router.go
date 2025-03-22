package router

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/etag"
	"github.com/stickpro/go-store/internal/config"
	"github.com/stickpro/go-store/internal/delivery/http/handlers"
	"github.com/stickpro/go-store/internal/delivery/http/handlers/admin"
	"github.com/stickpro/go-store/internal/delivery/middleware"
	"github.com/stickpro/go-store/internal/service"
	"github.com/stickpro/go-store/pkg/logger"
)

type Router struct {
	config   *config.Config
	services *service.Services
	logger   logger.Logger
}

func NewRouter(conf *config.Config, services *service.Services, logger logger.Logger) *Router {
	return &Router{
		config:   conf,
		services: services,
		logger:   logger,
	}
}

func (r *Router) Init(app *fiber.App) {
	app.Use(etag.New())

	if r.config.HTTP.Cors.Enabled {
		corsConfig := cors.ConfigDefault
		corsConfig.AllowMethods = []string{
			fiber.MethodGet,
			fiber.MethodPost,
			fiber.MethodHead,
			fiber.MethodPut,
			fiber.MethodDelete,
			fiber.MethodPatch,
			fiber.MethodOptions,
		}

		if len(r.config.HTTP.Cors.AllowedOrigins) > 0 {
			corsConfig.AllowOrigins = r.config.HTTP.Cors.AllowedOrigins
		}

		app.Use(cors.New(corsConfig))
	}

	app.Use(middleware.CacheControlMiddleware())

	app.Get("/ping", func(c fiber.Ctx) error {
		return c.SendString("pong")
	})

	r.initAPI(app)
	r.initStaticFiles(app)
}

func (r *Router) initAPI(app *fiber.App) {
	handlerV1 := handlers.NewHandler(r.services, r.logger)
	handlerV1.InitHandler(app)

	adminHandler := admin.NewAdminHandler(r.services)
	adminHandler.InitAdminHandler(app)
}
