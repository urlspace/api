package server

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/hreftools/api/internal/resource"
	"github.com/hreftools/api/internal/user"
)

type responseResource struct {
	ID          uuid.UUID `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	URL         string    `json:"url"`
	Favourite   bool      `json:"favourite"`
	ReadLater   bool      `json:"readLater"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

func newResponseResource(r resource.Resource) responseResource {
	return responseResource{
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

type responseUser struct {
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	Username  string    `json:"username"`
	IsAdmin   bool      `json:"isAdmin"`
	IsPro     bool      `json:"isPro"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func newResponseUser(u user.User) responseUser {
	return responseUser{
		ID:        u.ID,
		Email:     u.Email,
		Username:  u.Username,
		IsAdmin:   u.IsAdmin,
		IsPro:     u.IsPro,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}

type responseUserAdmin struct {
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

func newResponseUserAdmin(u user.User) responseUserAdmin {
	return responseUserAdmin{
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

// JSON response helpers

type errorResponse struct {
	Status string `json:"status"`
	Data   string `json:"data"`
}

func writeJSONSuccess(w http.ResponseWriter, statusCode int, res any) {
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(res); err != nil {
		log.Printf("Error encoding response: %v", err)
	}
}

func writeJSONError(w http.ResponseWriter, statusCode int, message string) {
	w.WriteHeader(statusCode)

	res := &errorResponse{
		Status: "error",
		Data:   message,
	}

	if err := json.NewEncoder(w).Encode(res); err != nil {
		log.Printf("Error encoding error response: %v", err)
	}
}

func handleDbError(w http.ResponseWriter, err error) {
	if errors.Is(err, user.ErrNotFound) || errors.Is(err, resource.ErrNotFound) {
		writeJSONError(w, http.StatusNotFound, "entry not found")
		return
	}

	if errors.Is(err, user.ErrConflict) || errors.Is(err, resource.ErrConflict) {
		writeJSONError(w, http.StatusConflict, "request conflict")
		return
	}

	if errors.Is(err, context.DeadlineExceeded) {
		writeJSONError(w, http.StatusRequestTimeout, "request timeout")
		return
	}

	if errors.Is(err, context.Canceled) {
		writeJSONError(w, 499, "request cancelled")
		return
	}

	log.Printf("Database error: %v", err)
	writeJSONError(w, http.StatusInternalServerError, "internal server error")
}

func handleClientError(w http.ResponseWriter, err error, message string) {
	log.Printf("Client error: %v", err)
	writeJSONError(w, http.StatusBadRequest, message)
}

func handleServerError(w http.ResponseWriter, err error, message string) {
	log.Printf("Server error: %v", err)
	writeJSONError(w, http.StatusInternalServerError, "internal server error")
}
