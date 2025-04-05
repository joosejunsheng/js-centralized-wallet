package middlewares_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"js-centralized-wallet/pkg/utils/middlewares"

	"github.com/go-redis/redismock/v8"
	"github.com/stretchr/testify/assert"
)

func TestThrottleMiddleware(t *testing.T) {
	rdb, mock := redismock.NewClientMock()

	handler := middlewares.ThrottleMiddleware(rdb, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	t.Run("Exceed rate limit", func(t *testing.T) {
		mock.ExpectGet("rate_limit:127.0.0.1").SetVal("10")

		req := httptest.NewRequest("GET", "/api/ping/v1", nil)
		req.RemoteAddr = "127.0.0.1"
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		t.Log("Res:", rr.Body)
		t.Log("Res:", rr.Code)
		t.Log("Expected Code: 429")

		assert.Equal(t, http.StatusTooManyRequests, rr.Code)
	})

	t.Run("Does not exceed rate limit", func(t *testing.T) {
		mock.ExpectGet("rate_limit:127.0.0.1").SetVal("3")

		mock.ExpectIncr("rate_limit:127.0.0.1").SetVal(5)
		mock.ExpectExpire("rate_limit:127.0.0.1", 10*time.Second).SetVal(true)

		req := httptest.NewRequest("GET", "/api/ping/v1", nil)
		req.RemoteAddr = "127.0.0.1"
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		t.Log("Res:", rr.Body)
		t.Log("Res:", rr.Code)
		t.Log("Expected Code: 200")

		assert.Equal(t, http.StatusOK, rr.Code)
	})
}
