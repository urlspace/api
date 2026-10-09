package collection

import (
	"regexp"
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
	collectionSlugLengthMin = 2
	collectionSlugLengthMax = 128
)

var slugPattern = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

func ValidateSlug(s string) (string, error) {
	s = strings.ToLower(strings.TrimSpace(s))

	// Slugs are ASCII-only, so byte length equals character length.
	if len(s) < collectionSlugLengthMin || len(s) > collectionSlugLengthMax {
		return s, ErrValidationSlugLength
	}

	if !slugPattern.MatchString(s) {
		return s, ErrValidationSlugInvalidCharacters
	}

	return s, nil
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

var reservedUsernames = map[string]bool{
	// brand
	"url":           true,
	"urlspace":      true,
	"url_space":     true,
	"url-space":     true,
	"urlspaceapp":   true,
	"url-space-app": true,
	"geturlspace":   true,
	// infrastructure
	"root":        true,
	"system":      true,
	"localhost":   true,
	"www":         true,
	"mail":        true,
	"ftp":         true,
	"smtp":        true,
	"email":       true,
	"dns":         true,
	"host":        true,
	"server":      true,
	"internal":    true,
	"dev":         true,
	"development": true,
	"test":        true,
	"testing":     true,
	"staging":     true,
	"stage":       true,
	"prod":        true,
	"production":  true,
	"beta":        true,
	"alpha":       true,
	"demo":        true,
	"example":     true,
	"sandbox":     true,
	"preview":     true,
	// impersonation targets
	"support":        true,
	"help":           true,
	"helpdesk":       true,
	"contact":        true,
	"contact-us":     true,
	"info":           true,
	"security":       true,
	"abuse":          true,
	"postmaster":     true,
	"webmaster":      true,
	"hostmaster":     true,
	"noreply":        true,
	"no-reply":       true,
	"donotreply":     true,
	"do-not-reply":   true,
	"admin":          true,
	"administrator":  true,
	"administration": true,
	"billing":        true,
	"sales":          true,
	"marketing":      true,
	"hello":          true,
	"feedback":       true,
	"report":         true,
	"reports":        true,
	"privacy":        true,
	"legal":          true,
	"compliance":     true,
	"dpo":            true,
	"press":          true,
	// authority
	"moderator":  true,
	"moderators": true,
	"mod":        true,
	"mods":       true,
	"staff":      true,
	"team":       true,
	"official":   true,
	"owner":      true,
	"founder":    true,
	"founders":   true,
	"ceo":        true,
	"superuser":  true,
	"sudo":       true,
	"operator":   true,
	"everyone":   true,
	"all":        true,
	"here":       true,
	// routing / URL conflicts
	"api":            true,
	"graphql":        true,
	"rpc":            true,
	"webhook":        true,
	"webhooks":       true,
	"callback":       true,
	"callbacks":      true,
	"static":         true,
	"assets":         true,
	"cdn":            true,
	"favicon":        true,
	"img":            true,
	"image":          true,
	"images":         true,
	"media":          true,
	"files":          true,
	"uploads":        true,
	"download":       true,
	"downloads":      true,
	"fonts":          true,
	"css":            true,
	"manifest":       true,
	"service-worker": true,
	"rss":            true,
	"feed":           true,
	"feeds":          true,
	"atom":           true,
	"sitemap":        true,
	"sitemaps":       true,
	"public":         true,
	"well-known":     true,
	// site pages, current and planned
	"about":            true,
	"about-us":         true,
	"blog":             true,
	"docs":             true,
	"documentation":    true,
	"doc":              true,
	"guide":            true,
	"guides":           true,
	"learn":            true,
	"tutorials":        true,
	"faq":              true,
	"faqs":             true,
	"changelog":        true,
	"roadmap":          true,
	"status":           true,
	"pricing":          true,
	"plans":            true,
	"plan":             true,
	"upgrade":          true,
	"pro":              true,
	"premium":          true,
	"plus":             true,
	"checkout":         true,
	"subscribe":        true,
	"subscription":     true,
	"subscriptions":    true,
	"donate":           true,
	"sponsor":          true,
	"sponsors":         true,
	"partners":         true,
	"affiliates":       true,
	"careers":          true,
	"jobs":             true,
	"company":          true,
	"brand":            true,
	"news":             true,
	"newsletter":       true,
	"events":           true,
	"community":        true,
	"forum":            true,
	"discord":          true,
	"home":             true,
	"index":            true,
	"welcome":          true,
	"start":            true,
	"onboarding":       true,
	"terms":            true,
	"terms-of-service": true,
	"tos":              true,
	"privacy-policy":   true,
	"cookies":          true,
	"cookie-policy":    true,
	"gdpr":             true,
	"dpa":              true,
	"imprint":          true,
	"impressum":        true,
	"license":          true,
	"licence":          true,
	"licenses":         true,
	"accessibility":    true,
	"trust":            true,
	"apps":             true,
	"app":              true,
	"extension":        true,
	"extensions":       true,
	"integrations":     true,
	"import":           true,
	"export":           true,
	"bookmarklet":      true,
	"embed":            true,
	"widget":           true,
	"share":            true,
	"invite":           true,
	"invites":          true,
	"search":           true,
	"explore":          true,
	"discover":         true,
	"trending":         true,
	"popular":          true,
	"featured":         true,
	"random":           true,
	"latest":           true,
	"top":              true,
	// app features
	"collection":    true,
	"collections":   true,
	"link":          true,
	"links":         true,
	"tag":           true,
	"tags":          true,
	"favourite":     true,
	"favourites":    true,
	"favorite":      true,
	"favorites":     true,
	"for-later":     true,
	"later":         true,
	"lists":         true,
	"list":          true,
	"notifications": true,
	"messages":      true,
	"inbox":         true,
	"new":           true,
	"create":        true,
	"edit":          true,
	"delete":        true,
	"user":          true,
	"users":         true,
	// auth flows
	"auth":            true,
	"oauth":           true,
	"sso":             true,
	"login":           true,
	"logout":          true,
	"signin":          true,
	"sign-in":         true,
	"signout":         true,
	"sign-out":        true,
	"signup":          true,
	"sign-up":         true,
	"register":        true,
	"verify":          true,
	"verify-email":    true,
	"confirm":         true,
	"activate":        true,
	"reset":           true,
	"reset-password":  true,
	"forgot":          true,
	"forgot-password": true,
	"password":        true,
	"2fa":             true,
	"mfa":             true,
	"session":         true,
	"sessions":        true,
	"token":           true,
	"tokens":          true,
	// profile / account paths
	"account":     true,
	"accounts":    true,
	"profile":     true,
	"profiles":    true,
	"settings":    true,
	"preferences": true,
	"dashboard":   true,
	"mine":        true,
	"you":         true,
	// sentinel values
	"null":      true,
	"undefined": true,
	"anonymous": true,
	"unknown":   true,
	"none":      true,
	"nil":       true,
	"true":      true,
	"false":     true,
	"nan":       true,
	// bots / crawlers
	"bot":     true,
	"bots":    true,
	"robot":   true,
	"robots":  true,
	"crawler": true,
	"spider":  true,
	"humans":  true,
}

// Only lowercase ASCII letters, digits, hyphens, and underscores are allowed.
// Consecutive separators (e.g. "a--b", "a__b") are permitted as they are URL-friendly.
var userPattern = regexp.MustCompile(`^[a-z0-9_-]+$`)

const (
	userUsernameLengthMin = 3
	userUsernameLengthMax = 32
)

// validateUsername, reservedUsernames and the username errors are duplicated
// on purpose in the user and collection packages, so neither domain depends on
// the other. Keep both copies identical when changing either one.
func validateUsername(u string) (string, error) {
	u = strings.ToLower(strings.TrimSpace(u))

	if len(u) == 0 {
		return u, ErrValidationUsernameRequired
	}

	// len() is used here instead of utf8.RuneCountInString() because the regex
	// (userPattern) already restricts usernames to single-byte ASCII characters.
	if len(u) < userUsernameLengthMin {
		return u, ErrValidationUsernameTooShort
	}

	if len(u) > userUsernameLengthMax {
		return u, ErrValidationUsernameTooLong
	}

	if !userPattern.MatchString(u) {
		return u, ErrValidationUsernameCharacters
	}

	if strings.HasPrefix(u, "-") || strings.HasPrefix(u, "_") {
		return u, ErrValidationUsernamePrefix
	}

	if strings.HasSuffix(u, "-") || strings.HasSuffix(u, "_") {
		return u, ErrValidationUsernameSuffix
	}

	if reserved := reservedUsernames[u]; reserved {
		return u, ErrValidationUsernameReserved
	}

	return u, nil
}
