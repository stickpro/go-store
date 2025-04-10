package handlers

import (
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/stickpro/go-store/internal/delivery/http/request/manufacturer_request"
	"github.com/stickpro/go-store/internal/delivery/http/response"
	"github.com/stickpro/go-store/internal/delivery/http/response/manufacturer_response"
	"github.com/stickpro/go-store/internal/service/manufacturer"
	"github.com/stickpro/go-store/internal/tools/apierror"
)

// getManufacturerBySlug is a function get Manufacturer by slug
//
//	@Summary		Manufacturer
//	@Description	Get Manufacturer by slug
//	@Tags			Manufacturer
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string	true	"Manufacturer Slug"
//	@Success		200	{object}	response.Result[manufacturer_response.ManufacturerResponse]
//	@Failure		400	{object}	apierror.Errors
//	@Failure		404	{object}	apierror.Errors
//	@Failure		500	{object}	apierror.Errors
//	@Router			/v1/manufacturer/:slug/ [get]
func (h Handler) getManufacturerBySlug(c fiber.Ctx) error {
	slug := c.Params("slug")
	mfc, err := h.services.ManufacturerService.GetManufacturerBySlug(c.Context(), slug)
	if err != nil {
		return h.handleError(err, "manufacturer")
	}
	return c.JSON(response.OkByData(manufacturer_response.NewFromModel(mfc)))
}

// getManufacturerByID is a function get Manufacturer by ID
//
//	@Summary		Manufacturer
//	@Description	Get Manufacturer by ID
//	@Tags			Manufacturer
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string	true	"Manufacturer ID"
//	@Success		200	{object}	response.Result[manufacturer_response.ManufacturerResponse]
//	@Failure		400	{object}	apierror.Errors
//	@Failure		404	{object}	apierror.Errors
//	@Failure		500	{object}	apierror.Errors
//	@Router			/v1/manufacturer/id/:id/ [get]
func (h Handler) getManufacturerByID(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}
	mfc, err := h.services.ManufacturerService.GetManufacturerByID(c.Context(), id)
	if err != nil {
		return h.handleError(err, "manufacturer")
	}
	return c.JSON(response.OkByData(manufacturer_response.NewFromModel(mfc)))
}

func (h Handler) getManufacturers(c fiber.Ctx) error {
	req := &manufacturer_request.GetManufacturerWithPagination{}
	if err := c.Bind().Query(req); err != nil {
		return err
	}

	mfcs, err := h.services.ManufacturerService.GetManufacturersWithPagination(c.Context(), manufacturer.GetDTO{Page: req.Page, PageSize: req.PageSize})
	if err != nil {
		return h.handleError(err, "manufacturers")
	}
	return c.JSON(response.OkByData(manufacturer_response.NewPaginatedFromFindRows(mfcs)))
}

func (h *Handler) initManufacturerRoutes(v1 fiber.Router) {
	m := v1.Group("/manufacturer")
	m.Get("/:slug", h.getManufacturerBySlug)
	m.Get("/id/:id", h.getManufacturerByID)
}
