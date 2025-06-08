package attribute_response

import (
	"github.com/google/uuid"
	"github.com/stickpro/go-store/internal/models"
	"github.com/stickpro/go-store/pkg/dbutils/pgtypeutils"
	"time"
)

type AttributeGroupResponse struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description *string   `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
} // @name AttributeGroupResponse

func NewFromModel(aGroup *models.AttributeGroup) AttributeGroupResponse {
	return AttributeGroupResponse{
		ID:          aGroup.ID,
		Name:        aGroup.Name,
		Description: pgtypeutils.DecodeText(aGroup.Description),
		CreatedAt:   aGroup.CreatedAt.Time,
		UpdatedAt:   aGroup.UpdatedAt.Time,
	}
}
