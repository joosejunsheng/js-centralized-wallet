package server

import (
	"context"
	"encoding/json"
	"js-centralized-wallet/pkg/model"
	"js-centralized-wallet/pkg/trace"
	"js-centralized-wallet/pkg/utils"
	"net/http"
	"time"
)

func (s *Server) transferBalanceV2(w http.ResponseWriter, r *http.Request) {

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

	sourceBalance, err := s.model.GetWalletBalance(ctx, userId)
	if err != nil {
		respondErr(w, r, err)
		return
	}

	if sourceBalance < req.Amount {
		respondErr(w, r, model.ErrBalanceInsufficient)
		return
	}

	s.jobChan <- model.TransferJob{
		Ctx:          transferBalanceCtx,
		SourceUserId: userId,
		DestUserId:   req.DestinationUserId,
		Amount:       req.Amount,
	}

	respondJSON(w, r, TransferBalanceResp{
		Success: err == nil,
	})
}
