package server

import (
	"context"
	"js-centralized-wallet/pkg/trace"
	"js-centralized-wallet/pkg/utils"
	"net/http"
	"time"
)

const (
	GET_WALLET_CTX_SECONDS         = 5
	GET_WALLET_HISTORY_CTX_SECONDS = 5
)

func (s *Server) getWalletBalance(w http.ResponseWriter, r *http.Request) {

	ctx, _ := trace.Logger(r.Context())
	getWalletCtx, cancel := context.WithTimeout(ctx, GET_WALLET_CTX_SECONDS*time.Second)
	defer cancel()

	userId, err := utils.GetUserIdFromCtx(getWalletCtx)
	if err != nil {
		respondErr(w, r, err)
		return
	}

	users, err := s.model.GetWalletBalance(getWalletCtx, userId)
	if err != nil {
		respondErr(w, r, err)
		return
	}

	respondJSON(w, r, users)
}
