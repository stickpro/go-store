package admin

import (
	"github.com/gofiber/fiber/v3"
	"github.com/stickpro/go-store/internal/delivery/middleware"
	"github.com/stickpro/go-store/internal/service"
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
	h.initProductRoutes(secured)
	h.initMediaRoutes(secured)
}
