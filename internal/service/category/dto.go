package category

import "github.com/google/uuid"

type CreateDTO struct {
	Name            string        `json:"name"`
	ParentID        uuid.NullUUID `json:"parent_id"`
	Slug            string        `json:"slug"`
	Description     *string       `json:"description"`
	MetaTitle       *string       `json:"meta_title"`
	MetaH1          *string       `json:"meta_h1"`
	MetaDescription *string       `json:"meta_description"`
	MetaKeyword     *string       `json:"meta_keyword"`
	IsEnable        bool          `json:"is_enable"`
}
