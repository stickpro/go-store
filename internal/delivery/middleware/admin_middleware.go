package middleware

import (
	"errors"
	"github.com/gofiber/fiber/v3"
	"github.com/stickpro/go-store/internal/models"
	"github.com/stickpro/go-store/internal/tools/apierror"
)

func AdminMiddleware() fiber.Handler {
	return func(c fiber.Ctx) error {
		user, ok := c.Locals("user").(*models.User)
		if !ok {
			return apierror.New().AddError(errors.New("undefined user")).SetHttpCode(fiber.StatusUnauthorized)
		}

		if !user.IsAdmin.Valid {
			return apierror.New().AddError(errors.New("unauthorized")).SetHttpCode(fiber.StatusUnauthorized)
		}
		return c.Next()
	}
}
