package admin

import (
	"github.com/gofiber/fiber/v3"
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

	h.initCategoryRoutes(v1)
}
