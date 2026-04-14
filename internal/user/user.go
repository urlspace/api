package user

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID                              uuid.UUID
	Email                           string
	EmailVerified                   bool
	EmailVerificationToken          uuid.NullUUID
	EmailVerificationTokenExpiresAt *time.Time
	Password                        string
	PasswordResetToken              uuid.NullUUID
	PasswordResetTokenExpiresAt     *time.Time
	Username                        string
	IsAdmin                         bool
	IsPro                           bool
	CreatedAt                       time.Time
	UpdatedAt                       time.Time
}

type Token struct {
	ID          uuid.UUID
	UserID      uuid.UUID
	Type        string
	Description *string
	ExpiresAt   time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
