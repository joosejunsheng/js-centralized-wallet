package utils

import (
	"net/http"
	"strconv"
	"time"
)

func AllowAllOrigins(next http.HandlerFunc) http.HandlerFunc {
	ttl := strconv.Itoa(int(24 * time.Hour.Seconds()))

	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		if r.Method == http.MethodOptions {
			w.Header().Set("Access-Control-Allow-Methods", "*")
			w.Header().Set("Access-Control-Allow-Headers", "*")
			w.Header().Set("Access-Control-Max-Age", ttl)
			w.WriteHeader(http.StatusOK)
			return
		}

		next(w, r)
	}
}
