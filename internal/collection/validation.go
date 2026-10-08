package collection

import (
	"strings"
	"unicode"
	"unicode/utf8"

	"golang.org/x/text/unicode/norm"
)

const (
	collectionNameLengthMin = 2
	collectionNameLengthMax = 128
)

func ValidateName(n string) (string, error) {
	n = strings.TrimSpace(n)

	// Reject control characters (null bytes, tabs, newlines, etc.) which can
	// cause issues in logs, CSV exports, database collation, or rendering.
	// Runs before whitespace is collapsed, so a tab is rejected rather than
	// silently turned into a space.
	for _, r := range n {
		if unicode.IsControl(r) {
			return n, ErrValidationNameInvalidCharacters
		}
	}

	// Store one canonical form so names that look identical compare equal:
	// NFC merges "e" + combining accent into "é", and Fields collapses runs of
	// any Unicode whitespace (including non-breaking spaces) into one space.
	n = norm.NFC.String(n)
	n = strings.Join(strings.Fields(n), " ")

	// Use RuneCountInString instead of len to count human-readable characters,
	// not bytes. Non-ASCII characters (e.g. Polish ąęł, CJK) are multi-byte
	// in UTF-8 and would inflate the byte count, causing valid names to be
	// rejected or invalid ones to pass.
	count := utf8.RuneCountInString(n)
	if count < collectionNameLengthMin || count > collectionNameLengthMax {
		return n, ErrValidationNameLength
	}

	return n, nil
}

const (
	collectionDescriptionLengthMax = 1024
)

func ValidateDescription(d string) (string, error) {
	d = strings.TrimSpace(d)

	// Use RuneCountInString instead of len to count human-readable characters,
	// not bytes. Non-ASCII characters (e.g. Polish ąęł, CJK) are multi-byte
	// in UTF-8 and would inflate the byte count, causing valid descriptions to be
	// rejected or invalid ones to pass.
	if utf8.RuneCountInString(d) > collectionDescriptionLengthMax {
		return d, ErrValidationDescriptionLength
	}

	// Reject control characters (null bytes, tabs, newlines, etc.) which can
	// cause issues in logs, CSV exports, database collation, or rendering.
	for _, r := range d {
		if unicode.IsControl(r) {
			return d, ErrValidationDescriptionInvalidCharacters
		}
	}

	return d, nil
}
