package server

import (
	"js-centralized-wallet/pkg/model"
	"net/http"
)

func (s *Server) getAllUsers(w http.ResponseWriter, r *http.Request) {
	users, err := s.model.GetAllUsers(r.Context(), model.PageInfo{
		Page:     1,
		PageSize: 10,
	})
	if err != nil {
		respondErr(w, r, err)
		return
	}

	respondJSON(w, r, users)
}
