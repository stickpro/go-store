package admin

import (
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/stickpro/go-store/internal/delivery/http/request/collection_request"
	"github.com/stickpro/go-store/internal/delivery/http/response"
	"github.com/stickpro/go-store/internal/delivery/http/response/collection_response"
	"github.com/stickpro/go-store/internal/dto"
	"github.com/stickpro/go-store/internal/tools/apierror"

	// swag-gen import
	_ "github.com/stickpro/go-store/internal/models"
	_ "github.com/stickpro/go-store/internal/storage/base"
)

// getCollections is a function to get collections with pagination
//
//	@Summary		Collections
//	@Description	Get collections with pagination
//	@Tags			Collections
//	@Accept			json
//	@Produce		json
//	@Param			string	query		collection_request.GetCollectionWithPagination	true	"GetCollectionWithPagination"
//	@Success		200		{object}	response.Result[base.FindResponseWithFullPagination[models.Collection]]
//	@Failure		400		{object}	apierror.Errors
//	@Failure		404		{object}	apierror.Errors
//	@Failure		500		{object}	apierror.Errors
//	@Router			/v1/collection/ [get]
//
//	@Security		BearerAuth
func (h *Handler) getCollections(c fiber.Ctx) error {
	req := &collection_request.GetCollectionWithPagination{}
	if err := c.Bind().Query(req); err != nil {
		return err
	}
	cls, err := h.services.CollectionService.GetCollectionsWithPagination(c.Context(), dto.GetCollectionDTO{
		Page:     req.Page,
		PageSize: req.PageSize,
	})
	if err != nil {
		return h.handleError(err, "collection")
	}
	return c.JSON(response.OkByData(cls))
}

// createCollection is a function to create a new collection
//
//	@Summary		Create Collection
//	@Description	Create a new collection
//	@Tags			Collections
//	@Accept			json
//	@Produce		json
//	@Param			create	body		collection_request.CreateCollectionRequest	true	"CreateCollectionRequest"
//	@Success		200		{object}	response.Result[collection_response.CollectionResponse]
//	@Failure		400		{object}	apierror.Errors
//	@Failure		422		{object}	apierror.Errors
//	@Failure		500		{object}	apierror.Errors
//	@Router			/v1/collection/ [post]
//
//	@Security		BearerAuth
func (h *Handler) createCollection(c fiber.Ctx) error {
	req := &collection_request.CreateCollectionRequest{}
	if err := c.Bind().Body(req); err != nil {
		return err
	}

	d := dto.RequestToCreateCollectionDTO(req)
	col, err := h.services.CollectionService.CreateCollection(c.Context(), d)
	if err != nil {
		return h.handleError(err, "collection")
	}
	return c.JSON(response.OkByData(collection_response.NewFromModel(col)))
}

// updateCollection is a function to update an existing collection
//
//	@Summary		Update Collection
//	@Description	Update an existing collection
//	@Tags			Collections
//	@Accept			json
//	@Produce		json
//	@Param			id		path		uuid.UUID									true	"Collection ID"
//	@Param			update	body		collection_request.UpdateCollectionRequest	true	"UpdateCollectionRequest"
//	@Success		200		{object}	response.Result[collection_response.CollectionResponse]
//	@Failure		400		{object}	apierror.Errors
//	@Failure		422		{object}	apierror.Errors
//	@Failure		500		{object}	apierror.Errors
//	@Router			/v1/collection/:id [put]
//
//	@Security		BearerAuth
func (h *Handler) updateCollection(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}

	req := &collection_request.UpdateCollectionRequest{}
	if err := c.Bind().Body(req); err != nil {
		return err
	}

	d := dto.RequestToUpdateCollectionDTO(req, id)
	col, err := h.services.CollectionService.UpdateCollection(c.Context(), d)
	if err != nil {
		return h.handleError(err, "collection")
	}
	return c.JSON(response.OkByData(collection_response.NewFromModel(col)))
}

func (h *Handler) initCollectionRoutes(v1 fiber.Router) {
	c := v1.Group("/collection")
	c.Get("/", h.getCollections)
	c.Post("/", h.createCollection)
	c.Put("/:id", h.updateCollection)
}
