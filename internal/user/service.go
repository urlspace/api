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
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/urlspace/api/internal/config"
	"github.com/urlspace/api/internal/emails"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"golang.org/x/crypto/argon2"
)

var tracer = otel.Tracer("github.com/urlspace/api/internal/user")

type CreateParams struct {
	Email                           string
	EmailVerified                   bool
	EmailVerificationTokenHash      *string
	EmailVerificationTokenExpiresAt *time.Time
	Password                        string
	Username                        string
	DisplayName                     string
	IsAdmin                         bool
	IsPro                           bool
}

type UpdateVerificationTokenParams struct {
	ID                              uuid.UUID
	EmailVerificationTokenHash      *string
	EmailVerificationTokenExpiresAt *time.Time
}

type UpdatePasswordResetTokenParams struct {
	ID                          uuid.UUID
	PasswordResetTokenHash      *string
	PasswordResetTokenExpiresAt *time.Time
}

type Repository interface {
	List(ctx context.Context) ([]User, error)
	GetById(ctx context.Context, id uuid.UUID) (User, error)
	GetByEmail(ctx context.Context, email string) (User, error)
	GetByEmailVerificationTokenHash(ctx context.Context, hash string) (User, error)
	GetByPasswordResetTokenHash(ctx context.Context, hash string) (User, error)
	Create(ctx context.Context, params CreateParams) (User, error)
	Verify(ctx context.Context, id uuid.UUID) (User, error)
	UpdateVerificationToken(ctx context.Context, params UpdateVerificationTokenParams) (User, error)
	UpdatePasswordResetToken(ctx context.Context, params UpdatePasswordResetTokenParams) (User, error)
	ResetPassword(ctx context.Context, id uuid.UUID, passwordHash string) (User, error)
	Delete(ctx context.Context, id uuid.UUID) (User, error)
}

type SessionCreateParams struct {
	UserID      uuid.UUID
	SessionHash string
	Description *string
	ExpiresAt   time.Time
}

type SessionUpdateExpiresAtParams struct {
	ID        uuid.UUID
	ExpiresAt time.Time
}

type SessionRepository interface {
	Create(ctx context.Context, params SessionCreateParams) (Session, error)
	GetByHash(ctx context.Context, sessionHash string) (Session, error)
	UpdateExpiresAt(ctx context.Context, params SessionUpdateExpiresAtParams) (Session, error)
	DeleteByHash(ctx context.Context, sessionHash string) error
	DeleteAllByUserID(ctx context.Context, userID uuid.UUID) error
}

type TokenCreateParams struct {
	UserID      uuid.UUID
	Description string
	TokenHash   string
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
func passwordHash(ctx context.Context, password string) (string, error) {
	_, span := tracer.Start(ctx, "user.password.hash")
	defer span.End()

	// Generate a cryptographically random salt. Each password gets its own
	// unique salt so that two users with the same password will have
	// completely different hashes (prevents rainbow table attacks).
	salt := make([]byte, argonSaltLength)
	if _, err := rand.Read(salt); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
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
func passwordMatchesHash(ctx context.Context, password, hash string) bool {
	_, span := tracer.Start(ctx, "user.password.match.hash")
	defer span.End()

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
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return false
	}

	// Sscanf's %s captured "salt$key" as one string because %s reads until
	// whitespace, not until $. Split on the first $ to separate them.
	parts := strings.SplitN(saltB64, "$", 2)
	if len(parts) != 2 {
		err := errors.New("invalid argon2 hash format")
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return false
	}
	saltB64 = parts[0]
	keyB64 = parts[1]

	// Decode the base64-encoded salt and expected key back into raw bytes.
	salt, err := base64.RawStdEncoding.DecodeString(saltB64)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return false
	}
	expectedKey, err := base64.RawStdEncoding.DecodeString(keyB64)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
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
	// Username validation errors.
	ErrValidationUsernameRequired   = errors.New("username is required")
	ErrValidationUsernameTooShort   = errors.New("username must be min 3 characters")
	ErrValidationUsernameTooLong    = errors.New("username must be max 32 characters")
	ErrValidationUsernameCharacters = errors.New("username can only contain lowercase characters, numbers, hyphens, and underscores")
	ErrValidationUsernamePrefix     = errors.New("username cannot start with hyphen or underscore")
	ErrValidationUsernameSuffix     = errors.New("username cannot end with hyphen or underscore")
	ErrValidationUsernameReserved   = errors.New("username is reserved")

	// Email validation errors.
	ErrValidationEmailRequired = errors.New("email is required")
	ErrValidationEmailFormat   = errors.New("email format is invalid")
	ErrValidationEmailTooLong  = errors.New("email must be at most 254 characters")

	// Password validation errors.
	ErrValidationPasswordRequired        = errors.New("password is required")
	ErrValidationPasswordTooShort        = errors.New("password must be at least 12 characters")
	ErrValidationPasswordTooLong         = errors.New("password must be at most 128 characters")
	ErrValidationPasswordContainsContext = errors.New("password cannot contain your username, display name, or email")

	// Token validation errors.
	ErrValidationTokenRequired = errors.New("token is required")
	ErrValidationTokenFormat   = errors.New("token is invalid")

	// Admin flag validation errors.
	ErrValidationIsAdminRequired = errors.New("isAdmin flag is required")

	// Pro flag validation errors.
	ErrValidationIsProRequired = errors.New("isPro flag is required")

	// API token description validation errors.
	ErrValidationTokenDescriptionRequired = errors.New("token description is required")
	ErrValidationTokenDescriptionTooLong  = errors.New("token description must be at most 255 characters")

	ErrNotFound           = errors.New("not found")
	ErrConflict           = errors.New("conflict")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrEmailNotVerified   = errors.New("invalid email or password")
	ErrTokenExpired       = errors.New("token has expired")
)

type Service struct {
	UserRepo    Repository
	SessionRepo SessionRepository
	TokenRepo   TokenRepository
	EmailSender emails.EmailSender
	AppURL      string
}

func NewService(userRepo Repository, sessionRepo SessionRepository, tokenRepo TokenRepository, emailSender emails.EmailSender, appURL string) *Service {
	return &Service{
		UserRepo:    userRepo,
		SessionRepo: sessionRepo,
		TokenRepo:   tokenRepo,
		EmailSender: emailSender,
		AppURL:      appURL,
	}
}

// Signup creates a new unverified user and sends a verification email.
func (s *Service) Signup(ctx context.Context, username, email, password string) error {
	username, err := validateUsername(username)
	if err != nil {
		return err
	}
	email, err = validateEmail(email)
	if err != nil {
		return err
	}
	emailLocalPart := strings.SplitN(email, "@", 2)[0]
	password, err = validatePassword(password, username, emailLocalPart)
	if err != nil {
		return err
	}

	passwordHash, err := passwordHash(ctx, password)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	token, err := generateToken()
	if err != nil {
		return fmt.Errorf("failed to generate verification token: %w", err)
	}
	tokenHash := hashToken(token)
	expiresAt := time.Now().Add(config.EmailVerificationTokenExpiryDuration)

	// Decision (future me): ErrConflict on duplicate email/username surfaces
	// as 409, which is an account-enumeration oracle. Deliberately not fixed
	// here — silent-success would leave real users wondering why no email
	// ever arrives, hurting signup conversion. Rate-limiting at the request
	// layer is the chosen compensating control. Revisit if rate-limiting
	// proves insufficient or enumeration becomes a real attack target.
	_, err = s.UserRepo.Create(ctx, CreateParams{
		Email:                           email,
		EmailVerified:                   false,
		EmailVerificationTokenHash:      &tokenHash,
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
		Url:      s.AppURL + "/auth/signup/" + token,
	}
	bodyHtml, err := emails.RenderTemplateHtml(emails.AuthSignupTemplateHtml, emailVerifyData)
	if err != nil {
		return fmt.Errorf("failed to render html email template: %w", err)
	}
	bodyText, err := emails.RenderTemplateTxt(emails.AuthSignupTemplateTxt, emailVerifyData)
	if err != nil {
		return fmt.Errorf("failed to render text email template: %w", err)
	}

	// Decision (future me): fire-and-forget. Synchronous Send creates a
	// timing oracle (registered-email path is ~Resend RTT slower than
	// silent-success branches) and couples request latency to Resend
	// availability. WithoutCancel is required, not defensive — the request
	// context is cancelled the moment this handler returns, which would
	// abort the in-flight Resend call. Recover protects the process from
	// a panic in a detached goroutine.
	detached := context.WithoutCancel(ctx)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				slog.ErrorContext(detached, "email send panicked", slog.Any("recover", r))
			}
		}()
		if err := s.EmailSender.Send(detached, emails.EmailSendParams{
			To:      []string{email},
			Text:    bodyText,
			Html:    bodyHtml,
			Subject: "Hello from url.space",
		}); err != nil {
			slog.ErrorContext(detached, "failed to send email", slog.String("error", err.Error()))
		}
	}()

	return nil
}

// dummyHash is a pre-computed Argon2id hash used to prevent timing-based user
// enumeration. When a sign-in attempt uses an unknown email, we still run the
// hash computation against this dummy so the response time is indistinguishable
// from a wrong-password attempt.
var dummyHash = func() string {
	h, err := passwordHash(context.Background(), "dummy-password-never-matches")
	if err != nil {
		panic("failed to generate dummy hash: " + err.Error())
	}
	return h
}()

type SigninResult struct {
	Session   string
	ExpiresAt time.Time
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
	userExists := err == nil
	if err != nil && !errors.Is(err, ErrNotFound) {
		return SigninResult{}, err
	}

	// Decision (future me): always run the hash check before the EmailVerified
	// check. Otherwise status code (403 vs 401) and timing (skipped argon2id)
	// both leak whether an email is registered. Do not "tidy" the order.
	hash := dummyHash
	if userExists {
		hash = u.Password
	}
	if !passwordMatchesHash(ctx, password, hash) {
		return SigninResult{}, ErrInvalidCredentials
	}
	if !userExists {
		return SigninResult{}, ErrInvalidCredentials
	}
	if !u.EmailVerified {
		return SigninResult{}, ErrEmailNotVerified
	}

	session, err := generateToken()
	if err != nil {
		return SigninResult{}, fmt.Errorf("failed to generate session: %w", err)
	}

	expiresAt := time.Now().Add(config.SessionExpiryDuration)
	_, err = s.SessionRepo.Create(ctx, SessionCreateParams{
		UserID:      u.ID,
		SessionHash: hashToken(session),
		Description: description,
		ExpiresAt:   expiresAt,
	})
	if err != nil {
		return SigninResult{}, err
	}

	return SigninResult{Session: session, ExpiresAt: expiresAt}, nil
}

// Verify validates a verification token and marks the user as verified.
func (s *Service) Verify(ctx context.Context, token string) error {
	token, err := validateToken(token)
	if err != nil {
		return err
	}

	u, err := s.UserRepo.GetByEmailVerificationTokenHash(ctx, hashToken(token))
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
// Throttles silently (returns nil) if the last token was issued < 5 min ago,
// so the response is indistinguishable from the unknown-email and already-
// verified branches and can't be used to enumerate accounts.
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
		// Decision (future me): silent return (not 429) — see function doc.
		// A 429 here would confirm the email is registered and unverified.
		slog.InfoContext(ctx, "verification resend throttled", slog.String("user_id", u.ID.String()))
		return nil
	}

	token, err := generateToken()
	if err != nil {
		return fmt.Errorf("failed to generate verification token: %w", err)
	}
	tokenHash := hashToken(token)
	expiresAt := time.Now().Add(config.EmailVerificationTokenExpiryDuration)

	_, err = s.UserRepo.UpdateVerificationToken(ctx, UpdateVerificationTokenParams{
		ID:                              u.ID,
		EmailVerificationTokenHash:      &tokenHash,
		EmailVerificationTokenExpiresAt: &expiresAt,
	})
	if err != nil {
		return err
	}

	templateParams := emails.AuthResendVerificationParams{
		Url: s.AppURL + "/auth/signup/" + token,
	}
	bodyHtml, err := emails.RenderTemplateHtml(emails.AuthResendVerificationTemplateHtml, templateParams)
	if err != nil {
		return fmt.Errorf("failed to render html email template: %w", err)
	}
	bodyText, err := emails.RenderTemplateTxt(emails.AuthResendVerificationTemplateTxt, templateParams)
	if err != nil {
		return fmt.Errorf("failed to render text email template: %w", err)
	}

	// Decision (future me): fire-and-forget. Synchronous Send creates a
	// timing oracle (registered-email path is ~Resend RTT slower than
	// silent-success branches) and couples request latency to Resend
	// availability. WithoutCancel is required, not defensive — the request
	// context is cancelled the moment this handler returns, which would
	// abort the in-flight Resend call. Recover protects the process from
	// a panic in a detached goroutine.
	detached := context.WithoutCancel(ctx)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				slog.ErrorContext(detached, "email send panicked", slog.Any("recover", r))
			}
		}()
		if err := s.EmailSender.Send(detached, emails.EmailSendParams{
			To:      []string{email},
			Text:    bodyText,
			Html:    bodyHtml,
			Subject: "Verification token has been requested",
		}); err != nil {
			slog.ErrorContext(detached, "failed to send email", slog.String("error", err.Error()))
		}
	}()

	return nil
}

// ResetPasswordRequest generates a password reset token and sends an email.
// Throttles silently (returns nil) if the last token was issued < 5 min ago,
// so the response is indistinguishable from the unknown-email branch and
// can't be used to enumerate accounts.
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
			// Decision (future me): silent return (not 429) — see function doc.
			// A 429 here would confirm the email is registered.
			slog.InfoContext(ctx, "password reset throttled", slog.String("user_id", u.ID.String()))
			return nil
		}
	}

	token, err := generateToken()
	if err != nil {
		return fmt.Errorf("failed to generate password reset token: %w", err)
	}
	tokenHash := hashToken(token)
	expiresAt := time.Now().Add(config.PasswordResetTokenExpiryDuration)

	_, err = s.UserRepo.UpdatePasswordResetToken(ctx, UpdatePasswordResetTokenParams{
		ID:                          u.ID,
		PasswordResetTokenHash:      &tokenHash,
		PasswordResetTokenExpiresAt: &expiresAt,
	})
	if err != nil {
		return err
	}

	templateParams := emails.AuthResetPasswordRequestParams{
		Url: s.AppURL + "/auth/reset-password/" + token,
	}
	bodyHtml, err := emails.RenderTemplateHtml(emails.AuthResetPasswordRequestTemplateHtml, templateParams)
	if err != nil {
		return fmt.Errorf("failed to render html email template: %w", err)
	}
	bodyText, err := emails.RenderTemplateTxt(emails.AuthResetPasswordRequestTemplateTxt, templateParams)
	if err != nil {
		return fmt.Errorf("failed to render text email template: %w", err)
	}

	// Decision (future me): fire-and-forget. Synchronous Send creates a
	// timing oracle (registered-email path is ~Resend RTT slower than
	// silent-success branches) and couples request latency to Resend
	// availability. WithoutCancel is required, not defensive — the request
	// context is cancelled the moment this handler returns, which would
	// abort the in-flight Resend call. Recover protects the process from
	// a panic in a detached goroutine.
	detached := context.WithoutCancel(ctx)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				slog.ErrorContext(detached, "email send panicked", slog.Any("recover", r))
			}
		}()
		if err := s.EmailSender.Send(detached, emails.EmailSendParams{
			To:      []string{email},
			Text:    bodyText,
			Html:    bodyHtml,
			Subject: "Password reset has been requested",
		}); err != nil {
			slog.ErrorContext(detached, "failed to send email", slog.String("error", err.Error()))
		}
	}()

	return nil
}

// ResetPasswordConfirm validates a reset token and sets the new password.
func (s *Service) ResetPasswordConfirm(ctx context.Context, token, newPassword string) error {
	token, err := validateToken(token)
	if err != nil {
		return err
	}

	u, err := s.UserRepo.GetByPasswordResetTokenHash(ctx, hashToken(token))
	if err != nil {
		return err
	}

	if u.PasswordResetTokenExpiresAt != nil && u.PasswordResetTokenExpiresAt.Before(time.Now()) {
		return ErrTokenExpired
	}

	emailLocalPart := strings.SplitN(u.Email, "@", 2)[0]
	newPassword, err = validatePassword(newPassword, u.Username, u.DisplayName, emailLocalPart)
	if err != nil {
		return err
	}

	passwordHash, err := passwordHash(ctx, newPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// ResetPassword also marks the user as verified and clears any pending
	// verification token. Completing a password reset proves the user controls
	// the inbox — the same guarantee email verification provides — so we treat
	// it as an implicit verification. This avoids a dead-end where an unverified
	// user resets their password but is still blocked at signin.
	_, err = s.UserRepo.ResetPassword(ctx, u.ID, passwordHash)
	if err != nil {
		return err
	}

	// Decision (future me): fire-and-forget session cleanup. We don't fail
	// the request on error — the password reset itself already succeeded and
	// the caller can't usefully retry. Log on error so we notice if old
	// sessions are silently surviving a reset. WithoutCancel preserves trace
	// context across the detached call. Recover protects the process from a
	// panic in pgx/tracing/etc., since an unrecovered panic in a goroutine
	// crashes the whole API server.
	detached := context.WithoutCancel(ctx)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				slog.ErrorContext(detached, "session deletion panicked", slog.Any("recover", r))
			}
		}()
		if err := s.SessionRepo.DeleteAllByUserID(detached, u.ID); err != nil {
			slog.ErrorContext(detached, "failed to delete sessions after password reset", slog.String("error", err.Error()), slog.String("user_id", u.ID.String()))
		}
	}()

	return nil
}

// Signout deletes a session identified by the raw cookie value.
// The cookie value is hashed before deletion; the DB never sees the secret.
func (s *Service) Signout(ctx context.Context, session string) error {
	return s.SessionRepo.DeleteByHash(ctx, hashToken(session))
}

// GetSession retrieves a session record by the raw cookie value (used by auth
// middleware). The cookie value is hashed before lookup; the DB never sees the
// secret, so a read-only DB compromise does not yield usable session tokens.
func (s *Service) GetSession(ctx context.Context, session string) (Session, error) {
	return s.SessionRepo.GetByHash(ctx, hashToken(session))
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
	email, err = validateEmail(email)
	if err != nil {
		return User{}, err
	}
	emailLocalPart := strings.SplitN(email, "@", 2)[0]
	password, err = validatePassword(password, username, emailLocalPart)
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
	passwordHash, err := passwordHash(ctx, password)
	if err != nil {
		return User{}, fmt.Errorf("failed to hash password: %w", err)
	}

	return s.UserRepo.Create(ctx, CreateParams{
		Email:                           email,
		EmailVerified:                   true,
		EmailVerificationTokenHash:      nil,
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

	if !passwordMatchesHash(ctx, password, u.Password) {
		return ErrInvalidCredentials
	}

	_, err = s.UserRepo.Delete(ctx, userID)
	return err
}

// tokenRandomBytes is the entropy budget for every server-issued bearer
// credential generated by this package — sessions, API tokens, email
// verification tokens, and password reset tokens. 32 bytes = 256 bits, far
// exceeding the OWASP minimum (64 bits) for session identifiers and
// unguessable by any realistic attacker. Replaces the earlier uuidv7-as-session
// and uuid.New-as-email-token designs, which carried only ~74 and ~122
// effective bits respectively and were stored as plaintext.
const tokenRandomBytes = 32

// tokenPrefix is prepended to every generated API token (and only API tokens)
// so leaked Bearer values can be recognised in source code and logs by secret
// scanners. Sessions live in cookies, email tokens live in URL paths — neither
// is ever pasted into a repository, so neither needs the marker.
const tokenPrefix = "urlspace_"

// generateToken returns a fresh, high-entropy random string suitable for any
// server-issued opaque bearer credential — a session cookie value, an API
// token (after the caller prepends tokenPrefix), or a verification/reset URL
// path segment. The output is base64 URL-safe with no padding so it drops
// cleanly into any of those contexts without escaping.
//
// The prefix on API tokens is added at the single call site that needs it
// rather than here: the prefix is purely cosmetic (for secret scanners) and
// keeping this helper prefix-free lets one function serve all four flows.
func generateToken() (string, error) {
	b := make([]byte, tokenRandomBytes)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// hashToken produces the SHA-256 hex digest stored in any *hash column —
// sessions.hash, tokens.hash, users.email_verification_token_hash,
// users.password_reset_token_hash. The raw token leaves the server only
// once (in a Set-Cookie header, a one-time TokenCreate response, or an
// email URL); the database only ever sees the digest.
//
// SHA-256 is appropriate here precisely because the input is 256-bit CSPRNG
// output: an attacker cannot brute-force what they cannot enumerate, so a
// fast hash is sufficient and a slow one (Argon2 etc.) would only burn CPU
// on every auth check. The hashing exists for defence in depth — a read-only
// DB compromise (leaked backup, replica access, SQLi elsewhere) yields no
// usable credentials of any kind, because the client always sends the raw
// value and we only ever compare its hash against the stored column.
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
// Tokens (and sessions) live in the user service because they're auth
// artifacts of a user identity, not independent domains. TokenCreate
// requires password verification, which would force a cross-domain
// import if tokens lived elsewhere.
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

	if !passwordMatchesHash(ctx, password, u.Password) {
		return TokenCreateResult{}, ErrInvalidCredentials
	}

	token, err := generateToken()
	if err != nil {
		return TokenCreateResult{}, fmt.Errorf("failed to generate token: %w", err)
	}
	// API tokens are the only credential that gets the human-readable prefix —
	// see tokenPrefix's doc comment. Sessions and email tokens stay prefix-free
	// because they're never pasted into source code.
	token = tokenPrefix + token

	_, err = s.TokenRepo.Create(ctx, TokenCreateParams{
		UserID:      userID,
		Description: description,
		TokenHash:   hashToken(token),
	})
	if err != nil {
		return TokenCreateResult{}, err
	}

	return TokenCreateResult{RawToken: token}, nil
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
func (s *Service) GetTokenByHash(ctx context.Context, token string) (Token, error) {
	return s.TokenRepo.GetByHash(ctx, hashToken(token))
}

// UpdateTokenLastUsedAt updates the last_used_at timestamp (used by auth middleware).
func (s *Service) UpdateTokenLastUsedAt(ctx context.Context, id uuid.UUID) error {
	return s.TokenRepo.UpdateLastUsedAt(ctx, id)
}
