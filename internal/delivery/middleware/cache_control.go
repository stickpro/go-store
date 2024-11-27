package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v3"
)

func CacheControlMiddleware() fiber.Handler {
	return func(c fiber.Ctx) error {
		if strings.Contains(c.Path(), "api/") {
			c.Set("Cache-Control", "no-store")
			return c.Next()
		}
		c.Set("Cache-Control", "no-store")
		c.Request().Header.Del("If-modified-since")
		return c.Next()
	}
}
