package handlers

import (
	"errors"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/stickpro/go-store/internal/delivery/http/request/viewed_request"
	"github.com/stickpro/go-store/internal/delivery/http/response"
	"github.com/stickpro/go-store/internal/delivery/http/response/viewed_response"
	"github.com/stickpro/go-store/internal/dto"
	"github.com/stickpro/go-store/internal/tools/apierror"
)

// trackViewed
//
//	@Summary		Track viewed product
//	@Tags			Viewed
//	@Description	Records a product variant as viewed for the current user or guest session
//	@Accept			json
//	@Produce		json
//	@Param			body	body		viewed_request.TrackViewedRequest	true	"Variant to track"
//	@Success		200		{object}	response.Result[any]
//	@Failure		400		{object}	apierror.Errors
//	@Failure		422		{object}	apierror.Errors
//	@Router			/api/v1/viewed-products [post]
func (h *Handler) trackViewed(c fiber.Ctx) error {
	owner, err := h.viewedOwner(c)
	if err != nil {
		return err
	}

	req := &viewed_request.TrackViewedRequest{}
	if err := c.Bind().Body(req); err != nil {
		return err
	}

	if err := h.services.ViewedService.Track(c.Context(), owner, req.VariantID); err != nil {
		return h.handleError(err, "viewed")
	}

	return c.JSON(response.OkByMessage("ok"))
}

// getViewed
//
//	@Summary		Get viewed products
//	@Tags			Viewed
//	@Description	Returns the last 10 viewed product variants for the current user or guest session
//	@Produce		json
//	@Success		200	{object}	response.Result[viewed_response.ViewedResponse]
//	@Failure		400	{object}	apierror.Errors
//	@Router			/api/v1/viewed-products [get]
func (h *Handler) getViewed(c fiber.Ctx) error {
	owner, err := h.viewedOwner(c)
	if err != nil {
		return err
	}

	viewedDTO, err := h.services.ViewedService.GetViewed(c.Context(), owner)
	if err != nil {
		return h.handleError(err, "viewed")
	}

	return c.JSON(response.OkByData(viewed_response.NewFromDTO(viewedDTO)))
}

func (h *Handler) initViewedRoutes(v1 fiber.Router) {
	v := v1.Group("/viewed-products")
	v.Post("/", h.trackViewed)
	v.Get("/", h.getViewed)
}

func (h *Handler) viewedOwner(c fiber.Ctx) (dto.Owner, error) {
	if user, err := loadAuthUser(c); err == nil {
		return dto.Owner{UserID: &user.ID}, nil
	}

	sessionStr := c.Cookies("session_id")
	if sessionStr == "" {
		sessionStr = c.Get("X-Session-ID")
	}

	if sessionStr == "" {
		return dto.Owner{}, apierror.New().
			AddError(errors.New("session_id cookie or X-Session-ID header is required")).
			SetHttpCode(fiber.StatusBadRequest)
	}

	sessionID, err := uuid.Parse(sessionStr)
	if err != nil {
		return dto.Owner{}, apierror.New().
			AddError(errors.New("session_id must be a valid UUID")).
			SetHttpCode(fiber.StatusBadRequest)
	}

	return dto.Owner{SessionID: &sessionID}, nil
}
