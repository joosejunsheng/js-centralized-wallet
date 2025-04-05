package middlewares

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/go-redis/redis/v8"
)

func ThrottleMiddleware(redis *redis.Client, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ip := r.RemoteAddr

		limit := 5
		window := 10 * time.Second

		ctx := context.Background()
		key := fmt.Sprintf("rate_limit:%s", ip)

		count, err := redis.Get(ctx, key).Int64()
		if err != nil {
			_, err = redis.Set(ctx, key, 0, window).Result()
			if err != nil {
				http.Error(w, "Error creating counter", http.StatusInternalServerError)
				return
			}
			count = 0
		}

		if count >= int64(limit) {
			http.Error(w, "Rate limited", http.StatusTooManyRequests)
			return
		}

		_, err = redis.Incr(ctx, key).Result()
		if err != nil {
			http.Error(w, "Error increasing counter", http.StatusInternalServerError)
			return
		}

		_, err = redis.Expire(ctx, key, window).Result()
		if err != nil {
			http.Error(w, "Error setting expiration for rate limit", http.StatusInternalServerError)
			return
		}

		next(w, r)
	}
}
