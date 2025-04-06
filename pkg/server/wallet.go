package server

import (
	"context"
	"fmt"
	"js-centralized-wallet/pkg/trace"
	"js-centralized-wallet/pkg/utils"
	"net/http"
	"strconv"
	"time"
)

const (
	GET_WALLET_CTX_SECONDS         = 10
	GET_WALLET_HISTORY_CTX_SECONDS = 10
)

type GetBalanceResp struct {
	Balance int64 `json:"balance"`
}

func (s *Server) getWalletBalance(w http.ResponseWriter, r *http.Request) {

	ctx, lg := trace.Logger(r.Context())
	getWalletCtx, cancel := context.WithTimeout(ctx, GET_WALLET_CTX_SECONDS*time.Second)
	defer cancel()

	userId, err := utils.GetUserIdFromCtx(getWalletCtx)
	if err != nil {
		respondErr(w, r, err)
		return
	}

	redis := s.model.GetRedis()
	balanceKey := fmt.Sprintf("balance:%d", userId)

	balanceStr, err := redis.Get(ctx, balanceKey).Result()
	if err == nil && balanceStr != "" {
		balance, convErr := strconv.ParseInt(balanceStr, 10, 64)
		if convErr == nil {
			respondJSON(w, r, GetBalanceResp{
				Balance: balance,
			})
			return
		}
	}

	lg.Info("Get wallet balance cache miss, getting from DB")
	balance, err := s.model.GetWalletBalance(getWalletCtx, userId)
	if err != nil {
		respondErr(w, r, err)
		return
	}

	_ = redis.Set(ctx, balanceKey, fmt.Sprintf("%d", balance), 5*time.Minute).Err()

	respondJSON(w, r, GetBalanceResp{
		Balance: balance,
	})
}
