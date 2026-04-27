package link

import (
	"net"
	"net/url"
	"strings"
	"unicode"
	"unicode/utf8"
)

const (
	linkTitleLengthMin = 3
	linkTitleLengthMax = 255
)

// ValidateTitle is exported because the uow service needs to validate
// link fields before writing them within a cross-repository transaction.
func ValidateTitle(t string) (string, error) {
	t = strings.TrimSpace(t)

	// Use RuneCountInString instead of len to count human-readable characters,
	// not bytes. Non-ASCII characters (e.g. Polish ąęł, CJK) are multi-byte
	// in UTF-8 and would inflate the byte count, causing valid titles to be
	// rejected or invalid ones to pass.
	n := utf8.RuneCountInString(t)
	if n < linkTitleLengthMin || n > linkTitleLengthMax {
		return t, ErrValidationTitleLength
	}

	// Reject control characters (null bytes, tabs, newlines, etc.) which can
	// cause issues in logs, CSV exports, database collation, or rendering.
	for _, r := range t {
		if unicode.IsControl(r) {
			return t, ErrValidationTitleInvalidCharacters
		}
	}

	return t, nil
}

const (
	linkDescriptionLengthMax = 512
)

// ValidateDescription is exported because the uow service needs to validate
// link fields before writing them within a cross-repository transaction.
func ValidateDescription(d string) (string, error) {
	d = strings.TrimSpace(d)

	// Use RuneCountInString instead of len to count human-readable characters,
	// not bytes. Non-ASCII characters (e.g. Polish ąęł, CJK) are multi-byte
	// in UTF-8 and would inflate the byte count, causing valid descriptions to be
	// rejected or invalid ones to pass.
	if utf8.RuneCountInString(d) > linkDescriptionLengthMax {
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

// ValidateURL is exported because the uow service needs to validate
// link fields before writing them within a cross-repository transaction.
func ValidateURL(u string) (string, error) {
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

	// Reject URLs containing credentials (e.g. http://user:secret@example.com).
	// These would be stored in plaintext and exposed in shared collections.
	if uParsed.User != nil {
		return u, ErrValidationURLFormat
	}

	// RFC 3986 §6.2.2.1: host is case-insensitive. Normalize to lowercase
	// so that HTTPS://EXAMPLE.COM and https://example.com are stored identically.
	uParsed.Host = strings.ToLower(uParsed.Host)

	if isPrivateHost(uParsed.Host) {
		return u, ErrValidationURLPrivate
	}

	return uParsed.String(), nil
}

var privateBlocks []*net.IPNet

func init() {
	for _, cidr := range []string{
		"127.0.0.0/8",    // loopback
		"10.0.0.0/8",     // RFC 1918 private
		"172.16.0.0/12",  // RFC 1918 private
		"192.168.0.0/16", // RFC 1918 private
		"169.254.0.0/16", // link-local (AWS/GCP/Azure metadata endpoint)
		"::1/128",        // IPv6 loopback
		"fc00::/7",       // IPv6 unique local
	} {
		_, block, err := net.ParseCIDR(cidr)
		if err != nil {
			panic(err)
		}
		privateBlocks = append(privateBlocks, block)
	}
}

// isPrivateHost checks whether a host (with optional port) resolves to a
// loopback, private (RFC 1918), link-local, or IPv6 unique local address.
// These are blocked to prevent SSRF if the backend ever fetches stored URLs
// (e.g. link previews, favicons) and to avoid unsafe links in shared collections.
func isPrivateHost(host string) bool {
	h := host
	if hostOnly, _, err := net.SplitHostPort(host); err == nil {
		h = hostOnly
	}
	h = strings.Trim(h, "[]")

	if strings.EqualFold(h, "localhost") {
		return true
	}

	ip := net.ParseIP(h)
	if ip == nil {
		return false
	}

	for _, block := range privateBlocks {
		if block.Contains(ip) {
			return true
		}
	}

	return false
}
