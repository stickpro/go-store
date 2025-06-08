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

// updateAtt0 is a function create category
//
//	@Summary		Create Category
//	@Description	Create category
//	@Tags			Category
//	@Accept			json
//	@Produce		json
//	@Param			create	body		category_request.CreateCategoryRequest	true	"Create category"
//	@Success		200		{object}	response.Result[category_response.CategoryResponse]
//	@Failure		400		{object}	apierror.Errors
//	@Failure		422		{object}	apierror.Errors
//	@Failure		500		{object}	apierror.Errors
//	@Router			/v1/category/ [POST]
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
