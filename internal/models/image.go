package models

import "github.com/google/uuid"

// ImageDTO is the front-end image contract: original id/alt/dimensions plus a
// map of preset -> format -> URL. Kept in models so lightweight product models
// (ShortProduct, ...) can embed it without an import cycle.
type ImageDTO struct {
	ID      uuid.UUID                    `json:"id"`
	Alt     string                       `json:"alt"`
	Width   int32                        `json:"width"`
	Height  int32                        `json:"height"`
	Presets map[string]map[string]string `json:"presets"`
} //	@name	ImageDTO
