package collection_response

import (
	"github.com/google/uuid"
	"github.com/stickpro/go-store/internal/models"
	"github.com/stickpro/go-store/pkg/dbutils/pgtypeutils"
	"time"
)

type CollectionResponse struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description *string   `json:"description,omitempty"`
	Slug        string    `json:"slug"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
} // @name CollectionResponse

func NewFromModel(collection *models.Collection) *CollectionResponse {
	return &CollectionResponse{
		ID:          collection.ID,
		Name:        collection.Name,
		Description: pgtypeutils.DecodeText(collection.Description),
		Slug:        collection.Slug,
		CreatedAt:   collection.CreatedAt.Time,
		UpdatedAt:   collection.UpdatedAt.Time,
	}
}

func NewFromModels(collection []*models.Collection) []*CollectionResponse {
	res := make([]*CollectionResponse, 0, len(collection))
	for _, c := range collection {
		res = append(res, NewFromModel(c))
	}
	return res
}
