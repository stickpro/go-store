package handlers

import (
	"errors"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/binder"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/stickpro/go-store/internal/service"
	"github.com/stickpro/go-store/internal/tools/apierror"
	"github.com/stickpro/go-store/pkg/logger"
	"reflect"
)

type Handler struct {
	services *service.Services
	logger   logger.Logger
}

func NewHandler(services *service.Services, logger logger.Logger) *Handler {
	return &Handler{
		services: services,
		logger:   logger,
	}
}

func (h *Handler) InitHandler(api *fiber.App) {
	h.configureBinders()

	v1 := api.Group("api/v1")

	h.initAuthRoutes(v1)
	h.initCategoryRoutes(v1)
	h.initProductRoutes(v1)
}

func (h *Handler) configureBinders() {
	binder.SetParserDecoder(binder.ParserConfig{
		IgnoreUnknownKeys: true,
		ParserType: []binder.ParserType{
			{
				Customtype: uuid.UUID{},
				Converter: func(value string) reflect.Value {
					if v, err := uuid.Parse(value); err == nil {
						return reflect.ValueOf(v)
					}

					return reflect.Value{}
				},
			},
		},
		ZeroEmpty: true,
	})
}

func (h Handler) handleError(err error, modelName string) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return apierror.New().AddError(errors.New(modelName + " not found")).SetHttpCode(fiber.StatusNotFound)
	}
	return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
}
