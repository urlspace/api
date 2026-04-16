package server

import "net/http"

type middleware func(http.Handler) http.Handler

func middlewareStack(mds ...middleware) middleware {
	return func(next http.Handler) http.Handler {
		for i := len(mds) - 1; i >= 0; i-- {
			next = mds[i](next)
		}
		return next
	}
}
