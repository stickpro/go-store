package server

import (
	"errors"
	"github.com/gofiber/fiber/v3"
	"github.com/stickpro/go-store/internal/config"
	"github.com/stickpro/go-store/internal/router"
	"github.com/stickpro/go-store/internal/service"
	"github.com/stickpro/go-store/internal/tools"
	"github.com/stickpro/go-store/internal/tools/apierror"
	"github.com/stickpro/go-store/pkg/logger"
)

type Server struct {
	app    *fiber.App
	cfg    config.HTTPConfig
	logger logger.Logger
}

func InitServer(cfg config.HTTPConfig, services *service.Services, logger logger.Logger) *Server {
	app := fiber.New(fiber.Config{
		ReadTimeout:     cfg.ReadTimeout,
		WriteTimeout:    cfg.WriteTimeout,
		StructValidator: tools.DefaultStructValidator(),
		ErrorHandler:    errorHandler,
	})

	router.NewRouter(cfg, services, logger).Init(app)

	return &Server{
		app:    app,
		cfg:    cfg,
		logger: logger,
	}
}

func (s *Server) Run() error {
	return s.app.Listen(":" + s.cfg.Port)
}

func (s *Server) Stop() error {
	return s.app.Shutdown()
}

func errorHandler(c fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError
	var e *fiber.Error
	if errors.As(err, &e) {
		code = e.Code
	}
	var ae *apierror.Errors
	if errors.As(err, &ae) && ae.HttpCode != 0 {
		c.Status(ae.HttpCode)
		return c.JSON(ae)
	}

	c.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

	return c.Status(code).SendString(err.Error())
}
