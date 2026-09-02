package dto

import (
	"fmt"
	"path"
	"strings"

	"github.com/google/uuid"
	"github.com/stickpro/go-store/internal/config"
	"github.com/stickpro/go-store/internal/models"
)

// ImageDTO is the front-end image contract. It lives in models to avoid an import
// cycle; this alias keeps the dto.ImageDTO name for existing call sites.
type ImageDTO = models.ImageDTO

// NewImageDTO builds the contract object for a media file. storedPath is media.path
// (e.g. "storage/public/images/xxx.jpg"); preset URLs follow "<base>_<size>.<ext>".
func NewImageDTO(id uuid.UUID, storedPath string, width, height int32, alt string, presets []config.ImagePreset) ImageDTO {
	base := strings.TrimSuffix(path.Base(storedPath), path.Ext(storedPath))
	out := make(map[string]map[string]string, len(presets))
	for _, p := range presets {
		if p.Name == "" || p.Size <= 0 {
			continue
		}
		out[p.Name] = map[string]string{
			"webp": fmt.Sprintf("/storage/public/images/%s_%d.webp", base, p.Size),
			"jpeg": fmt.Sprintf("/storage/public/images/%s_%d.jpg", base, p.Size),
		}
	}
	return ImageDTO{ID: id, Alt: alt, Width: width, Height: height, Presets: out}
}
