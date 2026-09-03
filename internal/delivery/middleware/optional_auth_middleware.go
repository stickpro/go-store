package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/stickpro/go-store/internal/service/auth"
	"github.com/stickpro/go-store/internal/tools/hash"
)

// OptionalAuthMiddleware populates c.Locals("user") when a valid Bearer token is
// present, and otherwise proceeds without error. Use it on endpoints that serve
// both guests and authenticated users (e.g. checkout).
func OptionalAuthMiddleware(auth auth.IAuthService) fiber.Handler {
	return func(c fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Next()
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return c.Next()
		}

		user, err := auth.GetUserByToken(c.Context(), hash.SHA256(parts[1]))
		if err != nil {
			return c.Next()
		}

		c.Locals("user", user)
		return c.Next()
	}
}
