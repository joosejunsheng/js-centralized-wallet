package server

import (
	"context"
	"encoding/json"
	"fmt"
	"js-centralized-wallet/pkg/model"
	"js-centralized-wallet/pkg/trace"
	"js-centralized-wallet/pkg/utils"
	"net/http"
	"time"
)

const (
	TRANSFER_BALANCE_CTX_SECONDS       = 15
	DEPOSIT_CTX_SECONDS                = 10
	WITHDRAW_CTX_SECONDS               = 10
	GET_TRANSCITON_HISTORY_CTX_SECONDS = 10
)

type TransactionHistoryReq struct {
	TransactionType int `json:"type"`
	Page            int `json:"page"`
	PageSize        int `json:"page_size"`
}

type TransactionItem struct {
	TransactionUUID       string                `json:"transaction_uuid"`
	Amount                int64                 `json:"amount"`
	TransactionType       model.TransactionType `json:"type"`
	TransactionTypeString string                `json:"transaction_type"`
	Description           string                `json:"desc"`
	CreatedAt             time.Time             `json:"created_at"`
}

type TransactionHistoryResp struct {
	StatementBalance int64             `json:"statement_balance"`
	Transactions     []TransactionItem `json:"transactions"`
}

type TransferBalanceReq struct {
	DestinationUserId uint64 `json:"destination_user_id"`
	Amount            int64  `json:"amount"`
}

type TransferBalanceResp struct {
	Success bool `json:"success"`
}

type DepositReq struct {
	Amount int64 `json:"amount"`
}

type DepositResp struct {
	Balance int64 `json:"balance"`
}

type WithdrawReq struct {
	Amount int64 `json:"amount"`
}

type WithdrawResp struct {
	Balance int64 `json:"balance"`
}

func (s *Server) getTransactionHistory(w http.ResponseWriter, r *http.Request) {

	ctx, lg := trace.Logger(r.Context())
	getTransactionHistoryCtx, cancel := context.WithTimeout(ctx, GET_WALLET_CTX_SECONDS*time.Second)
	defer cancel()

	userId, err := utils.GetUserIdFromCtx(getTransactionHistoryCtx)
	if err != nil {
		respondErr(w, r, err)
		return
	}

	q := r.URL.Query()
	transactionType := utils.GetQueryInt(q, "type", 0)
	page := utils.GetQueryInt(q, "page", 1)
	pageSize := utils.GetQueryInt(q, "page_size", 30)

	redis := s.model.GetRedis()
	transactionHistoryKey := fmt.Sprintf("transaction_history:%d-%d-%d-%d", userId, transactionType, page, pageSize)

	historyStr, err := redis.Get(ctx, transactionHistoryKey).Result()
	if err == nil && historyStr != "" {
		var resp TransactionHistoryResp
		err = json.Unmarshal([]byte(historyStr), &resp)
		if err == nil {
			respondJSON(w, r, resp)
			return
		}
	}

	lg.Info("Get transaction history cache miss, getting from DB")

	transactions, err := s.model.GetTransactionHistory(getTransactionHistoryCtx, userId, transactionType, model.PageInfo{
		Page:     page,
		PageSize: pageSize,
	})
	if err != nil {
		respondErr(w, r, err)
		return
	}

	transactionResp := make([]TransactionItem, len(transactions))
	var filteredBalance int64

	for i, transaction := range transactions {

		transactionResp[i] = TransactionItem{
			TransactionUUID:       transaction.TransactionUUID,
			Amount:                transaction.Amount,
			TransactionType:       transaction.Type,
			TransactionTypeString: transaction.Type.String(),
			CreatedAt:             transaction.CreatedAt,
		}

		if transaction.Amount > 0 {
			transactionResp[i].Description = fmt.Sprintf("Received $%d", transaction.Amount)
		} else {
			transactionResp[i].Description = fmt.Sprintf("Sent $%d", transaction.Amount*-1)
		}

		filteredBalance += transaction.Amount
	}

	resp := TransactionHistoryResp{
		Transactions:     transactionResp,
		StatementBalance: filteredBalance,
	}

	redisData, err := json.Marshal(resp)
	if err != nil {
		lg.Error("failed to marshal resp into redis", "error", err)
	} else {
		_ = redis.Set(ctx, transactionHistoryKey, redisData, 5*time.Minute).Err()
	}

	respondJSON(w, r, resp)
}

func (s *Server) deposit(w http.ResponseWriter, r *http.Request) {

	ctx, _ := trace.Logger(r.Context())
	depositCtx, cancel := context.WithTimeout(ctx, DEPOSIT_CTX_SECONDS*time.Second)
	defer cancel()

	userId, err := utils.GetUserIdFromCtx(depositCtx)
	if err != nil {
		respondErr(w, r, err)
		return
	}

	var req DepositReq
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		respondErr(w, r, model.ErrBadInput)
		return
	}

	if req.Amount == 0 {
		respondErr(w, r, model.ErrInvalidAmount)
		return
	}

	newBalance, err := s.model.Deposit(depositCtx, userId, req.Amount)
	if err != nil {
		respondErr(w, r, err)
		return
	}

	s.model.InvalidateWalletCache(depositCtx, userId)

	respondJSON(w, r, DepositResp{
		Balance: newBalance,
	})
}

func (s *Server) withdraw(w http.ResponseWriter, r *http.Request) {

	ctx, _ := trace.Logger(r.Context())
	withdrawCtx, cancel := context.WithTimeout(ctx, WITHDRAW_CTX_SECONDS*time.Second)
	defer cancel()

	userId, err := utils.GetUserIdFromCtx(withdrawCtx)
	if err != nil {
		respondErr(w, r, err)
		return
	}

	var req DepositReq
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		respondErr(w, r, model.ErrBadInput)
		return
	}

	if req.Amount == 0 {
		respondErr(w, r, model.ErrInvalidAmount)
		return
	}

	newBalance, err := s.model.Withdraw(withdrawCtx, userId, req.Amount)
	if err != nil {
		respondErr(w, r, err)
		return
	}

	s.model.InvalidateWalletCache(withdrawCtx, userId)

	respondJSON(w, r, WithdrawResp{
		Balance: newBalance,
	})
}

func (s *Server) transferBalance(w http.ResponseWriter, r *http.Request) {

	ctx, _ := trace.Logger(r.Context())
	transferBalanceCtx, cancel := context.WithTimeout(ctx, TRANSFER_BALANCE_CTX_SECONDS*time.Second)
	defer cancel()

	userId, err := utils.GetUserIdFromCtx(transferBalanceCtx)
	if err != nil {
		respondErr(w, r, err)
		return
	}

	var req TransferBalanceReq
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		respondErr(w, r, model.ErrBadInput)
		return
	}

	if req.Amount == 0 {
		respondErr(w, r, model.ErrInvalidAmount)
		return
	}

	// Self transfer prohibited
	if userId == req.DestinationUserId {
		respondErr(w, r, model.ErrSelfTransferInvalid)
		return
	}

	err = s.model.TransferBalance(transferBalanceCtx, userId, req.DestinationUserId, req.Amount)
	if err != nil {
		respondErr(w, r, err)
		return
	}

	s.model.InvalidateWalletCache(transferBalanceCtx, userId, req.DestinationUserId)

	respondJSON(w, r, TransferBalanceResp{
		Success: err == nil,
	})
}
