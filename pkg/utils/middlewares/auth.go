package middlewares

import (
	"context"
	"net/http"
	"strconv"
)

type ctxKey string

const (
	USER_ID_KEY ctxKey = "userId"
)

func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userIdStr := r.Header.Get("Authorization")
		userId, err := strconv.ParseUint(userIdStr, 10, 64)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), USER_ID_KEY, userId)
		next(w, r.WithContext(ctx))
	}
}
