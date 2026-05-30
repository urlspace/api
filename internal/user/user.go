package user

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID                              uuid.UUID
	Email                           string
	EmailVerified                   bool
	EmailVerificationTokenHash      *string
	EmailVerificationTokenExpiresAt *time.Time
	Password                        string
	PasswordResetTokenHash          *string
	PasswordResetTokenExpiresAt     *time.Time
	Username                        string
	DisplayName                     string
	IsAdmin                         bool
	IsPro                           bool
	CreatedAt                       time.Time
	UpdatedAt                       time.Time
}

type Session struct {
	ID          uuid.UUID
	UserID      uuid.UUID
	SessionHash string
	Description *string
	ExpiresAt   time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type Token struct {
	ID          uuid.UUID
	UserID      uuid.UUID
	Description string
	TokenHash   string
	LastUsedAt  time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
