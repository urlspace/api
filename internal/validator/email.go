package validator

import (
	"errors"
	"net/mail"
)

func Email(e string) error {
	if len(e) == 0 {
		return errors.New("email is required")
	}

	// validate format RFC 5322
	if _, err := mail.ParseAddress(e); err != nil {
		return errors.New("email format is invalid")
	}

	// limit length as per smtp spec RFC 5321
	if len(e) > 254 {
		return errors.New("email must be at most 254 characters")
	}

	return nil
}
