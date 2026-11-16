package subscription

// Website paths the checkout's back button returns to, keyed by the place the
// checkout was started from. The client only sends the key, so it can never
// point the checkout at another site.
var returnPaths = map[string]string{
	"home":      "/#pricing",
	"dashboard": "/dashboard",
	"settings":  "/settings",
}

func validateReturnTo(returnTo string) (string, error) {
	if returnTo == "" {
		return "", ErrValidationReturnToRequired
	}

	path, ok := returnPaths[returnTo]
	if !ok {
		return "", ErrValidationReturnToInvalid
	}

	return path, nil
}
