package resource

import (
	"net/url"
	"strings"
	"unicode/utf8"
)

const (
	resourceTitleLengthMin = 3
	resourceTitleLengthMax = 255
)

func validateTitle(t string) (string, error) {
	t = strings.TrimSpace(t)

	// Use RuneCountInString instead of len to count human-readable characters,
	// not bytes. Non-ASCII characters (e.g. Polish ąęł, CJK) are multi-byte
	// in UTF-8 and would inflate the byte count, causing valid titles to be
	// rejected or invalid ones to pass.
	if utf8.RuneCountInString(t) < resourceTitleLengthMin || utf8.RuneCountInString(t) > resourceTitleLengthMax {
		return t, ErrValidationTitleLength
	}

	return t, nil
}

const (
	resourceDescriptionLengthMax = 512
)

func validateDescription(d string) (string, error) {
	d = strings.TrimSpace(d)

	// Use RuneCountInString instead of len to count human-readable characters,
	// not bytes. See validateTitle for details.
	if utf8.RuneCountInString(d) > resourceDescriptionLengthMax {
		return d, ErrValidationDescriptionLength
	}

	return d, nil
}

func validateURL(u string) (string, error) {
	u = strings.TrimSpace(u)

	// 2048 characters is the practical URL length limit enforced by most
	// modern browsers (Chrome, Firefox, Safari)
	if len(u) > 2048 {
		return u, ErrValidationURLTooLong
	}

	uParsed, err := url.Parse(u)
	if err != nil || uParsed.Host == "" {
		return u, ErrValidationURLFormat
	}

	// Only allow http and https schemes to prevent XSS via javascript:,
	// data:, or other dangerous URI schemes when rendered as clickable links.
	if uParsed.Scheme != "http" && uParsed.Scheme != "https" {
		return u, ErrValidationURLFormat
	}

	return uParsed.String(), nil
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
