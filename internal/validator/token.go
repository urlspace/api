package validator

import (
	"errors"
	"github.com/google/uuid"
)

func Token(token string) error {
	if len(token) == 0 {
		return errors.New("token is required")
	}

	if _, err := uuid.Parse(token); err != nil {
		return errors.New("token is invalid")
	}

	return nil
}
