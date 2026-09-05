package handlers

import (
	"errors"

	"github.com/gofiber/fiber/v3"

	"github.com/stickpro/go-store/internal/delivery/http/request/auth_request"
	"github.com/stickpro/go-store/internal/delivery/http/response"
	"github.com/stickpro/go-store/internal/delivery/http/response/auth_response"
	"github.com/stickpro/go-store/internal/dto"
	"github.com/stickpro/go-store/internal/service/auth"
	"github.com/stickpro/go-store/internal/tools/apierror"
)

// requestCode starts passwordless auth (registration or login).
//
//	@Summary		Request login code
//	@Description	Emails a 6-digit one-time code. Works for both new and existing accounts. Admin accounts are ignored (they log in with a password); the response is the same either way.
//	@Tags			Auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		auth_request.SendCodeRequest	true	"Email to send the code to"
//	@Success		200		{object}	response.Result[auth_response.SendCodeResponse]
//	@Failure		422		{object}	apierror.Errors
//	@Failure		429		{object}	apierror.Errors
//	@Router			/v1/auth/code [post]
func (h *Handler) requestCode(c fiber.Ctx) error {
	req := &auth_request.SendCodeRequest{}
	if err := c.Bind().Body(req); err != nil {
		return err
	}

	if err := h.services.AuthService.RequestCode(c.Context(), req.Email); err != nil {
		return h.authError(err)
	}

	return c.JSON(response.OkByData(auth_response.SendCodeResponse{Sent: true}))
}

// verifyCode exchanges a one-time code for an auth token, creating the account on
// first login.
//
//	@Summary		Verify login code
//	@Description	Validates the emailed code and returns a bearer token. Creates the account if it does not exist yet.
//	@Tags			Auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		auth_request.VerifyCodeRequest	true	"Email + code"
//	@Success		200		{object}	response.Result[auth_response.AuthResponse]
//	@Failure		400		{object}	apierror.Errors
//	@Failure		422		{object}	apierror.Errors
//	@Router			/v1/auth/verify [post]
func (h *Handler) verifyCode(c fiber.Ctx) error {
	req := &auth_request.VerifyCodeRequest{}
	if err := c.Bind().Body(req); err != nil {
		return err
	}

	token, user, err := h.services.AuthService.VerifyCode(c.Context(), req.Email, req.Code)
	if err != nil {
		return h.authError(err)
	}

	if sessionID := parseCartSessionID(c); sessionID != nil {
		if _, mErr := h.services.CartService.MergeCarts(c.Context(), *sessionID, user.ID); mErr != nil {
			h.logger.Errorw("auth: cart merge failed", "error", mErr, "user_id", user.ID)
		}
		if mErr := h.services.ViewedService.MergeViewed(c.Context(), *sessionID, user.ID); mErr != nil {
			h.logger.Errorw("auth: viewed merge failed", "error", mErr, "user_id", user.ID)
		}
	}

	return c.JSON(response.OkByData(auth_response.AuthResponse{Token: token.FullToken}))
}

// login authenticates an admin account by email + password. Regular users have
// no password and must use the code flow.
//
//	@Summary		Admin login
//	@Description	Email + password login. Admin accounts only.
//	@Tags			Auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		auth_request.AuthRequest	true	"Email + password"
//	@Success		200		{object}	response.Result[auth_response.AuthResponse]
//	@Failure		401		{object}	apierror.Errors
//	@Failure		422		{object}	apierror.Errors
//	@Router			/v1/auth/login [post]
func (h *Handler) login(c fiber.Ctx) error {
	req := &auth_request.AuthRequest{}
	if err := c.Bind().Body(req); err != nil {
		return err
	}

	token, err := h.services.AuthService.Auth(c.Context(), dto.AuthDTO{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		return h.authError(err)
	}

	return c.JSON(response.OkByData(auth_response.AuthResponse{Token: token.FullToken}))
}

// authError maps auth-service sentinel errors to HTTP responses without leaking
// which part of the check failed.
func (h *Handler) authError(err error) error {
	switch {
	case errors.Is(err, auth.ErrResendTooSoon), errors.Is(err, auth.ErrRateLimited):
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusTooManyRequests)
	case errors.Is(err, auth.ErrCodeInvalid),
		errors.Is(err, auth.ErrCodeExpired),
		errors.Is(err, auth.ErrTooManyAttempts):
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	case errors.Is(err, auth.ErrInvalidCredentials):
		return apierror.New().AddError(errors.New("invalid email or password")).SetHttpCode(fiber.StatusUnauthorized)
	case errors.Is(err, auth.ErrUsePasswordLogin):
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusConflict)
	case errors.Is(err, auth.ErrUserBanned):
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusForbidden)
	default:
		h.logger.Errorw("auth: unexpected error", "error", err)
		return apierror.New().AddError(errors.New("authentication failed")).SetHttpCode(fiber.StatusInternalServerError)
	}
}

func (h *Handler) initAuthRoutes(v1 fiber.Router) {
	a := v1.Group("/auth")
	a.Post("/code", h.requestCode)
	a.Post("/verify", h.verifyCode)
	a.Post("/login", h.login)
}
