package tasks

import "github.com/google/uuid"

const ImageSyncQueue = "image_sync"

type ImageSyncTask struct {
	ProductID uuid.UUID `json:"product_id"`
	ImageMain *string   `json:"image_main,omitempty"`
	Images    []string  `json:"images,omitempty"`
}
