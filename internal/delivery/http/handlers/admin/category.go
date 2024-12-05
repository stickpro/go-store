package admin

import (
	"github.com/gofiber/fiber/v3"
	"github.com/stickpro/go-store/internal/delivery/http/request/category_request"
)

func (h Handler) createCategory(c fiber.Ctx) error {
	req := &category_request.CreateCategoryRequest{}
	if err := c.Bind().Body(req); err != nil {
		return err
	}

	// wip
	return nil
}

func (h *Handler) initCategoryRoutes(v1 fiber.Router) {
	c := v1.Group("/category")
	c.Post("/", h.createCategory)
}
