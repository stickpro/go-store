package admin

import (
	"errors"

	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5"
	"github.com/stickpro/go-store/internal/delivery/middleware"
	"github.com/stickpro/go-store/internal/service"
	"github.com/stickpro/go-store/internal/tools/apierror"
	"github.com/stickpro/go-store/pkg/dbutils/pgerror"
)

type Handler struct {
	services *service.Services
}

func NewAdminHandler(services *service.Services) *Handler {
	return &Handler{
		services: services,
	}
}

func (h *Handler) InitAdminHandler(api *fiber.App) {
	secured := api.Group(
		"api/v1",
		middleware.AuthMiddleware(h.services.AuthService),
		middleware.AdminMiddleware(),
	)

	h.initCategoryRoutes(secured)
	h.initCollectionRoutes(secured)
	h.initProductRoutes(secured)
	h.initMediaRoutes(secured)
	h.initAttributeRoutes(secured)
	h.initManufacturerRoutes(secured)
	h.initOrderRoutes(secured)
	h.initDashboardRoutes(secured)
	h.initProductReviewRoutes(secured)
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
