package admin

import (
	"errors"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/stickpro/go-store/internal/delivery/http/request/attribute_request"
	"github.com/stickpro/go-store/internal/delivery/http/response"
	"github.com/stickpro/go-store/internal/delivery/http/response/attribute_response"
	"github.com/stickpro/go-store/internal/dto"
	"github.com/stickpro/go-store/internal/tools/apierror"
	"github.com/stickpro/go-store/pkg/dbutils/pgerror"
)

// getAttributeGroups is a function get attribute groups with pagination
//
//	@Summary		Get attribute groups
//	@Description	Get attribute groups with pagination
//	@Tags			Attribute
//	@Accept			json
//	@Produce		json
//	@Param			string	query		attribute_request.GetAttributeGroupWithPagination	true	"GetAttributeGroupWithPagination"
//	@Success		200		{object}	response.Result[[]attribute_response.AttributeGroupResponse]
//	@Failure		400		{object}	apierror.Errors
//	@Failure		422		{object}	apierror.Errors
//	@Failure		500		{object}	apierror.Errors
//	@Router			/v1/attribute/ [GET]
//
//	@Security		BearerAuth
func (h *Handler) getAttributeGroups(c fiber.Ctx) error {
	req := &attribute_request.GetAttributeGroupWithPagination{}
	if err := c.Bind().Query(req); err != nil {
		return err
	}

	d := dto.GetDTO{Page: req.Page, PageSize: req.PageSize}
	aGroups, err := h.services.AttributeService.GetAttributeGroup(c.Context(), d)
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusInternalServerError)
	}

	return c.JSON(response.OkByData(aGroups))
}

// createAttributeGroup a is a function create attribute group
//
//	@Summary		Create attribute group
//	@Description	Create attribute group
//	@Tags			Attribute
//	@Accept			json
//	@Produce		json
//	@Param			create	body		attribute_request.CreateAttributeGroupRequest	true	"Create category"
//	@Success		200		{object}	response.Result[attribute_response.AttributeGroupResponse]
//	@Failure		400		{object}	apierror.Errors
//	@Failure		422		{object}	apierror.Errors
//	@Failure		500		{object}	apierror.Errors
//	@Router			/v1/attribute/ [POST]
//
//	@Security		BearerAuth
func (h *Handler) createAttributeGroup(c fiber.Ctx) error {
	req := &attribute_request.CreateAttributeGroupRequest{}
	if err := c.Bind().Body(req); err != nil {
		return err
	}

	d := dto.RequestToCreateAttributeGroupDTO(req)
	aGroup, err := h.services.AttributeService.CreateAttributeGroup(c.Context(), d)
	if err != nil {
		var uniqueErr *pgerror.UniqueConstraintError
		if errors.As(err, &uniqueErr) {
			return apierror.New().AddError(uniqueErr).SetHttpCode(fiber.StatusUnprocessableEntity)
		}
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusInternalServerError)
	}
	return c.JSON(response.OkByData(attribute_response.NewFromGroupModel(aGroup)))
}

// updateAttributeGroup is a function update attribute group
//
//	@Summary		Update attribute group
//	@Description	Update attribute group
//	@Tags			Attribute
//	@Accept			json
//	@Produce		json
//	@Param			update	body		attribute_request.UpdateAttributeGroupRequest	true	"Update attribute group"
//	@Success		200		{object}	response.Result[attribute_response.AttributeGroupResponse]
//	@Failure		400		{object}	apierror.Errors
//	@Failure		422		{object}	apierror.Errors
//	@Failure		500		{object}	apierror.Errors
//	@Router			/v1/attribute/:id [PUT]
//
//	@Security		BearerAuth
func (h *Handler) updateAttributeGroup(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}

	req := &attribute_request.UpdateAttributeGroupRequest{}
	if err := c.Bind().Body(req); err != nil {
		return err
	}

	d := dto.RequestToUpdateAttributeGroupDTO(req)
	aGroup, err := h.services.AttributeService.UpdateAttributeGroup(c.Context(), d, id)
	if err != nil {
		var uniqueErr *pgerror.UniqueConstraintError
		if errors.As(err, &uniqueErr) {
			return apierror.New().AddError(uniqueErr).SetHttpCode(fiber.StatusUnprocessableEntity)
		}
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusInternalServerError)
	}
	return c.JSON(response.OkByData(attribute_response.NewFromGroupModel(aGroup)))
}

// getAttributeGroupByID is a function get attribute group by ID
//
//	@Summary		Get attribute group by ID
//	@Description	Get attribute group by ID
//	@Tags			Attribute
//	@Accept			json
//	@Produce		json
//	@Param			id	path		uuid.UUID	true	"Attribute group ID"
//	@Success		200	{object}	response.Result[attribute_response.AttributeGroupResponse
//	@Failure		400	{object}	apierror.Errors
//	@Failure		422	{object}	apierror.Errors
//	@Failure		500	{object}	apierror.Errors
//	@Router			/v1/attribute/group/:id [GET]
//
//	@Security		BearerAuth
func (h *Handler) getAttributeGroupByID(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}

	aGroup, err := h.services.AttributeService.GetAttributeGroupByID(c.Context(), id)
	if err != nil {
		return h.handleError(err, "attribute group")
	}

	return c.JSON(response.OkByData(attribute_response.NewFromGroupModel(aGroup)))
}

// deleteAttributeGroup is a function delete attribute group by ID
//
//	@Summary		Delete attribute group by ID
//	@Description	Delete attribute group by ID
//	@Tags			Attribute
//	@Accept			json
//	@Produce		json
//	@Param			id	path		uuid.UUID	true	"Attribute group ID"
//	@Success		200	{object}	response.Result[string]
//	@Failure		400	{object}	apierror.Errors
//	@Failure		422	{object}	apierror.Errors
//	@Failure		500	{object}	apierror.Errors
//	@Router			/v1/attribute/group/:id [DELETE]
//
//	@Security		BearerAuth
func (h *Handler) deleteAttributeGroup(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}

	if err := h.services.AttributeService.DeleteAttributeGroup(c.Context(), id); err != nil {
		return h.handleError(err, "attribute group")
	}

	return c.JSON(response.OkByMessage("Attribute groups successfully deleted"))
}

func (h *Handler) createAttribute(c fiber.Ctx) error {
	req := &attribute_request.CreateAttributeRequest{}
	if err := c.Bind().Body(req); err != nil {
		return err
	}

	d := dto.RequestToCreateAttributeDTO(req)
	attr, err := h.services.AttributeService.CreateAttribute(c.Context(), d)
	if err != nil {
		var uniqueErr *pgerror.UniqueConstraintError
		if errors.As(err, &uniqueErr) {
			return apierror.New().AddError(uniqueErr).SetHttpCode(fiber.StatusUnprocessableEntity)
		}
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusInternalServerError)
	}
	return c.JSON(response.OkByData(attribute_response.NewFromAttributeModel(attr)))
}

func (h *Handler) initAttributeRoutes(v1 fiber.Router) {
	c := v1.Group("/attribute")
	c.Get("/group/", h.getAttributeGroups)
	c.Get("group/:id", h.getAttributeGroupByID)
	c.Post("/group/", h.createAttributeGroup)
	c.Put("/group/:id", h.updateAttributeGroup)
	c.Delete("/group/:id", h.deleteAttributeGroup)
}
