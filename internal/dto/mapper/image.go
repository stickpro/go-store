package mapper

import (
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stickpro/go-store/internal/config"
	"github.com/stickpro/go-store/internal/dto"
	"github.com/stickpro/go-store/internal/models"
)

// shortImage builds the ImageDTO from nullable media columns, or nil when absent.
func shortImage(id uuid.NullUUID, imgPath pgtype.Text, w, h pgtype.Int4, alt string, presets []config.ImagePreset) *models.ImageDTO {
	if !id.Valid || !imgPath.Valid {
		return nil
	}
	img := dto.NewImageDTO(id.UUID, imgPath.String, w.Int32, h.Int32, alt, presets)
	return &img
}
