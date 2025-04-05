package server

import (
	"js-centralized-wallet/pkg/utils/middlewares"
	"net/http"
)

func (s *Server) apiRoutes(next http.HandlerFunc) http.HandlerFunc {
	r := middlewares.Router(next)
	r.HandleFunc("GET /api/ping/v1", middlewares.ThrottleMiddleware(s.model.GetRedis(), s.ping))

	{ // Test get all users
		r.HandleFunc("GET /api/users/v1", s.getAllUsers)
	}

	{
		r.HandleFunc("GET /api/wallet/balance/v1", middlewares.AuthMiddleware(s.getWalletBalance))
		r.HandleFunc("GET /api/transactions/v1", middlewares.AuthMiddleware(s.getTransactionHistory))
		r.HandleFunc("POST /api/deposit/v1", middlewares.AuthMiddleware(s.deposit))
		r.HandleFunc("POST /api/withdraw/v1", middlewares.AuthMiddleware(s.withdraw))
		r.HandleFunc("POST /api/transfer/v1", middlewares.ThrottleMiddleware(s.model.GetRedis(), middlewares.AuthMiddleware(s.transferBalance)))
	}

	return r.ServeHTTP
}
