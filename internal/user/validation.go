package user

import (
	"net/mail"
	"regexp"
	"strings"

	"github.com/google/uuid"
)

func validateEmail(e string) (string, error) {
	e = strings.ToLower(strings.TrimSpace(e))

	if len(e) == 0 {
		return e, ErrValidationEmailRequired
	}

	// validate format RFC 5322
	if _, err := mail.ParseAddress(e); err != nil {
		return e, ErrValidationEmailFormat
	}

	// limit length as per smtp spec RFC 5321
	if len(e) > 254 {
		return e, ErrValidationEmailTooLong
	}

	return e, nil
}

const (
	passwordLengthMin = 12
)

func validatePassword(p string) error {
	if len(p) == 0 {
		return ErrValidationPasswordRequired
	}

	if len(p) < passwordLengthMin {
		return ErrValidationPasswordTooShort
	}

	return nil
}

func validateToken(token string) error {
	if len(token) == 0 {
		return ErrValidationTokenRequired
	}

	if _, err := uuid.Parse(token); err != nil {
		return ErrValidationTokenFormat
	}

	return nil
}

var reservedUsernames = map[string]bool{
	"admin": true,
}

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
