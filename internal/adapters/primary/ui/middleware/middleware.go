package middleware

import "net/http"

type Middleware func(http.Handler) http.Handler

func NewChainMiddleware(middlewar ...Middleware) Middleware {
	return func(next http.Handler) http.Handler {
		for i := len(middlewar) - 1; i >= 0; i-- {
			next = middlewar[i](next)
		}
		return next
	}
}
