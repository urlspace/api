package resource

import (
	"net/url"
	"strings"
)

const (
	resourceTitleLengthMin = 3
	resourceTitleLengthMax = 255
)

func validateTitle(t string) (string, error) {
	t = strings.TrimSpace(t)

	if len(t) < resourceTitleLengthMin || len(t) > resourceTitleLengthMax {
		return t, ErrValidationTitleLength
	}

	return t, nil
}

const (
	resourceDescriptionLengthMax = 512
)

func validateDescription(d string) (string, error) {
	d = strings.TrimSpace(d)

	if len(d) > resourceDescriptionLengthMax {
		return d, ErrValidationDescriptionLength
	}

	return d, nil
}

func validateURL(u string) (string, error) {
	uParsed, err := url.Parse(u)
	if err != nil || uParsed.Scheme == "" || uParsed.Host == "" {
		return u, ErrValidationURLFormat
	}

	return u, nil
}

func validateFavourite(f *bool) (bool, error) {
	if f == nil {
		return false, ErrValidationFavouriteRequired
	}

	return *f, nil
}

func validateReadLater(r *bool) (bool, error) {
	if r == nil {
		return false, ErrValidationReadLaterRequired
	}

	return *r, nil
}
