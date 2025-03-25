package middleware

import (
	"github.com/gofiber/fiber/v3"
	"github.com/stickpro/go-store/internal/service/auth"
	"github.com/stickpro/go-store/internal/tools/apierror"
	"github.com/stickpro/go-store/internal/tools/hash"
	"strings"
)

func AuthMiddleware(auth auth.IAuthService) fiber.Handler {
	return func(c fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return apierror.New().AddError(fiber.ErrUnauthorized).SetHttpCode(fiber.StatusUnauthorized)
		}

		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			return apierror.New().AddError(fiber.ErrUnauthorized).SetHttpCode(fiber.StatusUnauthorized)
		}

		token := tokenParts[1]
		hashedToken := hash.SHA256(token)

		user, err := auth.GetUserByToken(c.Context(), hashedToken)
		if err != nil {
			return apierror.New().AddError(fiber.ErrUnauthorized).SetHttpCode(fiber.StatusUnauthorized)
		}

		c.Locals("user", user)
		return c.Next()
	}
}
