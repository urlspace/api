package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/hreftools/api/internal/resource"
	"github.com/hreftools/api/internal/user"
)

type ResponseResource struct {
	ID          uuid.UUID `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	URL         string    `json:"url"`
	Favourite   bool      `json:"favourite"`
	ReadLater   bool      `json:"readLater"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

func NewResponseResource(r resource.Resource) ResponseResource {
	return ResponseResource{
		ID:          r.ID,
		Title:       r.Title,
		Description: r.Description,
		URL:         r.Url,
		Favourite:   r.Favourite,
		ReadLater:   r.ReadLater,
		CreatedAt:   r.CreatedAt,
		UpdatedAt:   r.UpdatedAt,
	}
}

type ResponseUser struct {
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	Username  string    `json:"username"`
	IsAdmin   bool      `json:"isAdmin"`
	IsPro     bool      `json:"isPro"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func NewResponseUser(u user.User) ResponseUser {
	return ResponseUser{
		ID:        u.ID,
		Email:     u.Email,
		Username:  u.Username,
		IsAdmin:   u.IsAdmin,
		IsPro:     u.IsPro,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}

type ResponseUserAdmin struct {
	ID                              uuid.UUID     `json:"id"`
	Email                           string        `json:"email"`
	EmailVerified                   bool          `json:"emailVerified"`
	EmailVerificationToken          uuid.NullUUID `json:"emailVerifcationToken"`
	EmailVerificationTokenExpiresAt *time.Time    `json:"emailVerificationTokenExpiresAt"`
	PasswordResetToken              uuid.NullUUID `json:"passwordResetToken"`
	PasswordResetTokenExpiresAt     *time.Time    `json:"passwordResetTokenExpiresAt"`
	Username                        string        `json:"username"`
	IsAdmin                         bool          `json:"isAdmin"`
	IsPro                           bool          `json:"isPro"`
	CreatedAt                       time.Time     `json:"createdAt"`
	UpdatedAt                       time.Time     `json:"updatedAt"`
}

func NewResponseUserAdmin(u user.User) ResponseUserAdmin {
	return ResponseUserAdmin{
		ID:                              u.ID,
		Email:                           u.Email,
		EmailVerified:                   u.EmailVerified,
		EmailVerificationToken:          u.EmailVerificationToken,
		EmailVerificationTokenExpiresAt: u.EmailVerificationTokenExpiresAt,
		PasswordResetToken:              u.PasswordResetToken,
		PasswordResetTokenExpiresAt:     u.PasswordResetTokenExpiresAt,
		Username:                        u.Username,
		IsAdmin:                         u.IsAdmin,
		IsPro:                           u.IsPro,
		CreatedAt:                       u.CreatedAt,
		UpdatedAt:                       u.UpdatedAt,
	}
}
