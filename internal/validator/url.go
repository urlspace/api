package validator

import (
	"errors"
	"net/url"
)

func Url(uri string) error {
	u, err := url.Parse(uri)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return errors.New("url is invalid")
	}

	return nil
}
