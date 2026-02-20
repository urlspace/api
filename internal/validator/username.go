package validator

import (
	"errors"
	"regexp"
	"strings"
)

var reservedUsernames = map[string]bool{
	"admin": true,
}

const (
	userUsernameLengthMin = 3
	userUsernameLengthMax = 32
)

var UserPattern = regexp.MustCompile(`^[a-z0-9_-]+$`)

func Username(u string) error {
	if len(u) == 0 {
		return errors.New("username is required")
	}

	if len(u) < userUsernameLengthMin {
		return errors.New("username must be min 3 characters")
	}

	if len(u) > userUsernameLengthMax {
		return errors.New("username must be max 32 characters")
	}

	if !UserPattern.MatchString(u) {
		return errors.New("username can only contain lowercase characters, numbers, hyphens, and underscores")
	}

	if strings.HasPrefix(u, "-") || strings.HasPrefix(u, "_") {
		return errors.New("username cannot start with hyphen or underscore")
	}

	if strings.HasSuffix(u, "-") || strings.HasSuffix(u, "_") {
		return errors.New("username cannot end with hyphen or underscore")
	}

	if reserved := reservedUsernames[u]; reserved {
		return errors.New("username is reserved")
	}

	return nil
}
