package user

import (
	"encoding/base64"
	"net/mail"
	"regexp"
	"strings"
	"unicode/utf8"
)

func validateEmail(e string) (string, error) {
	e = strings.ToLower(strings.TrimSpace(e))

	if len(e) == 0 {
		return e, ErrValidationEmailRequired
	}

	// RFC 5321 limits the total length of an email address to 254 characters.
	if len(e) > 254 {
		return e, ErrValidationEmailTooLong
	}

	parsed, err := mail.ParseAddress(e)
	if err != nil {
		return e, ErrValidationEmailFormat
	}

	// mail.ParseAddress accepts RFC 5322 display names like "Joe" <joe@evil.com>.
	// Reject these by ensuring the raw input matches the parsed address exactly.
	if parsed.Address != e {
		return e, ErrValidationEmailFormat
	}

	parts := strings.SplitN(e, "@", 2)

	// RFC 5322 permits TLD-less domains like "user@mail" or "user@localhost",
	// but a public SaaS can never deliver a verification email to such an
	// address, so reject anything without at least one dot in the domain and
	// a non-empty TLD. The LastIndex check rejects "user@mail" (no dot) and
	// "user@example." (empty TLD); leading-dot cases like "user@.com" are
	// already rejected by mail.ParseAddress above (empty domain label).
	domain := parts[1]
	if i := strings.LastIndex(domain, "."); i <= 0 || i == len(domain)-1 {
		return e, ErrValidationEmailFormat
	}

	// Strip plus-addressing (subaddressing) from the local part to prevent
	// users from creating multiple accounts with the same mailbox.
	// e.g. "user+tag@gmail.com" becomes "user@gmail.com".
	local := strings.SplitN(parts[0], "+", 2)
	e = local[0] + "@" + parts[1]

	// An email like "+tag@gmail.com" is valid per RFC 5322, but after stripping
	// the plus-addressing above the local part becomes empty, producing "@gmail.com".
	// Reject this to avoid storing an invalid email address.
	if len(local[0]) == 0 {
		return e, ErrValidationEmailFormat
	}

	// RFC 5321 §4.5.3.1 limits the local part (before @) to 64 characters.
	if len(local[0]) > 64 {
		return e, ErrValidationEmailFormat
	}

	return e, nil
}

const (
	passwordLengthMin = 12
	// 128 characters is generous enough to accommodate the maximum generated
	// password length of popular password managers (1Password: 100, LastPass: 100,
	// Bitwarden: 128, Dashlane: 40, KeePass: unlimited). An upper bound is
	// necessary because Argon2id is deliberately memory-intensive (64 MB per hash),
	// so accepting unbounded input would allow attackers to tie up server resources
	// with a small number of concurrent requests containing very large passwords.
	passwordLengthMax = 128
)

// minContextualLen is the minimum length of a contextual value (username,
// email local-part, etc.) before it's checked against a candidate password.
// Shorter values produce too many false positives — e.g. a 3-char username
// "joe" would reject a strong password that happens to contain "joe".
const minContextualLen = 4

func validatePassword(p string, contextualValues ...string) (string, error) {
	if len(p) == 0 {
		return p, ErrValidationPasswordRequired
	}

	if strings.TrimSpace(p) == "" {
		return p, ErrValidationPasswordRequired
	}

	// Use RuneCountInString instead of len to count human-readable characters,
	// not bytes. Multi-byte characters like emoji (4 bytes) or CJK (3 bytes)
	// would inflate the byte count and pass a len() check with very few actual
	// characters of entropy (e.g. 3 emoji = 12 bytes but only 3 characters).
	if utf8.RuneCountInString(p) < passwordLengthMin {
		return p, ErrValidationPasswordTooShort
	}

	if utf8.RuneCountInString(p) > passwordLengthMax {
		return p, ErrValidationPasswordTooLong
	}

	// Reject passwords that contain a known contextual value (username,
	// display name, email local-part). Catches the common pattern of users
	// recycling their identity into their password.
	if len(contextualValues) == 0 {
		return p, nil
	}
	pLower := strings.ToLower(p)
	for _, v := range contextualValues {
		v = strings.ToLower(strings.TrimSpace(v))
		if utf8.RuneCountInString(v) < minContextualLen {
			continue
		}
		if strings.Contains(pLower, v) {
			return p, ErrValidationPasswordContainsContext
		}
	}

	return p, nil
}

// emailTokenBase64Length is the length of a verification or password-reset
// token after base64url-encoding 32 random bytes with no padding
// (ceil(32 * 4 / 3) = 43). Tokens always have exactly this length; anything
// else is malformed and we can reject it without a DB lookup.
const emailTokenBase64Length = 43

func validateToken(token string) (string, error) {
	if len(token) == 0 {
		return token, ErrValidationTokenRequired
	}

	if len(token) != emailTokenBase64Length {
		return token, ErrValidationTokenFormat
	}

	if _, err := base64.RawURLEncoding.DecodeString(token); err != nil {
		return token, ErrValidationTokenFormat
	}

	return token, nil
}

var emailChangeCodePattern = regexp.MustCompile(`^[0-9]{6}$`)

func validateEmailChangeCode(c string) (string, error) {
	c = strings.TrimSpace(c)

	if len(c) == 0 {
		return c, ErrValidationEmailChangeCodeRequired
	}

	if !emailChangeCodePattern.MatchString(c) {
		return c, ErrValidationEmailChangeCodeFormat
	}

	return c, nil
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

const displayNameLengthMax = 50

func validateDisplayName(d string) (string, error) {
	d = strings.TrimSpace(d)

	if len(d) == 0 {
		return d, ErrValidationDisplayNameRequired
	}

	if utf8.RuneCountInString(d) > displayNameLengthMax {
		return d, ErrValidationDisplayNameTooLong
	}

	return d, nil
}

const tokenDescriptionLengthMax = 255

func validateTokenDescription(d string) (string, error) {
	d = strings.TrimSpace(d)

	if len(d) == 0 {
		return d, ErrValidationTokenDescriptionRequired
	}

	if utf8.RuneCountInString(d) > tokenDescriptionLengthMax {
		return d, ErrValidationTokenDescriptionTooLong
	}

	return d, nil
}

func validateIsAdmin(isAdmin *bool) (bool, error) {
	if isAdmin == nil {
		return false, ErrValidationIsAdminRequired
	}

	return *isAdmin, nil
}

func validateIsPro(isPro *bool) (bool, error) {
	if isPro == nil {
		return false, ErrValidationIsProRequired
	}

	return *isPro, nil
}
