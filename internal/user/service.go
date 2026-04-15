package user

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/hreftools/api/internal/config"
	"github.com/hreftools/api/internal/emails"
	"github.com/hreftools/api/internal/utils"
)

type CreateParams struct {
	Email                           string
	EmailVerified                   bool
	EmailVerificationToken          uuid.NullUUID
	EmailVerificationTokenExpiresAt *time.Time
	Password                        string
	Username                        string
	IsAdmin                         bool
	IsPro                           bool
}

type UpdateVerificationTokenParams struct {
	ID                              uuid.UUID
	EmailVerificationToken          uuid.NullUUID
	EmailVerificationTokenExpiresAt *time.Time
}

type UpdatePasswordResetTokenParams struct {
	ID                          uuid.UUID
	PasswordResetToken          uuid.NullUUID
	PasswordResetTokenExpiresAt *time.Time
}

type Repository interface {
	List(ctx context.Context) ([]User, error)
	GetById(ctx context.Context, id uuid.UUID) (User, error)
	GetByEmail(ctx context.Context, email string) (User, error)
	GetByEmailVerificationToken(ctx context.Context, token uuid.UUID) (User, error)
	GetByPasswordResetToken(ctx context.Context, token uuid.UUID) (User, error)
	Create(ctx context.Context, params CreateParams) (User, error)
	Verify(ctx context.Context, id uuid.UUID) (User, error)
	UpdateVerificationToken(ctx context.Context, params UpdateVerificationTokenParams) (User, error)
	UpdatePasswordResetToken(ctx context.Context, params UpdatePasswordResetTokenParams) (User, error)
	ResetPassword(ctx context.Context, id uuid.UUID, passwordHash string) (User, error)
	Delete(ctx context.Context, id uuid.UUID) (User, error)
}

type TokenCreateParams struct {
	UserID      uuid.UUID
	Type        string
	Description *string
	ExpiresAt   time.Time
}

type TokenUpdateExpiresAtParams struct {
	ID        uuid.UUID
	ExpiresAt time.Time
}

type TokenRepository interface {
	Create(ctx context.Context, params TokenCreateParams) (Token, error)
	GetByID(ctx context.Context, id uuid.UUID) (Token, error)
	UpdateExpiresAt(ctx context.Context, params TokenUpdateExpiresAtParams) (Token, error)
	Delete(ctx context.Context, id uuid.UUID) error
	DeleteAllByUserID(ctx context.Context, userID uuid.UUID) error
}

var (
	ErrNotFound           = errors.New("not found")
	ErrConflict           = errors.New("conflict")
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrEmailNotVerified   = errors.New("invalid email or password")
	ErrTokenExpired       = errors.New("token has expired")
	ErrRateLimited        = errors.New("rate limited")
)

type Service struct {
	Repo        Repository
	TokenRepo   TokenRepository
	EmailSender emails.EmailSender
}

func NewService(repo Repository, tokenRepo TokenRepository, emailSender emails.EmailSender) *Service {
	return &Service{
		Repo:        repo,
		TokenRepo:   tokenRepo,
		EmailSender: emailSender,
	}
}

// Signup creates a new unverified user and sends a verification email.
func (s *Service) Signup(ctx context.Context, username, email, password string) error {
	passwordHash, err := utils.PasswordHash(password)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	token := uuid.NullUUID{Valid: true, UUID: uuid.New()}
	expiresAt := time.Now().Add(config.EmailVerificationTokenExpiryDuration)

	_, err = s.Repo.Create(ctx, CreateParams{
		Email:                           email,
		EmailVerified:                   false,
		EmailVerificationToken:          token,
		EmailVerificationTokenExpiresAt: &expiresAt,
		Password:                        passwordHash,
		Username:                        username,
		IsAdmin:                         false,
		IsPro:                           false,
	})
	if err != nil {
		return err
	}

	emailVerifyData := emails.AuthSignupParams{
		Username: username,
		Email:    email,
		Token:    token.UUID.String(),
	}
	bodyHtml, err := emails.RenderTemplateHtml(emails.AuthSignupTemplateHtml, emailVerifyData)
	if err != nil {
		return fmt.Errorf("failed to render html email template: %w", err)
	}
	bodyText, err := emails.RenderTemplateTxt(emails.AuthSignupTemplateTxt, emailVerifyData)
	if err != nil {
		return fmt.Errorf("failed to render text email template: %w", err)
	}

	err = s.EmailSender.Send(emails.EmailSendParams{
		To:      []string{email},
		Text:    bodyText,
		Html:    bodyHtml,
		Subject: "Hello from href.tools",
	})
	if err != nil {
		log.Printf("Failed to send email: %v", err)
	}

	return nil
}

type SigninResult struct {
	Token Token
}

// Signin validates credentials and creates a session token.
func (s *Service) Signin(ctx context.Context, email, password string, description *string) (SigninResult, error) {
	u, err := s.Repo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return SigninResult{}, ErrInvalidCredentials
		}
		return SigninResult{}, err
	}

	if !u.EmailVerified {
		return SigninResult{}, ErrEmailNotVerified
	}

	if !utils.PasswordValidate(password, u.Password) {
		return SigninResult{}, ErrInvalidCredentials
	}

	token, err := s.TokenRepo.Create(ctx, TokenCreateParams{
		UserID:      u.ID,
		Type:        config.TokenTypeSession,
		Description: description,
		ExpiresAt:   time.Now().Add(config.SessionExpiryDuration),
	})
	if err != nil {
		return SigninResult{}, err
	}

	return SigninResult{Token: token}, nil
}

// Verify validates a verification token and marks the user as verified.
func (s *Service) Verify(ctx context.Context, tokenStr string) error {
	token, _ := uuid.Parse(tokenStr)

	u, err := s.Repo.GetByEmailVerificationToken(ctx, token)
	if err != nil {
		return err
	}

	if u.EmailVerificationTokenExpiresAt != nil && u.EmailVerificationTokenExpiresAt.Before(time.Now()) {
		return ErrTokenExpired
	}

	_, err = s.Repo.Verify(ctx, u.ID)
	return err
}

// ResendVerification generates a new verification token and sends an email.
// Returns ErrRateLimited if the last token was generated less than 5 minutes ago.
func (s *Service) ResendVerification(ctx context.Context, email string) error {
	u, err := s.Repo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil // silent success for non-existent emails
		}
		return err
	}

	if u.EmailVerified {
		return nil // silent success for already verified
	}

	tokenAge := config.EmailVerificationTokenExpiryDuration - time.Until(*u.EmailVerificationTokenExpiresAt)
	if tokenAge < time.Minute*5 {
		return ErrRateLimited
	}

	token := uuid.NullUUID{Valid: true, UUID: uuid.New()}
	expiresAt := time.Now().Add(config.EmailVerificationTokenExpiryDuration)

	_, err = s.Repo.UpdateVerificationToken(ctx, UpdateVerificationTokenParams{
		ID:                              u.ID,
		EmailVerificationToken:          token,
		EmailVerificationTokenExpiresAt: &expiresAt,
	})
	if err != nil {
		return err
	}

	templateParams := emails.AuthResendVerificationParams{
		Token: token.UUID.String(),
	}
	bodyHtml, err := emails.RenderTemplateHtml(emails.AuthResendVerificationTemplateHtml, templateParams)
	if err != nil {
		return fmt.Errorf("failed to render html email template: %w", err)
	}
	bodyText, err := emails.RenderTemplateTxt(emails.AuthResendVerificationTemplateTxt, templateParams)
	if err != nil {
		return fmt.Errorf("failed to render text email template: %w", err)
	}

	err = s.EmailSender.Send(emails.EmailSendParams{
		To:      []string{email},
		Text:    bodyText,
		Html:    bodyHtml,
		Subject: "Verification token has been requested",
	})
	if err != nil {
		log.Printf("Failed to send email: %v", err)
	}

	return nil
}

// ResetPasswordRequest generates a password reset token and sends an email.
// Returns ErrRateLimited if the last token was generated less than 5 minutes ago.
func (s *Service) ResetPasswordRequest(ctx context.Context, email string) error {
	u, err := s.Repo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil // silent success
		}
		return err
	}

	if u.PasswordResetTokenExpiresAt != nil {
		tokenAge := config.PasswordResetTokenExpiryDuration - time.Until(*u.PasswordResetTokenExpiresAt)
		if tokenAge < time.Minute*5 {
			return ErrRateLimited
		}
	}

	token := uuid.NullUUID{Valid: true, UUID: uuid.New()}
	expiresAt := time.Now().Add(config.PasswordResetTokenExpiryDuration)

	_, err = s.Repo.UpdatePasswordResetToken(ctx, UpdatePasswordResetTokenParams{
		ID:                          u.ID,
		PasswordResetToken:          token,
		PasswordResetTokenExpiresAt: &expiresAt,
	})
	if err != nil {
		return err
	}

	templateParams := emails.AuthResetPasswordRequestParams{
		Token: token.UUID.String(),
	}
	bodyHtml, err := emails.RenderTemplateHtml(emails.AuthResetPasswordRequestTemplateHtml, templateParams)
	if err != nil {
		return fmt.Errorf("failed to render html email template: %w", err)
	}
	bodyText, err := emails.RenderTemplateTxt(emails.AuthResetPasswordRequestTemplateTxt, templateParams)
	if err != nil {
		return fmt.Errorf("failed to render text email template: %w", err)
	}

	err = s.EmailSender.Send(emails.EmailSendParams{
		To:      []string{email},
		Text:    bodyText,
		Html:    bodyHtml,
		Subject: "Password reset has been requested",
	})
	if err != nil {
		log.Printf("Failed to send email: %v", err)
	}

	return nil
}

// ResetPasswordConfirm validates a reset token and sets the new password.
func (s *Service) ResetPasswordConfirm(ctx context.Context, tokenStr, newPassword string) error {
	token, _ := uuid.Parse(tokenStr)

	u, err := s.Repo.GetByPasswordResetToken(ctx, token)
	if err != nil {
		return err
	}

	if u.PasswordResetTokenExpiresAt != nil && u.PasswordResetTokenExpiresAt.Before(time.Now()) {
		return ErrTokenExpired
	}

	passwordHash, err := utils.PasswordHash(newPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	_, err = s.Repo.ResetPassword(ctx, u.ID, passwordHash)
	if err != nil {
		return err
	}

	go s.TokenRepo.DeleteAllByUserID(context.Background(), u.ID)

	return nil
}

// Signout deletes a session token.
func (s *Service) Signout(ctx context.Context, tokenID uuid.UUID) error {
	return s.TokenRepo.Delete(ctx, tokenID)
}

// GetTokenByID retrieves a token by its ID (used by auth middleware).
func (s *Service) GetTokenByID(ctx context.Context, id uuid.UUID) (Token, error) {
	return s.TokenRepo.GetByID(ctx, id)
}

// UpdateTokenExpiresAt updates the expiry of a token (used by auth middleware for sliding sessions).
func (s *Service) UpdateTokenExpiresAt(ctx context.Context, params TokenUpdateExpiresAtParams) (Token, error) {
	return s.TokenRepo.UpdateExpiresAt(ctx, params)
}

// GetById retrieves a user by ID.
func (s *Service) GetById(ctx context.Context, id uuid.UUID) (User, error) {
	return s.Repo.GetById(ctx, id)
}

// List returns all users.
func (s *Service) List(ctx context.Context) ([]User, error) {
	return s.Repo.List(ctx)
}

// AdminCreate creates a new user (admin operation — pre-verified, no email sent).
func (s *Service) AdminCreate(ctx context.Context, username, email, password string, isAdmin, isPro bool) (User, error) {
	passwordHash, err := utils.PasswordHash(password)
	if err != nil {
		return User{}, fmt.Errorf("failed to hash password: %w", err)
	}

	return s.Repo.Create(ctx, CreateParams{
		Email:                           email,
		EmailVerified:                   true,
		EmailVerificationToken:          uuid.NullUUID{},
		EmailVerificationTokenExpiresAt: nil,
		Password:                        passwordHash,
		Username:                        username,
		IsAdmin:                         isAdmin,
		IsPro:                           isPro,
	})
}

// Delete removes a user by ID.
func (s *Service) Delete(ctx context.Context, id uuid.UUID) (User, error) {
	return s.Repo.Delete(ctx, id)
}
