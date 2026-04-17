package user

import (
	"net/mail"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"
)

func validateEmail(e string) (string, error) {
	e = strings.ToLower(strings.TrimSpace(e))

	if len(e) == 0 {
		return e, ErrValidationEmailRequired
	}

	// RFC 5321 limits the total length of an email address to 254 characters.
	if len(e) > 254 {
		return e, ErrValidationEmailTooLong
	}

	parsed, err := mail.ParseAddress(e)
	if err != nil {
		return e, ErrValidationEmailFormat
	}

	// mail.ParseAddress accepts RFC 5322 display names like "Joe" <joe@evil.com>.
	// Reject these by ensuring the raw input matches the parsed address exactly.
	if parsed.Address != e {
		return e, ErrValidationEmailFormat
	}

	return e, nil
}

const (
	passwordLengthMin = 12
)

func validatePassword(p string) (string, error) {
	if len(p) == 0 || strings.TrimSpace(p) == "" {
		return p, ErrValidationPasswordRequired
	}

	// Use RuneCountInString instead of len to count human-readable characters,
	// not bytes. Multi-byte characters like emoji (4 bytes) or CJK (3 bytes)
	// would inflate the byte count and pass a len() check with very few actual
	// characters of entropy (e.g. 3 emoji = 12 bytes but only 3 characters).
	if utf8.RuneCountInString(p) < passwordLengthMin {
		return p, ErrValidationPasswordTooShort
	}

	return p, nil
}

func validateToken(token string) (string, error) {
	if len(token) == 0 {
		return token, ErrValidationTokenRequired
	}

	if _, err := uuid.Parse(token); err != nil {
		return token, ErrValidationTokenFormat
	}

	return token, nil
}

var reservedUsernames = map[string]bool{
	// brand
	"href":       true,
	"hreftools":  true,
	"href_tools": true,
	"href-tools": true,
	"yank":       true,
	"yankpage":   true,
	"yank-page":  true,
	"yank_page":  true,
	// infrastructure
	"root":      true,
	"system":    true,
	"localhost": true,
	"www":       true,
	"mail":      true,
	"ftp":       true,
	// impersonation targets
	"support":    true,
	"help":       true,
	"contact":    true,
	"info":       true,
	"security":   true,
	"abuse":      true,
	"postmaster": true,
	"webmaster":  true,
	"noreply":    true,
	"no-reply":   true,
	"admin":      true,
	// authority
	"moderator": true,
	"mod":       true,
	"staff":     true,
	"team":      true,
	"official":  true,
}

// Only lowercase ASCII letters, digits, hyphens, and underscores are allowed.
// Consecutive separators (e.g. "a--b", "a__b") are permitted as they are URL-friendly.
var userPattern = regexp.MustCompile(`^[a-z0-9_-]+$`)

const (
	userUsernameLengthMin = 3
	userUsernameLengthMax = 32
)

func validateUsername(u string) (string, error) {
	u = strings.ToLower(strings.TrimSpace(u))

	if len(u) == 0 {
		return u, ErrValidationUsernameRequired
	}

	// len() is used here instead of utf8.RuneCountInString() because the regex
	// (userPattern) already restricts usernames to single-byte ASCII characters.
	if len(u) < userUsernameLengthMin {
		return u, ErrValidationUsernameTooShort
	}

	if len(u) > userUsernameLengthMax {
		return u, ErrValidationUsernameTooLong
	}

	if !userPattern.MatchString(u) {
		return u, ErrValidationUsernameCharacters
	}

	if strings.HasPrefix(u, "-") || strings.HasPrefix(u, "_") {
		return u, ErrValidationUsernamePrefix
	}

	if strings.HasSuffix(u, "-") || strings.HasSuffix(u, "_") {
		return u, ErrValidationUsernameSuffix
	}

	if reserved := reservedUsernames[u]; reserved {
		return u, ErrValidationUsernameReserved
	}

	return u, nil
}

func validateIsAdmin(isAdmin *bool) (bool, error) {
	if isAdmin == nil {
		return false, ErrValidationIsAdminRequired
	}

	return *isAdmin, nil
}

func validateIsPro(isPro *bool) (bool, error) {
	if isPro == nil {
		return false, ErrValidationIsProRequired
	}

	return *isPro, nil
}
