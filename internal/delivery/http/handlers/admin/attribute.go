package admin

import (
	"errors"
	"github.com/gofiber/fiber/v3"
	"github.com/stickpro/go-store/internal/delivery/http/request/attribute_request"
	"github.com/stickpro/go-store/internal/delivery/http/response"
	"github.com/stickpro/go-store/internal/delivery/http/response/attribute_response"
	"github.com/stickpro/go-store/internal/service/attribute"
	"github.com/stickpro/go-store/internal/tools/apierror"
	"github.com/stickpro/go-store/pkg/dbutils/pgerror"
)

// updateAtt0 is a function create attribute group
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
func (h Handler) createAtributeGroup(c fiber.Ctx) error {
	req := &attribute_request.CreateAttributeGroupRequest{}
	if err := c.Bind().Body(req); err != nil {
		return err
	}

	dto := attribute.RequestToCreateGroupDTO(req)
	aGroup, err := h.services.AttributeService.CreateAttributeGroup(c.Context(), dto)
	if err != nil {
		var uniqueErr *pgerror.UniqueConstraintError
		if errors.As(err, &uniqueErr) {
			return apierror.New().AddError(uniqueErr).SetHttpCode(fiber.StatusUnprocessableEntity)
		}
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusInternalServerError)
	}
	return c.JSON(response.OkByData(attribute_response.NewFromModel(aGroup)))
}
