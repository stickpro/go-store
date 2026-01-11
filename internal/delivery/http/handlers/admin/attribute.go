package admin

import (
	"fmt"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/stickpro/go-store/internal/delivery/http/request/attribute_request"
	"github.com/stickpro/go-store/internal/delivery/http/response"
	"github.com/stickpro/go-store/internal/delivery/http/response/attribute_response"
	"github.com/stickpro/go-store/internal/dto"
	"github.com/stickpro/go-store/internal/tools/apierror"
	// swag-gen
	_ "github.com/stickpro/go-store/internal/models"
	_ "github.com/stickpro/go-store/internal/storage/base"
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
//	@Router			/v1/attribute-group [GET]
//
//	@Security		BearerAuth
func (h *Handler) getAttributeGroups(c fiber.Ctx) error {
	req := &attribute_request.GetAttributeGroupWithPagination{}
	if err := c.Bind().Query(req); err != nil {
		return err
	}

	d := dto.GetDTO{Page: req.Page, PageSize: req.PageSize}
	aGroups, err := h.services.AttributeService.GetAttributeGroups(c.Context(), d)
	if err != nil {
		return h.handleError(err, "attribute group")
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
//	@Router			/v1/attribute-group [POST]
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
		return h.handleError(err, "attribute group")
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
//	@Router			/v1/attribute-group/:id [PUT]
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
		return h.handleError(err, "attribute group")
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
//	@Success		200	{object}	response.Result[attribute_response.AttributeGroupResponse]
//	@Failure		400	{object}	apierror.Errors
//	@Failure		422	{object}	apierror.Errors
//	@Failure		500	{object}	apierror.Errors
//	@Router			/v1/attribute-group/:id [GET]
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
//	@Router			/v1/attribute-group/:id [DELETE]
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

// createAttribute is a function create attribute
//
//	@Summary		Create attribute
//	@Description	Create attribute
//	@Tags			Attribute
//	@Accept			json
//	@Produce		json
//	@Param			create	body		attribute_request.CreateAttributeRequest	true	"Create attribute"
//	@Success		200		{object}	response.Result[attribute_response.AttributeResponse]
//	@Failure		400		{object}	apierror.Errors
//	@Failure		422		{object}	apierror.Errors
//	@Failure		500		{object}	apierror.Errors
//	@Router			/v1/attribute/ [POST]
//
//	@Security		BearerAuth
func (h *Handler) createAttribute(c fiber.Ctx) error {
	req := &attribute_request.CreateAttributeRequest{}
	if err := c.Bind().Body(req); err != nil {
		return err
	}

	d := dto.RequestToCreateAttributeDTO(req)
	attr, err := h.services.AttributeService.CreateAttribute(c.Context(), d)
	if err != nil {
		return h.handleError(err, "attribute")
	}
	return c.JSON(response.OkByData(attribute_response.NewFromAttributeModel(attr)))
}

// getAttributeByID is a function get attribute by ID
//
//	@Summary		Get attribute by ID
//	@Description	Get attribute by ID
//	@Tags			Attribute
//	@Accept			json
//	@Produce		json
//	@Param			id	path		uuid.UUID	true	"Attribute ID"
//	@Success		200	{object}	response.Result[attribute_response.AttributeResponse]
//	@Failure		400	{object}	apierror.Errors
//	@Failure		422	{object}	apierror.Errors
//	@Failure		500	{object}	apierror.Errors
//	@Router			/v1/attribute/:id [GET]
//
//	@Security		BearerAuth
func (h *Handler) getAttributeByID(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}

	attr, err := h.services.AttributeService.GetAttributeByID(c.Context(), id)
	if err != nil {
		return h.handleError(err, "attribute")
	}

	return c.JSON(response.OkByData(attribute_response.NewFromAttributeModel(attr)))
}

// getAttributes is a function get attributes with pagination
//
//	@Summary		Get attributes
//	@Description	Get attributes with pagination
//	@Tags			Attribute
//	@Accept			json
//	@Produce		json
//	@Param			string	query		attribute_request.GetAttributeWithPagination	true	"GetAttributesWithPagination"
//	@Success		200		{object}	response.Result[[]attribute_response.AttributeResponse]
//	@Failure		400		{object}	apierror.Errors
//	@Failure		422		{object}	apierror.Errors
//	@Failure		500		{object}	apierror.Errors
//	@Router			/v1/attribute/ [GET]
//
//	@Security		BearerAuth
func (h *Handler) getAttributes(c fiber.Ctx) error {
	req := &attribute_request.GetAttributeWithPagination{}
	if err := c.Bind().Query(req); err != nil {
		return err
	}
	d := dto.GetDTO{Page: req.Page, PageSize: req.PageSize}
	attrs, err := h.services.AttributeService.GetAttributes(c.Context(), d)
	if err != nil {
		return h.handleError(err, "attributes")
	}

	return c.JSON(response.OkByData(attrs))
}

func (h *Handler) updateAttribute(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}

	req := &attribute_request.UpdateAttributeRequest{}
	if err := c.Bind().Body(req); err != nil {
		return err
	}

	d := dto.RequestToUpdateAttributeDTO(req)
	attr, err := h.services.AttributeService.UpdateAttribute(c.Context(), d, id)
	if err != nil {
		return h.handleError(err, "attribute")
	}
	return c.JSON(response.OkByData(attribute_response.NewFromAttributeModel(attr)))
}

// deleteAttribute is a function delete attribute by ID
//
//	@Summary		Delete attribute by ID
//	@Description	Delete attribute by ID
//	@Tags			Attribute
//	@Accept			json
//	@Produce		json
//	@Param			id	path		uuid.UUID	true	"Attribute ID"
//	@Success		200	{object}	response.Result[string]
//	@Failure		400	{object}	apierror.Errors
//	@Failure		422	{object}	apierror.Errors
//	@Failure		500	{object}	apierror.Errors
//	@Router			/v1/attribute/:id [DELETE]
//
//	@Security		BearerAuth
func (h *Handler) deleteAttribute(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}

	if err := h.services.AttributeService.DeleteAttribute(c.Context(), id); err != nil {
		return h.handleError(err, "attribute")
	}

	return c.JSON(response.OkByMessage("Attribute successfully deleted"))
}

// findAttribute is a function find attributes by name
//
//	@Summary		Find attribute
//	@Description	Find attribute by name with pagination
//	@Tags			Attribute
//	@Accept			json
//	@Produce		json
//	@Param			string	query		attribute_request.FindAttributeWithPagination	true	"FindAttributeWithPagination"
//	@Success		200		{object}	response.Result[base.FindResponseWithFullPagination[models.Attribute]]
//	@Failure		400		{object}	apierror.Errors
//	@Failure		500		{object}	apierror.Errors
//	@Router			/v1/attribute/find [get]
func (h *Handler) findAttribute(c fiber.Ctx) error {
	req := &attribute_request.FindAttributeWithPagination{}
	if err := c.Bind().Query(req); err != nil {
		return err
	}

	if req.Attribute == "" {
		return apierror.New().AddError(fmt.Errorf("attribute is required")).SetHttpCode(fiber.StatusBadRequest)
	}

	d := dto.GetDTO{Page: req.Page, PageSize: req.PageSize}
	attributesResponse, err := h.services.AttributeService.SearchAttributes(c.Context(), req.Attribute, d)
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}

	return c.JSON(response.OkByData(attributesResponse))
}

// findAttributeGroups is a function find attributes by name
//
//	@Summary		Find attribute group
//	@Description	Find attribute group by name with pagination
//	@Tags			Attribute
//	@Accept			json
//	@Produce		json
//	@Param			string	query		attribute_request.FindAttributeWithPagination	true	"FindAttributeWithPagination"
//	@Success		200		{object}	response.Result[base.FindResponseWithFullPagination[models.AttributeGroup]]
//	@Failure		400		{object}	apierror.Errors
//	@Failure		500		{object}	apierror.Errors
//	@Router			/v1/attribute-group/find [get]
func (h *Handler) findAttributeGroups(c fiber.Ctx) error {
	req := &attribute_request.FindAttributeWithPagination{}
	if err := c.Bind().Query(req); err != nil {
		return err
	}

	if req.Attribute == "" {
		return apierror.New().AddError(fmt.Errorf("attribute group is required")).SetHttpCode(fiber.StatusBadRequest)
	}

	d := dto.GetDTO{Page: req.Page, PageSize: req.PageSize}
	attributesResponse, err := h.services.AttributeService.SearchAttributeGroup(c.Context(), req.Attribute, d)
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}

	return c.JSON(response.OkByData(attributesResponse))
}

func (h *Handler) getAttributeValue(c fiber.Ctx) error {
	req := &attribute_request.GetAttributeValueWithPagination{}
	if err := c.Bind().Query(req); err != nil {
		return err
	}

	d := dto.GetDTO{Page: req.Page, PageSize: req.PageSize}
	aGroups, err := h.services.AttributeService.GetAttributeValues(c.Context(), d)
	if err != nil {
		return h.handleError(err, "attribute values")
	}

	return c.JSON(response.OkByData(aGroups))
}

// createAttributeValue is a function create attribute value
//
//	@Summary		Create attribute value
//	@Description	Create attribute value
//	@Tags			Attribute
//	@Accept			json
//	@Produce		json
//	@Param			create	body		attribute_request.CreateAttributeValueRequest	true	"Create attribute value"
//	@Success		200		{object}	response.Result[attribute_response.AttributeValueResponse]
//	@Failure		400		{object}	apierror.Errors
//	@Failure		422		{object}	apierror.Errors
//	@Failure		500		{object}	apierror.Errors
//	@Router			/v1/attribute-value [POST]
//
//	@Security		BearerAuth
func (h *Handler) createAttributeValue(c fiber.Ctx) error {
	req := &attribute_request.CreateAttributeValueRequest{}
	if err := c.Bind().Body(req); err != nil {
		return err
	}

	d := dto.CreateAttributeValueDTO{
		AttributeID:     req.AttributeID,
		Value:           req.Value,
		ValueNormalized: req.ValueNormalized,
		ValueNumeric:    req.ValueNumeric,
		DisplayOrder:    req.DisplayOrder,
	}

	value, err := h.services.AttributeService.CreateAttributeValue(c.Context(), d)
	if err != nil {
		return h.handleError(err, "attribute value")
	}

	return c.JSON(response.OkByData(attribute_response.NewFromAttributeValueModel(value)))
}

// getAttributeValues is a function get attribute values by attribute ID
//
//	@Summary		Get attribute values
//	@Description	Get attribute values by attribute ID
//	@Tags			Attribute
//	@Accept			json
//	@Produce		json
//	@Param			attribute_id	path		uuid.UUID	true	"Attribute ID"
//	@Success		200				{object}	response.Result[[]dto.AttributeValueDTO]
//	@Failure		400				{object}	apierror.Errors
//	@Failure		500				{object}	apierror.Errors
//	@Router			/v1/attribute/:attribute_id/values [GET]
//
//	@Security		BearerAuth
func (h *Handler) getAttributeValues(c fiber.Ctx) error {
	attributeID, err := uuid.Parse(c.Params("attribute_id"))
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}

	values, err := h.services.AttributeService.GetAttributeValuesByAttributeID(c.Context(), attributeID)
	if err != nil {
		return h.handleError(err, "attribute values")
	}

	return c.JSON(response.OkByData(values))
}

// updateAttributeValue is a function update attribute value
//
//	@Summary		Update attribute value
//	@Description	Update attribute value
//	@Tags			Attribute
//	@Accept			json
//	@Produce		json
//	@Param			id		path		uuid.UUID										true	"Attribute value ID"
//	@Param			update	body		attribute_request.UpdateAttributeValueRequest	true	"Update attribute value"
//	@Success		200		{object}	response.Result[attribute_response.AttributeValueResponse]
//	@Failure		400		{object}	apierror.Errors
//	@Failure		422		{object}	apierror.Errors
//	@Failure		500		{object}	apierror.Errors
//	@Router			/v1/attribute-value/:id [PUT]
//
//	@Security		BearerAuth
func (h *Handler) updateAttributeValue(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}

	req := &attribute_request.UpdateAttributeValueRequest{}
	if err := c.Bind().Body(req); err != nil {
		return err
	}

	d := dto.UpdateAttributeValueDTO{
		Value:           req.Value,
		ValueNormalized: req.ValueNormalized,
		ValueNumeric:    req.ValueNumeric,
		DisplayOrder:    req.DisplayOrder,
		IsActive:        req.IsActive,
	}

	value, err := h.services.AttributeService.UpdateAttributeValue(c.Context(), d, id)
	if err != nil {
		return h.handleError(err, "attribute value")
	}

	return c.JSON(response.OkByData(attribute_response.NewFromAttributeValueModel(value)))
}

// deleteAttributeValue is a function delete attribute value by ID
//
//	@Summary		Delete attribute value by ID
//	@Description	Delete attribute value by ID
//	@Tags			Attribute
//	@Accept			json
//	@Produce		json
//	@Param			id	path		uuid.UUID	true	"Attribute value ID"
//	@Success		200	{object}	response.Result[string]
//	@Failure		400	{object}	apierror.Errors
//	@Failure		500	{object}	apierror.Errors
//	@Router			/v1/attribute-value/:id [DELETE]
//
//	@Security		BearerAuth
func (h *Handler) deleteAttributeValue(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}

	if err := h.services.AttributeService.DeleteAttributeValue(c.Context(), id); err != nil {
		return h.handleError(err, "attribute value")
	}

	return c.JSON(response.OkByMessage("Attribute value successfully deleted"))
}

// getAvailableFilters is a function get available filters for category
//
//	@Summary		Get available filters
//	@Description	Get available filters for category (faceted search)
//	@Tags			Attribute
//	@Accept			json
//	@Produce		json
//	@Param			category_id	query		uuid.UUID	true	"Category ID"
//	@Success		200			{object}	response.Result[[]dto.AttributeGroupWithValuesDTO]
//	@Failure		400			{object}	apierror.Errors
//	@Failure		500			{object}	apierror.Errors
//	@Router			/v1/filters [GET]
func (h *Handler) getAvailableFilters(c fiber.Ctx) error {
	categoryIDStr := c.Query("category_id")
	if categoryIDStr == "" {
		return apierror.New().AddError(fmt.Errorf("category_id is required")).SetHttpCode(fiber.StatusBadRequest)
	}

	categoryID, err := uuid.Parse(categoryIDStr)
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}

	filters, err := h.services.AttributeService.GetAvailableFiltersForCategory(c.Context(), &categoryID)
	if err != nil {
		return h.handleError(err, "filters")
	}

	return c.JSON(response.OkByData(filters))
}

func (h *Handler) initAttributeRoutes(v1 fiber.Router) {
	ag := v1.Group("/attribute-group")
	ag.Get("/", h.getAttributeGroups)
	ag.Get("/find", h.findAttributeGroups)
	ag.Get("/:id", h.getAttributeGroupByID)
	ag.Post("/", h.createAttributeGroup)
	ag.Put("/:id", h.updateAttributeGroup)
	ag.Delete("/:id", h.deleteAttributeGroup)

	a := v1.Group("/attribute")
	a.Get("/", h.getAttributes)
	a.Get("/find", h.findAttribute)
	a.Post("/", h.createAttribute)
	a.Get("/:id", h.getAttributeByID)
	a.Get("/:attribute_id/values", h.getAttributeValues)
	a.Put("/:id", h.updateAttribute)
	a.Delete("/:id", h.deleteAttribute)

	av := v1.Group("/attribute-value")
	av.Get("/", h.getAttributeValue)
	av.Post("/", h.createAttributeValue)
	av.Put("/:id", h.updateAttributeValue)
	av.Delete("/:id", h.deleteAttributeValue)

	// Filters endpoint (public)
	v1.Get("/filters", h.getAvailableFilters)
}
