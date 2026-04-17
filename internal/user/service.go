package user

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/hreftools/api/internal/config"
	"github.com/hreftools/api/internal/emails"
	"golang.org/x/crypto/argon2"
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

// Argon2id configuration parameters (OWASP recommended minimums).
//
// Argon2id is a memory-hard password hashing algorithm — the winner of the
// Password Hashing Competition (2015). "Memory-hard" means it deliberately
// uses a large amount of RAM during hashing, which makes it extremely
// expensive to attack with GPUs or ASICs (which have limited per-core memory).
//
// The "id" variant combines Argon2i (resistant to side-channel attacks) and
// Argon2d (resistant to GPU cracking), making it the recommended choice for
// password hashing.
//
// Unlike bcrypt, Argon2id has no input length limit — bcrypt silently
// truncates passwords at 72 bytes, which means two passwords that share
// the first 72 bytes but differ after that would hash identically.
const (
	// argonMemory is the amount of RAM (in KiB) used during hashing.
	// 64 * 1024 = 65536 KiB = 64 MB. Higher values make brute-force attacks
	// more expensive because each hash attempt must allocate this much memory.
	argonMemory = 64 * 1024

	// argonIterations (also called "time cost") is the number of passes over
	// the memory. More iterations = slower hashing = harder to brute-force.
	// With 64 MB of memory, 1 iteration is sufficient per OWASP guidelines.
	argonIterations = 1

	// argonParallel is the number of threads used during hashing. This should
	// match the number of CPU cores you can dedicate to a single hash operation.
	// Set to 2 as a reasonable default for most server environments.
	argonParallel = 2

	// argonKeyLength is the length of the derived key (hash output) in bytes.
	// 32 bytes = 256 bits, which provides strong collision resistance.
	argonKeyLength = 32

	// argonSaltLength is the length of the random salt in bytes. Each password
	// gets a unique salt, so even identical passwords produce different hashes.
	// 16 bytes = 128 bits, which is more than sufficient to prevent rainbow
	// table attacks and precomputation.
	argonSaltLength = 16
)

// passwordHash generates an Argon2id hash for the given password.
//
// It returns a self-describing string in the PHC (Password Hashing Competition)
// format: $argon2id$v=<version>$m=<memory>,t=<iterations>,p=<parallelism>$<salt>$<key>
//
// Storing the parameters alongside the hash means you can change the cost
// parameters in the future (e.g. increase memory) without breaking existing
// hashes — each hash carries the exact parameters needed to verify it.
func passwordHash(password string) (string, error) {
	// Generate a cryptographically random salt. Each password gets its own
	// unique salt so that two users with the same password will have
	// completely different hashes (prevents rainbow table attacks).
	salt := make([]byte, argonSaltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	// Derive the hash key from the password and salt using Argon2id.
	// IDKey is the Argon2id variant — it takes the password, salt, and all
	// cost parameters, then returns a fixed-length derived key.
	key := argon2.IDKey([]byte(password), salt, argonIterations, argonMemory, argonParallel, argonKeyLength)

	// Encode the hash in PHC string format. Both salt and key are
	// base64-encoded (RawStdEncoding = no padding characters).
	// This format is widely understood by password hashing libraries
	// across languages, making it portable if you ever need to migrate.
	return fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		argonMemory,
		argonIterations,
		argonParallel,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(key),
	), nil
}

// passwordValidate checks whether a plaintext password matches a stored
// Argon2id hash. It parses the parameters, salt, and expected key from
// the stored hash string, then re-derives the key from the candidate
// password using the same parameters. If the derived key matches the
// stored key, the password is correct.
func passwordValidate(password, hash string) bool {
	// Parse the stored hash string to extract the algorithm parameters,
	// salt, and expected key. Sscanf reads structured data from a string
	// using a format template — each %d/%s maps to one of the variables.
	// The salt and key are in the last two $-separated segments, but Sscanf's
	// %s is greedy and captures both (including the $ between them), so we
	// split them manually below.
	var version int
	var memory uint32
	var iterations uint32
	var parallelism uint8
	var saltB64, keyB64 string

	_, err := fmt.Sscanf(hash, "$argon2id$v=%d$m=%d,t=%d,p=%d$%s",
		&version, &memory, &iterations, &parallelism, &saltB64)
	if err != nil {
		return false
	}

	// Sscanf's %s captured "salt$key" as one string because %s reads until
	// whitespace, not until $. Split on the first $ to separate them.
	parts := strings.SplitN(saltB64, "$", 2)
	if len(parts) != 2 {
		return false
	}
	saltB64 = parts[0]
	keyB64 = parts[1]

	// Decode the base64-encoded salt and expected key back into raw bytes.
	salt, err := base64.RawStdEncoding.DecodeString(saltB64)
	if err != nil {
		return false
	}
	expectedKey, err := base64.RawStdEncoding.DecodeString(keyB64)
	if err != nil {
		return false
	}

	// Re-derive the key from the candidate password using the same salt and
	// parameters that were used during hashing. If the password is correct,
	// this will produce the exact same key bytes.
	key := argon2.IDKey([]byte(password), salt, iterations, memory, parallelism, uint32(len(expectedKey)))

	// Use constant-time comparison to prevent timing attacks. A naive
	// byte-by-byte comparison (==) returns early on the first mismatch,
	// which leaks information about how many leading bytes matched. An
	// attacker could exploit this timing difference to guess the hash
	// one byte at a time. ConstantTimeCompare always takes the same
	// amount of time regardless of where (or whether) the bytes differ.
	return subtle.ConstantTimeCompare(key, expectedKey) == 1
}

var (
	// validation username
	ErrValidationUsernameRequired   = errors.New("username is required")
	ErrValidationUsernameTooShort   = errors.New("username must be min 3 characters")
	ErrValidationUsernameTooLong    = errors.New("username must be max 32 characters")
	ErrValidationUsernameCharacters = errors.New("username can only contain lowercase characters, numbers, hyphens, and underscores")
	ErrValidationUsernamePrefix     = errors.New("username cannot start with hyphen or underscore")
	ErrValidationUsernameSuffix     = errors.New("username cannot end with hyphen or underscore")
	ErrValidationUsernameReserved   = errors.New("username is reserved")

	// validation email
	ErrValidationEmailRequired = errors.New("email is required")
	ErrValidationEmailFormat   = errors.New("email format is invalid")
	ErrValidationEmailTooLong  = errors.New("email must be at most 254 characters")

	// validation password
	ErrValidationPasswordRequired = errors.New("password is required")
	ErrValidationPasswordTooShort = errors.New("password must be at least 12 characters")
	ErrValidationPasswordTooLong  = errors.New("password must be at most 128 characters")

	// validation token
	ErrValidationTokenRequired = errors.New("token is required")
	ErrValidationTokenFormat   = errors.New("token is invalid")

	// validation is admin
	ErrValidationIsAdminRequired = errors.New("isAdmin flag is required")

	// validation is pro
	ErrValidationIsProRequired = errors.New("isPro flag is required")

	ErrNotFound                 = errors.New("not found")
	ErrConflict                 = errors.New("conflict")
	ErrInvalidCredentials       = errors.New("invalid email or password")
	ErrEmailNotVerified         = errors.New("invalid email or password")
	ErrTokenExpired             = errors.New("token has expired")
	ErrResendTooFrequent        = errors.New("verification email already sent, please wait before requesting a new one")
	ErrPasswordResetTooFrequent = errors.New("password reset email already sent, please wait before requesting a new one")
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
	username, err := validateUsername(username)
	if err != nil {
		return err
	}
	password, err = validatePassword(password)
	if err != nil {
		return err
	}
	email, err = validateEmail(email)
	if err != nil {
		return err
	}

	passwordHash, err := passwordHash(password)
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
	email, err := validateEmail(email)
	if err != nil {
		return SigninResult{}, err
	}
	password, err = validatePassword(password)
	if err != nil {
		return SigninResult{}, err
	}

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

	if !passwordValidate(password, u.Password) {
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
	tokenStr, err := validateToken(tokenStr)
	if err != nil {
		return err
	}
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
// Returns ErrResendTooFrequent if the last token was generated less than 5 minutes ago.
func (s *Service) ResendVerification(ctx context.Context, email string) error {
	email, err := validateEmail(email)
	if err != nil {
		return err
	}

	u, err := s.Repo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			// silent success for non-existent emails
			return nil
		}
		return err
	}

	if u.EmailVerified {
		// silent success for already verified
		return nil
	}

	tokenAge := config.EmailVerificationTokenExpiryDuration - time.Until(*u.EmailVerificationTokenExpiresAt)
	if tokenAge < time.Minute*5 {
		return ErrResendTooFrequent
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
// Returns ErrPasswordResetTooFrequent if the last token was generated less than 5 minutes ago.
func (s *Service) ResetPasswordRequest(ctx context.Context, email string) error {
	email, err := validateEmail(email)
	if err != nil {
		return err
	}
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
			return ErrPasswordResetTooFrequent
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
	tokenStr, err := validateToken(tokenStr)
	if err != nil {
		return err
	}
	newPassword, err = validatePassword(newPassword)
	if err != nil {
		return err
	}
	token, _ := uuid.Parse(tokenStr)

	u, err := s.Repo.GetByPasswordResetToken(ctx, token)
	if err != nil {
		return err
	}

	if u.PasswordResetTokenExpiresAt != nil && u.PasswordResetTokenExpiresAt.Before(time.Now()) {
		return ErrTokenExpired
	}

	passwordHash, err := passwordHash(newPassword)
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
func (s *Service) AdminCreate(ctx context.Context, username, email, password string, isAdmin, isPro *bool) (User, error) {
	username, err := validateUsername(username)
	if err != nil {
		return User{}, err
	}
	password, err = validatePassword(password)
	if err != nil {
		return User{}, err
	}
	email, err = validateEmail(email)
	if err != nil {
		return User{}, err
	}
	isAdminValue, err := validateIsAdmin(isAdmin)
	if err != nil {
		return User{}, err
	}
	isProValue, err := validateIsPro(isPro)
	if err != nil {
		return User{}, err
	}
	passwordHash, err := passwordHash(password)
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
		IsAdmin:                         isAdminValue,
		IsPro:                           isProValue,
	})
}

// Delete removes a user by ID.
func (s *Service) Delete(ctx context.Context, id uuid.UUID) (User, error) {
	return s.Repo.Delete(ctx, id)
}
