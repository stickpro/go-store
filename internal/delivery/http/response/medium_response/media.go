package medium_response

import (
	"time"

	"github.com/google/uuid"
	"github.com/stickpro/go-store/internal/models"
)

type MediumResponse struct {
	ID        uuid.UUID `db:"id" json:"id"`
	Name      string    `db:"name" json:"name"`
	Path      string    `db:"path" json:"path"`
	FileName  string    `db:"file_name" json:"file_name"`
	MimeType  string    `db:"mime_type" json:"mime_type"`
	DiskType  string    `db:"disk_type" json:"disk_type"`
	Size      int64     `db:"size" json:"size"`
	CreatedAt time.Time `json:"created_at"`
} // @name MediumResponse

func NewFromModel(medium *models.Medium) MediumResponse {
	return MediumResponse{
		ID:        medium.ID,
		Name:      medium.Name,
		Path:      medium.Path,
		FileName:  medium.FileName,
		MimeType:  medium.MimeType,
		DiskType:  medium.DiskType,
		Size:      medium.Size,
		CreatedAt: medium.CreatedAt.Time,
	}
}

func NewFromModels(mediums []*models.Medium) []MediumResponse {
	res := make([]MediumResponse, len(mediums))
	for i, medium := range mediums {
		res[i] = NewFromModel(medium)
	}
	return res
}
