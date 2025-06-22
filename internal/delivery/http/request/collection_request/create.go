package collection_request

type CreateCollectionRequest struct {
	Name        string  `json:"name" validate:"required"`
	Description *string `json:"description,omitempty" validate:"omitempty,max=500"`
	Slug        string  `json:"slug" validate:"required,slug"`
} // @name CreateCollectionRequest

type UpdateCollectionRequest struct {
	Name        string  `json:"name" validate:"required"`
	Description *string `json:"description,omitempty" validate:"omitempty,max=500"`
	Slug        string  `json:"slug" validate:"required,slug"`
} // @name UpdateCollectionRequest
