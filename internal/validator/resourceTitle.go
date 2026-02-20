package validator

import (
	"errors"
)

const (
	resourceTitleLengthMin = 3
	resourceTitleLengthMax = 255
)

func ResourceTitle(rt string) error {
	if len(rt) < resourceTitleLengthMin || len(rt) > resourceTitleLengthMax {
		return errors.New("title must be between 3 and 255 characters")
	}

	return nil
}
