package admin

import (
	"errors"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/stickpro/go-store/internal/delivery/http/request/manufacturer_request"
	"github.com/stickpro/go-store/internal/delivery/http/response"
	"github.com/stickpro/go-store/internal/delivery/http/response/manufacturer_response"
	"github.com/stickpro/go-store/internal/service/manufacturer"
	"github.com/stickpro/go-store/internal/tools/apierror"
	"github.com/stickpro/go-store/pkg/dbutils/pgerror"
)

// createManufacturer is a function create manufacturer
//
//	@Summary		Create Manufacturer
//	@Description	Create Manufacturer
//	@Tags			Manufacturer
//	@Accept			json
//	@Produce		json
//	@Param			create	body		manufacturer_request.CreateManufacturerRequest	true	"Create manufacturer"
//	@Success		200		{object}	response.Result[manufacturer_response.ManufacturerResponse]
//	@Failure		400		{object}	apierror.Errors
//	@Failure		422		{object}	apierror.Errors
//	@Failure		500		{object}	apierror.Errors
//	@Router			/v1/manufacturer/ [POST]
func (h *Handler) createManufacturer(c fiber.Ctx) error {
	req := &manufacturer_request.CreateManufacturerRequest{}
	if err := c.Bind().Body(req); err != nil {
		return err
	}

	dto := manufacturer.RequestToCreateDTO(req)
	mfc, err := h.services.ManufacturerService.CreateManufacturer(c.Context(), dto)
	if err != nil {
		var uniqueErr *pgerror.UniqueConstraintError
		if errors.As(err, &uniqueErr) {
			return apierror.New().AddError(uniqueErr).SetHttpCode(fiber.StatusUnprocessableEntity)
		}
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusInternalServerError)
	}
	return c.JSON(response.OkByData(manufacturer_response.NewFromModel(mfc)))
}

// updateManufacturer is a function update manufacturer
//
//	@Summary		Update Manufacturer
//	@Description	Update Manufacturer
//	@Tags			Manufacturer
//	@Accept			json
//	@Produce		json
//	@Param			create	body		manufacturer_request.UpdateManufacturerRequest	true	"Update manufacturer"
//	@Success		200		{object}	response.Result[manufacturer_response.ManufacturerResponse]
//	@Failure		400		{object}	apierror.Errors
//	@Failure		422		{object}	apierror.Errors
//	@Failure		500		{object}	apierror.Errors
//	@Router			/v1/manufacturer/:id [PUT]
func (h *Handler) updateManufacturer(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}

	req := &manufacturer_request.UpdateManufacturerRequest{}
	if err := c.Bind().Body(req); err != nil {
		return err
	}

	dto := manufacturer.RequestToUpdateDTO(req, id)
	mfc, err := h.services.ManufacturerService.UpdateManufacturer(c.Context(), dto)
	if err != nil {
		var uniqueErr *pgerror.UniqueConstraintError
		if errors.As(err, &uniqueErr) {
			return apierror.New().AddError(uniqueErr).SetHttpCode(fiber.StatusUnprocessableEntity)
		}
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusInternalServerError)
	}
	return c.JSON(response.OkByData(manufacturer_response.NewFromModel(mfc)))
}

func (h *Handler) initManufacturerRoutes(v1 fiber.Router) {
	m := v1.Group("/manufacturer")
	m.Post("/", h.createManufacturer)
	m.Put("/:id", h.updateManufacturer)
}
