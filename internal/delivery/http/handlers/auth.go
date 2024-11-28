package handlers

import (
	errs "github.com/stickpro/go-store/internal/delivery/http/errors"
	"github.com/stickpro/go-store/internal/delivery/http/request/auth_request"
	"github.com/stickpro/go-store/internal/delivery/http/response"
	"github.com/stickpro/go-store/internal/delivery/http/response/auth_response"
	"github.com/stickpro/go-store/internal/service/auth"
	"github.com/stickpro/go-store/internal/tools/apierror"

	"github.com/gofiber/fiber/v3"
)

// register is a function to register a new user
//
//	@Summary		Register user
//	@Description	Register a new user
//	@Tags			Auth
//	@Accept			json
//	@Produce		json
//	@Param			register	body		auth_request.RegisterRequest	true	"Register account"
//	@Success		200			{object}	response.Result[auth_response.RegisterUserResponse]
//	@Failure		422			{object}	apierror.Errors
//	@Failure		503			{object}	apierror.Errors
//	@Router			/v1/auth/register [post]
func (h *Handler) register(c fiber.Ctx) error {
	request := &auth_request.RegisterRequest{}
	if err := c.Bind().Body(request); err != nil {
		return err
	}

	user, err := h.services.AuthService.RegisterUser(c.Context(), auth.RegisterDTO{
		Email:    request.Email,
		Password: request.Password,
		Location: request.Location,
		Language: request.Language,
	})

	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}

	token, err := h.services.AuthService.AuthByUser(c.Context(), user)
	if err != nil {
		return apierror.New(errs.ErrNoMatchesFound).SetHttpCode(fiber.StatusBadRequest)
	}

	return c.JSON(response.OkByData(auth_response.RegisterUserResponse{
		Token: token.FullToken,
	}))
}

// login is a function to auth a user
//
//	@Summary		Auth user
//	@Description	Auth a user
//	@Tags			Auth
//	@Accept			json
//	@Produce		json
//	@Param			register	body		auth_request.AuthRequest	true	"Register account"
//	@Success		200			{object}	response.Result[auth_response.AuthResponse]
//	@Failure		400			{object}	apierror.Errors
//	@Failure		422			{object}	apierror.Errors
//	@Failure		503			{object}	apierror.Errors
//	@Router			/v1/auth/login [post]
func (h *Handler) login(c fiber.Ctx) error {
	request := &auth_request.AuthRequest{}
	if err := c.Bind().Body(request); err != nil {
		return err
	}
	ctx := c.Context()
	token, err := h.services.AuthService.Auth(ctx, auth.AuthDTO{
		Email:    request.Email,
		Password: request.Password,
	})

	if token == nil || err != nil {
		return apierror.New(errs.ErrNoMatchesFound).SetHttpCode(fiber.StatusBadRequest)
	}

	return c.JSON(response.OkByData(auth_response.AuthResponse{
		Token: token.FullToken,
	}))
}

func (h *Handler) initAuthRoutes(v1 fiber.Router) {
	a := v1.Group("/auth")
	a.Post("/register", h.register)
	a.Post("/login", h.login)
}
