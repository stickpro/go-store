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
	cfg    *config.Config
	logger logger.Logger
}

func InitServer(cfg *config.Config, services *service.Services, logger logger.Logger) *Server {
	app := fiber.New(fiber.Config{
		ReadTimeout:     cfg.HTTP.ReadTimeout,
		WriteTimeout:    cfg.HTTP.WriteTimeout,
		StructValidator: tools.DefaultStructValidator(),
		ErrorHandler:    errorHandler,
		BodyLimit:       cfg.HTTP.MaxBodyLimit * 1024 * 1024,
	})

	router.NewRouter(cfg, services, logger).Init(app)
	return &Server{
		app:    app,
		cfg:    cfg,
		logger: logger,
	}
}

func (s *Server) Run() error {
	return s.app.Listen(":" + s.cfg.HTTP.Port)
}

func (s *Server) Stop() error {
	return s.app.Shutdown()
}

func errorHandler(c fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError

	var fiberErr *fiber.Error
	if errors.As(err, &fiberErr) {
		code = fiberErr.Code
		return c.Status(code).JSON(apierror.New(apierror.Error{
			Message: fiberErr.Message,
		}).SetHttpCode(code))
	}

	var apiErr *apierror.Errors
	if errors.As(err, &apiErr) && apiErr.HttpCode != 0 {
		return c.Status(apiErr.HttpCode).JSON(apiErr)
	}

	var jsonUnmarshalTypeError *fiber.UnmarshalTypeError
	if errors.As(err, &jsonUnmarshalTypeError) {
		return c.Status(fiber.StatusBadRequest).JSON(apierror.New(apierror.Error{
			Message: jsonUnmarshalTypeError.Error(),
			Field:   jsonUnmarshalTypeError.Field,
		}).SetHttpCode(fiber.StatusBadRequest))
	}

	c.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	return c.Status(code).JSON(apierror.New(apierror.Error{
		Message: "Internal Server Error",
	}).SetHttpCode(code))
}
