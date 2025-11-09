package handlers

import (
	"errors"
	"reflect"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/binder"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/stickpro/go-store/internal/delivery/middleware"
	"github.com/stickpro/go-store/internal/models"
	"github.com/stickpro/go-store/internal/service"
	"github.com/stickpro/go-store/internal/tools/apierror"
	"github.com/stickpro/go-store/pkg/dbutils/pgerror"
	"github.com/stickpro/go-store/pkg/logger"
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
	h.initCollectionRoutes(v1)
	h.initGeoRoutes(v1)
	h.initManufacturerRoutes(v1)
	h.initProductReviewRoutes(v1)

	secured := v1.Group(
		"/",
		middleware.AuthMiddleware(h.services.AuthService),
		middleware.AdminMiddleware(),
	)
	h.initUserRoutes(secured)
}

func (h *Handler) configureBinders() {
	binder.SetParserDecoder(binder.ParserConfig{
		IgnoreUnknownKeys: true,
		ParserType: []binder.ParserType{
			{
				CustomType: uuid.UUID{},
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

func (h *Handler) handleError(err error, modelName string) error {
	var (
		notFoundErr *pgerror.NotFoundError
		uniqueErr   *pgerror.UniqueConstraintError
		fkErr       *pgerror.ForeignKeyViolationError
	)

	if errors.Is(err, pgx.ErrNoRows) || errors.As(err, &notFoundErr) {
		return apierror.New().AddError(errors.New(modelName + " not found")).SetHttpCode(fiber.StatusNotFound)
	}
	if errors.As(err, &uniqueErr) {
		return apierror.New().AddError(uniqueErr).SetHttpCode(fiber.StatusUnprocessableEntity)
	}
	if errors.As(err, &fkErr) {
		return apierror.New().AddError(errors.New("referenced " + modelName + " does not exist")).SetHttpCode(fiber.StatusUnprocessableEntity)
	}
	return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
}

func loadAuthUser(c fiber.Ctx) (*models.User, error) {
	user, ok := c.Locals("user").(*models.User)
	if !ok {
		return nil, apierror.New().AddError(errors.New("undefined user")).SetHttpCode(fiber.StatusUnauthorized)
	}
	return user, nil
}
