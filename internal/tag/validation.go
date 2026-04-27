package tag

import (
	"regexp"
	"strings"
)

const (
	tagNameLengthMin  = 2
	tagNameLengthMax  = 50
	tagMaxPerLink = 10
)

// Only lowercase ASCII letters, digits, and hyphens are allowed.
var tagNamePattern = regexp.MustCompile(`^[a-z0-9-]+$`)

func validateName(n string) (string, error) {
	n = strings.ToLower(strings.TrimSpace(n))

	// len() is used here instead of utf8.RuneCountInString() because the regex
	// restricts tag names to single-byte ASCII characters.
	if len(n) < tagNameLengthMin || len(n) > tagNameLengthMax {
		return n, ErrValidationNameLength
	}

	if !tagNamePattern.MatchString(n) {
		return n, ErrValidationNameCharacters
	}

	if strings.HasPrefix(n, "-") || strings.HasSuffix(n, "-") {
		return n, ErrValidationNameHyphens
	}

	if strings.Contains(n, "--") {
		return n, ErrValidationNameHyphens
	}

	return n, nil
}

// ValidateTagNames is exported because the uow service needs to validate
// tag names before upserting them within a cross-repository transaction.
func ValidateTagNames(names []string) ([]string, error) {
	seen := make(map[string]struct{}, len(names))
	validated := make([]string, 0, len(names))

	for _, name := range names {
		name, err := validateName(name)
		if err != nil {
			return nil, err
		}

		if _, exists := seen[name]; exists {
			continue
		}
		seen[name] = struct{}{}
		validated = append(validated, name)
	}

	if len(validated) > tagMaxPerLink {
		return nil, ErrValidationTooManyTags
	}

	return validated, nil
}
