package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/link-rift/link-rift/internal/repository/sqlc"
)

type User struct {
	ID               uuid.UUID  `json:"id"`
	Email            string     `json:"email"`
	PasswordHash     string     `json:"-"`
	Name             string     `json:"name"`
	AvatarURL        *string    `json:"avatar_url,omitempty"`
	EmailVerifiedAt  *time.Time `json:"email_verified_at,omitempty"`
	TwoFactorEnabled bool       `json:"two_factor_enabled"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

type UserResponse struct {
	ID               uuid.UUID  `json:"id"`
	Email            string     `json:"email"`
	Name             string     `json:"name"`
	AvatarURL        *string    `json:"avatar_url,omitempty"`
	EmailVerifiedAt  *time.Time `json:"email_verified_at,omitempty"`
	TwoFactorEnabled bool       `json:"two_factor_enabled"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

func UserFromSqlc(u sqlc.User) *User {
	user := &User{
		ID:               u.ID,
		Email:            u.Email,
		PasswordHash:     u.PasswordHash,
		Name:             u.Name,
		TwoFactorEnabled: u.TwoFactorEnabled,
	}

	if u.AvatarUrl.Valid {
		user.AvatarURL = &u.AvatarUrl.String
	}
	if u.EmailVerifiedAt.Valid {
		t := u.EmailVerifiedAt.Time
		user.EmailVerifiedAt = &t
	}
	if u.CreatedAt.Valid {
		user.CreatedAt = u.CreatedAt.Time
	}
	if u.UpdatedAt.Valid {
		user.UpdatedAt = u.UpdatedAt.Time
	}

	return user
}

func (u *User) ToResponse() *UserResponse {
	return &UserResponse{
		ID:               u.ID,
		Email:            u.Email,
		Name:             u.Name,
		AvatarURL:        u.AvatarURL,
		EmailVerifiedAt:  u.EmailVerifiedAt,
		TwoFactorEnabled: u.TwoFactorEnabled,
		CreatedAt:        u.CreatedAt,
		UpdatedAt:        u.UpdatedAt,
	}
}
