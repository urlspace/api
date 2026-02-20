package validator

import (
	"errors"
)

const (
	passwordLengthMin = 12
)

func Password(p string) error {
	if len(p) == 0 {
		return errors.New("password is required")
	}

	if len(p) < passwordLengthMin {
		return errors.New("password must be at least 12 characters")
	}

	return nil
}
