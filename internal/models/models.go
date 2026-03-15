package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/hreftools/api/internal/db"
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

func NewResponseResource(u db.Resource) ResponseResource {
	return ResponseResource{
		ID:          u.ID,
		Title:       u.Title,
		Description: u.Description,
		URL:         u.Url,
		Favourite:   u.Favourite,
		ReadLater:   u.ReadLater,
		CreatedAt:   u.CreatedAt,
		UpdatedAt:   u.UpdatedAt,
	}
}

type ResponseUser struct {
	ID            uuid.UUID `json:"id"`
	Email         string    `json:"email"`
	EmailVerified bool      `json:"emailVerified"`
	Username      string    `json:"username"`
	IsAdmin       bool      `json:"isAdmin"`
	IsPro         bool      `json:"isPro"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

func NewResponseUser(u db.User) ResponseUser {
	return ResponseUser{
		ID:            u.ID,
		Email:         u.Email,
		EmailVerified: u.EmailVerified,
		Username:      u.Username,
		IsAdmin:       u.IsAdmin,
		IsPro:         u.IsPro,
		CreatedAt:     u.CreatedAt,
		UpdatedAt:     u.UpdatedAt,
	}
}
