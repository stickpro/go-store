package handlers

import (
	"github.com/gofiber/fiber/v3"
	"github.com/stickpro/go-store/internal/delivery/http/request/auth_request"
)

func (h *Handler) register(c fiber.Ctx) error {
	dto := &auth_request.RegisterRequest{}
	if err := c.Bind().Body(dto); err != nil {
		return err
	}

}
