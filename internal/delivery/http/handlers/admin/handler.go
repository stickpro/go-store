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
	v1 := api.Group("api/v1")

	secured := v1.Group(
		"/",
		middleware.AuthMiddleware(h.services.AuthService),
		middleware.AdminMiddleware(),
	)

	h.initCategoryRoutes(secured)
	h.initCollectionRoutes(secured)
	h.initProductRoutes(secured)
	h.initMediaRoutes(secured)
}

func (h Handler) handleError(err error, modelName string) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return apierror.New().AddError(errors.New(modelName + " not found")).SetHttpCode(fiber.StatusNotFound)
	}
	var uniqueErr *pgerror.UniqueConstraintError
	if errors.As(err, &uniqueErr) {
		return apierror.New().AddError(uniqueErr).SetHttpCode(fiber.StatusUnprocessableEntity)
	}
	return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
}
