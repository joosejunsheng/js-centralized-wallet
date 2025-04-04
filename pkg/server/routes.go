package server

import (
	"js-centralized-wallet/pkg/utils"
	"net/http"
)

func (s *Server) apiRoutes(next http.HandlerFunc) http.HandlerFunc {
	r := utils.Router(next)
	r.HandleFunc("GET /api/ping/v1", s.ping)

	{ // Users
		r.HandleFunc("GET /api/users/v1", s.getAllUsers)
	}

	return r.ServeHTTP
}
