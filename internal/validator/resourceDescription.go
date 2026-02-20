package validator

import (
	"errors"
)

const (
	resourceDescriptionLengthMax = 512
)

func ResourceDescription(rd string) error {
	if len(rd) > resourceDescriptionLengthMax {
		return errors.New("description must be less than 512 characters")
	}

	return nil
}
