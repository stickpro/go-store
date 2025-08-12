package handlers

import (
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/stickpro/go-store/internal/delivery/http/response"
	"github.com/stickpro/go-store/internal/delivery/http/response/collection_response"
	"github.com/stickpro/go-store/internal/tools/apierror"
)

// getCollectionByID is a function get collection by id
//
//	@Summary		Collection
//	@Description	Get collection by ID
//	@Tags			Collection
//	@Accept			json
//	@Produce		json
//	@Param			id	path		uuid.UUID	true	"Collection ID"
//	@Success		200	{object}	response.Result[collection_response.CollectionResponse]
//	@Failure		400	{object}	apierror.Errors
//	@Failure		404	{object}	apierror.Errors
//	@Failure		500	{object}	apierror.Errors
//	@Router			/v1/collection/id/:id/ [get]
func (h Handler) getCollectionByID(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}
	collectionDTO, err := h.services.CollectionService.GetCollectionByID(c.Context(), id)
	if err != nil {
		return h.handleError(err, "collection")
	}

	return c.JSON(response.OkByData(collection_response.NewFromDTO(collectionDTO)))
}

// getCollectionBySlug is a function get collection by slug
//
//	@Summary		Collection
//	@Description	Get collection by slug
//	@Tags			Collection
//	@Accept			json
//	@Produce		json
//	@Param			slug	path		string	true	"Collection Slug"
//	@Success		200		{object}	response.Result[collection_response.CollectionResponse]
//	@Failure		400		{object}	apierror.Errors
//	@Failure		404		{object}	apierror.Errors
//	@Failure		500		{object}	apierror.Errors
//	@Router			/v1/collection/:slug/ [get]
func (h Handler) getCollectionSlug(c fiber.Ctx) error {
	slug := c.Params("slug")
	collectionDTO, err := h.services.CollectionService.GetCollectionBySlug(c.Context(), slug)
	if err != nil {
		return h.handleError(err, "collection")
	}
	return c.JSON(response.OkByData(collection_response.NewFromDTO(collectionDTO)))
}

func (h Handler) initCollectionRoutes(v1 fiber.Router) {
	c := v1.Group("/collection")
	c.Get("/:slug", h.getCategoryBySlug)
	c.Get("/id/:id", h.getCollectionByID)
}
