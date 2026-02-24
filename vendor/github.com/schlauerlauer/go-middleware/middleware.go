package middleware

import "net/http"

type Middleware func(http.Handler) http.Handler

func CreateStack(middlewareStack ...Middleware) Middleware {
	return func(next http.Handler) http.Handler {
		for i := len(middlewareStack) - 1; i >= 0; i-- {
			middleware := middlewareStack[i]
			next = middleware(next)
		}
		return next
	}
}
