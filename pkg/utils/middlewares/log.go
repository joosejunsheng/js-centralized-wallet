package middlewares

import (
	"js-centralized-wallet/pkg/trace"
	"net/http"
	"time"
)

func AccessLog(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, lg := trace.Logger(r.Context())

		start := time.Now()

		next(w, r.WithContext(ctx))

		lg.Info("access log",
			"method", r.Method,
			"account", r.Header.Get("Account"),
			"res-encoding", w.Header().Get("Content-Encoding"),
			"url", r.URL.String(),
			"duration", time.Since(start).String(),
		)
	}
}
