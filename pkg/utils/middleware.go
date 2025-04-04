package utils

import "net/http"

type Middleware func(next http.HandlerFunc) http.HandlerFunc

func ComposeMiddlewares(middlewares ...Middleware) Middleware {
	return func(next http.HandlerFunc) http.HandlerFunc {
		for i := len(middlewares) - 1; i >= 0; i-- {
			next = middlewares[i](next)
		}
		return next
	}
}

func Router(next http.HandlerFunc) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/", next)
	return mux
}
