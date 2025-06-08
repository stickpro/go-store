package attribute_request

type CreateAttributeGroupRequest struct {
	Name        string  `json:"name" validate:"required"`
	Description *string `json:"description" validate:"omitempty,min=1,max=100"`
}
