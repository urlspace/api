package validator

import (
	"errors"
)

func ResourceFavourite(f *bool) error {
	if f == nil {
		return errors.New("favourite field is required")
	}

	return nil
}
