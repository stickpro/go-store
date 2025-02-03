package user_response

import (
	"github.com/google/uuid"
	"github.com/stickpro/go-store/internal/models"
	"time"
)

type UserInfoResponse struct {
	ID              uuid.UUID  `json:"id" format:"uuid"`
	Email           string     `json:"email" validate:"required,email" format:"email"`
	EmailVerifiedAt *time.Time `json:"email_verified_at" format:"date-time"`
	Location        string     `json:"location" validate:"required,timezone"`
	Language        string     `json:"language"`
	IsAdmin         bool       `json:"is_admin"`
	CreatedAt       time.Time  `json:"created_at" format:"date-time"`
	UpdatedAt       *time.Time `json:"updated_at" format:"date-time"`
} // @name UserInfoResponse

func NewFromModel(user *models.User) UserInfoResponse {
	var emailVerifiedAt *time.Time
	if user.EmailVerifiedAt.Valid {
		emailVerifiedAt = &user.EmailVerifiedAt.Time
	}

	return UserInfoResponse{
		ID:              user.ID,
		Email:           user.Email,
		EmailVerifiedAt: emailVerifiedAt,
		Location:        user.Location,
		Language:        user.Language,
		IsAdmin:         user.IsAdmin.Bool,
		CreatedAt:       user.CreatedAt.Time,
		UpdatedAt:       &user.UpdatedAt.Time,
	}
}
