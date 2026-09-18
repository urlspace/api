package user

import (
	"time"

	"uuid"
)

type PublicUser struct {
	DisplayName string
	Collections []PublicCollection
}

type PublicCollection struct {
	ID          uuid.UUID
	Name        string
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type User struct {
	ID                              uuid.UUID
	Email                           string
	EmailVerified                   bool
	EmailVerificationTokenHash      *string
	EmailVerificationTokenExpiresAt *time.Time
	EmailNew                        *string
	EmailNewCodeHash                *string
	EmailNewCodeHashExpiresAt       *time.Time
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
	UserAgent   *string
	ExpiresAt   time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type Token struct {
	ID          uuid.UUID
	UserID      uuid.UUID
	Description string
	TokenHash   string
	TokenSuffix string
	LastUsedAt  time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
