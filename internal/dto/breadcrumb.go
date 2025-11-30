package dto

import "github.com/google/uuid"

type BreadcrumbDTO struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	MetaTitle *string   `json:"meta_title,omitempty"`
	MetaH1    *string   `json:"meta_h1,omitempty"`
	Depth     int32     `json:"depth"`
} // @name BreadcrumbDTO

type CategoryChildDTO struct {
	ID       uuid.UUID     `json:"id"`
	ParentID uuid.NullUUID `json:"parent_id"`
	Name     string        `json:"name"`
	Slug     string        `json:"slug"`
	IsEnable bool          `json:"is_enable"`
} // @name CategoryChildDTO
