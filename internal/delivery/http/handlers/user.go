package handlers

import (
	"github.com/gofiber/fiber/v3"
	"github.com/stickpro/go-store/internal/delivery/http/response"
	"github.com/stickpro/go-store/internal/delivery/http/response/user_response"
	_ "github.com/stickpro/go-store/internal/tools/apierror"
)

// authUser is a function get user info for auth
//
//	@Summary		Auth user
//	@Description	Auth a user
//	@Tags			User
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	response.Result[user_response.UserInfoResponse]
//	@Failure		400	{object}	apierror.Errors
//	@Failure		401	{object}	apierror.Errors
//	@Failure		422	{object}	apierror.Errors
//	@Failure		503	{object}	apierror.Errors
//	@Router			/v1/user [get]
//	@Security		BearerAuth
func (h *Handler) authUser(c fiber.Ctx) error {
	user, err := loadAuthUser(c)
	if err != nil {
		return err
	}

	return c.JSON(response.OkByData(user_response.NewFromModel(user)))
}

func (h *Handler) initUserRoutes(v1 fiber.Router) {
	u := v1.Group("/user")
	u.Get("/info", h.authUser)
}
