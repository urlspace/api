package validator

import (
	"errors"
)

func ResourceReadLater(f *bool) error {
	if f == nil {
		return errors.New("readLater field is required")
	}

	return nil
}
