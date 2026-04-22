package user

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/urlspace/api/internal/config"
	"github.com/urlspace/api/internal/emails"
	"golang.org/x/crypto/argon2"
)

type CreateParams struct {
	Email                           string
	EmailVerified                   bool
	EmailVerificationToken          uuid.NullUUID
	EmailVerificationTokenExpiresAt *time.Time
	Password                        string
	Username                        string
	DisplayName                     string
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

type SessionCreateParams struct {
	UserID      uuid.UUID
	Description *string
	ExpiresAt   time.Time
}

type SessionUpdateExpiresAtParams struct {
	ID        uuid.UUID
	ExpiresAt time.Time
}

type SessionRepository interface {
	Create(ctx context.Context, params SessionCreateParams) (Session, error)
	GetByID(ctx context.Context, id uuid.UUID) (Session, error)
	UpdateExpiresAt(ctx context.Context, params SessionUpdateExpiresAtParams) (Session, error)
	Delete(ctx context.Context, id uuid.UUID) error
	DeleteAllByUserID(ctx context.Context, userID uuid.UUID) error
}

type TokenCreateParams struct {
	UserID      uuid.UUID
	Description string
	Hash        string
}

type TokenRepository interface {
	Create(ctx context.Context, params TokenCreateParams) (Token, error)
	GetByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) (Token, error)
	GetByHash(ctx context.Context, hash string) (Token, error)
	List(ctx context.Context, userID uuid.UUID) ([]Token, error)
	UpdateLastUsedAt(ctx context.Context, id uuid.UUID) error
	Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID) error
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

// passwordMatchesHash checks whether a plaintext password matches a stored
// Argon2id hash. It parses the parameters, salt, and expected key from
// the stored hash string, then re-derives the key from the candidate
// password using the same parameters. If the derived key matches the
// stored key, the password is correct.
func passwordMatchesHash(password, hash string) bool {
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

	// validation display name
	ErrValidationDisplayNameRequired         = errors.New("display name is required")
	ErrValidationDisplayNameTooShort         = errors.New("display name must be min 3 characters")
	ErrValidationDisplayNameTooLong          = errors.New("display name must be max 32 characters")
	ErrValidationDisplayNameCharacters       = errors.New("display name can only contain letters, numbers, spaces, hyphens, and underscores")
	ErrValidationDisplayNameConsecutiveSpaces = errors.New("display name cannot contain consecutive spaces")

	// validation token
	ErrValidationTokenRequired = errors.New("token is required")
	ErrValidationTokenFormat   = errors.New("token is invalid")

	// validation is admin
	ErrValidationIsAdminRequired = errors.New("isAdmin flag is required")

	// validation is pro
	ErrValidationIsProRequired = errors.New("isPro flag is required")

	// validation api token description
	ErrValidationTokenDescriptionRequired = errors.New("token description is required")
	ErrValidationTokenDescriptionTooLong  = errors.New("token description must be at most 255 characters")

	ErrNotFound                 = errors.New("not found")
	ErrConflict                 = errors.New("conflict")
	ErrInvalidCredentials       = errors.New("invalid credentials")
	ErrEmailNotVerified         = errors.New("invalid email or password")
	ErrTokenExpired             = errors.New("token has expired")
	ErrResendTooFrequent        = errors.New("verification email already sent, please wait before requesting a new one")
	ErrPasswordResetTooFrequent = errors.New("password reset email already sent, please wait before requesting a new one")
)

type Service struct {
	UserRepo    Repository
	SessionRepo SessionRepository
	TokenRepo   TokenRepository
	EmailSender emails.EmailSender
}

func NewService(userRepo Repository, sessionRepo SessionRepository, tokenRepo TokenRepository, emailSender emails.EmailSender) *Service {
	return &Service{
		UserRepo:    userRepo,
		SessionRepo: sessionRepo,
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

	_, err = s.UserRepo.Create(ctx, CreateParams{
		Email:                           email,
		EmailVerified:                   false,
		EmailVerificationToken:          token,
		EmailVerificationTokenExpiresAt: &expiresAt,
		Password:                        passwordHash,
		Username:                        username,
		DisplayName:                     username,
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
		Subject: "Hello from url.space",
	})
	if err != nil {
		log.Printf("Failed to send email: %v", err)
	}

	return nil
}

// dummyHash is a pre-computed Argon2id hash used to prevent timing-based user
// enumeration. When a sign-in attempt uses an unknown email, we still run the
// hash computation against this dummy so the response time is indistinguishable
// from a wrong-password attempt.
var dummyHash = func() string {
	h, err := passwordHash("dummy-password-never-matches")
	if err != nil {
		panic("failed to generate dummy hash: " + err.Error())
	}
	return h
}()

type SigninResult struct {
	Session Session
}

// Signin validates credentials and creates a session.
func (s *Service) Signin(ctx context.Context, email, password string, description *string) (SigninResult, error) {
	email, err := validateEmail(email)
	if err != nil {
		return SigninResult{}, err
	}
	password, err = validatePassword(password)
	if err != nil {
		return SigninResult{}, err
	}

	u, err := s.UserRepo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			passwordMatchesHash(password, dummyHash) // prevent timing-based user enumeration
			return SigninResult{}, ErrInvalidCredentials
		}
		return SigninResult{}, err
	}

	if !u.EmailVerified {
		return SigninResult{}, ErrEmailNotVerified
	}

	if !passwordMatchesHash(password, u.Password) {
		return SigninResult{}, ErrInvalidCredentials
	}

	session, err := s.SessionRepo.Create(ctx, SessionCreateParams{
		UserID:      u.ID,
		Description: description,
		ExpiresAt:   time.Now().Add(config.SessionExpiryDuration),
	})
	if err != nil {
		return SigninResult{}, err
	}

	return SigninResult{Session: session}, nil
}

// Verify validates a verification token and marks the user as verified.
func (s *Service) Verify(ctx context.Context, tokenStr string) error {
	tokenStr, err := validateToken(tokenStr)
	if err != nil {
		return err
	}
	token, _ := uuid.Parse(tokenStr)

	u, err := s.UserRepo.GetByEmailVerificationToken(ctx, token)
	if err != nil {
		return err
	}

	if u.EmailVerificationTokenExpiresAt != nil && u.EmailVerificationTokenExpiresAt.Before(time.Now()) {
		return ErrTokenExpired
	}

	_, err = s.UserRepo.Verify(ctx, u.ID)
	return err
}

// ResendVerification generates a new verification token and sends an email.
// Returns ErrResendTooFrequent if the last token was generated less than 5 minutes ago.
func (s *Service) ResendVerification(ctx context.Context, email string) error {
	email, err := validateEmail(email)
	if err != nil {
		return err
	}

	u, err := s.UserRepo.GetByEmail(ctx, email)
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

	_, err = s.UserRepo.UpdateVerificationToken(ctx, UpdateVerificationTokenParams{
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
	u, err := s.UserRepo.GetByEmail(ctx, email)
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

	_, err = s.UserRepo.UpdatePasswordResetToken(ctx, UpdatePasswordResetTokenParams{
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

	u, err := s.UserRepo.GetByPasswordResetToken(ctx, token)
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

	_, err = s.UserRepo.ResetPassword(ctx, u.ID, passwordHash)
	if err != nil {
		return err
	}

	go s.SessionRepo.DeleteAllByUserID(context.Background(), u.ID)

	return nil
}

// Signout deletes a session.
func (s *Service) Signout(ctx context.Context, sessionID uuid.UUID) error {
	return s.SessionRepo.Delete(ctx, sessionID)
}

// GetSessionByID retrieves a session by its ID (used by auth middleware).
func (s *Service) GetSessionByID(ctx context.Context, id uuid.UUID) (Session, error) {
	return s.SessionRepo.GetByID(ctx, id)
}

// UpdateSessionExpiresAt updates the expiry of a session (used by auth middleware for sliding expiry).
func (s *Service) UpdateSessionExpiresAt(ctx context.Context, params SessionUpdateExpiresAtParams) (Session, error) {
	return s.SessionRepo.UpdateExpiresAt(ctx, params)
}

// GetById retrieves a user by ID.
func (s *Service) GetById(ctx context.Context, id uuid.UUID) (User, error) {
	return s.UserRepo.GetById(ctx, id)
}

// List returns all users.
func (s *Service) List(ctx context.Context) ([]User, error) {
	return s.UserRepo.List(ctx)
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

	return s.UserRepo.Create(ctx, CreateParams{
		Email:                           email,
		EmailVerified:                   true,
		EmailVerificationToken:          uuid.NullUUID{},
		EmailVerificationTokenExpiresAt: nil,
		Password:                        passwordHash,
		Username:                        username,
		DisplayName:                     username,
		IsAdmin:                         isAdminValue,
		IsPro:                           isProValue,
	})
}

// Delete removes a user by ID (admin operation).
func (s *Service) Delete(ctx context.Context, id uuid.UUID) (User, error) {
	return s.UserRepo.Delete(ctx, id)
}

// DeleteSelf removes the authenticated user after verifying their password.
func (s *Service) DeleteSelf(ctx context.Context, userID uuid.UUID, password string) error {
	password, err := validatePassword(password)
	if err != nil {
		return err
	}

	u, err := s.UserRepo.GetById(ctx, userID)
	if err != nil {
		return err
	}

	if !passwordMatchesHash(password, u.Password) {
		return ErrInvalidCredentials
	}

	_, err = s.UserRepo.Delete(ctx, userID)
	return err
}

// tokenPrefix is prepended to every generated API token so that leaked tokens
// can be identified and traced back to this service (e.g. by secret scanners).
const tokenPrefix = "urlspace_"

// tokenRandomBytes is the number of cryptographically random bytes used to
// generate the random part of an API token. 32 bytes = 256 bits of entropy,
// which is more than sufficient to prevent brute-force guessing.
const tokenRandomBytes = 32

// generateToken creates a new API token in the format "urlspace_<random>", where
// <random> is a base64url-encoded string derived from 32 random bytes.
func generateToken() (string, error) {
	b := make([]byte, tokenRandomBytes)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return tokenPrefix + base64.RawURLEncoding.EncodeToString(b), nil
}

// hashToken produces a SHA-256 hex digest of the given token. SHA-256 is
// appropriate here (unlike passwords) because API tokens are high-entropy
// random strings that cannot be brute-forced — a fast hash is sufficient.
func hashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}

type TokenCreateResult struct {
	RawToken string
}

// TokenCreate generates a new API token for the user after verifying their
// password. The raw token is returned once and never stored — only its
// SHA-256 hash is persisted.
func (s *Service) TokenCreate(ctx context.Context, userID uuid.UUID, password, description string) (TokenCreateResult, error) {
	password, err := validatePassword(password)
	if err != nil {
		return TokenCreateResult{}, err
	}
	description, err = validateTokenDescription(description)
	if err != nil {
		return TokenCreateResult{}, err
	}

	u, err := s.UserRepo.GetById(ctx, userID)
	if err != nil {
		return TokenCreateResult{}, err
	}

	if !passwordMatchesHash(password, u.Password) {
		return TokenCreateResult{}, ErrInvalidCredentials
	}

	rawToken, err := generateToken()
	if err != nil {
		return TokenCreateResult{}, fmt.Errorf("failed to generate token: %w", err)
	}

	_, err = s.TokenRepo.Create(ctx, TokenCreateParams{
		UserID:      userID,
		Description: description,
		Hash:        hashToken(rawToken),
	})
	if err != nil {
		return TokenCreateResult{}, err
	}

	return TokenCreateResult{RawToken: rawToken}, nil
}

// TokenList returns all API tokens for the given user.
func (s *Service) TokenList(ctx context.Context, userID uuid.UUID) ([]Token, error) {
	return s.TokenRepo.List(ctx, userID)
}

// TokenGetByID retrieves a single API token by ID, scoped to the given user.
func (s *Service) TokenGetByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) (Token, error) {
	return s.TokenRepo.GetByID(ctx, id, userID)
}

// TokenDelete removes a single API token by ID, scoped to the given user.
func (s *Service) TokenDelete(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	return s.TokenRepo.Delete(ctx, id, userID)
}

// TokenDeleteAll removes all API tokens for the given user.
func (s *Service) TokenDeleteAll(ctx context.Context, userID uuid.UUID) error {
	return s.TokenRepo.DeleteAllByUserID(ctx, userID)
}

// GetTokenByHash retrieves a token by its hash (used by auth middleware).
func (s *Service) GetTokenByHash(ctx context.Context, rawToken string) (Token, error) {
	return s.TokenRepo.GetByHash(ctx, hashToken(rawToken))
}

// UpdateTokenLastUsedAt updates the last_used_at timestamp (used by auth middleware).
func (s *Service) UpdateTokenLastUsedAt(ctx context.Context, id uuid.UUID) error {
	return s.TokenRepo.UpdateLastUsedAt(ctx, id)
}
