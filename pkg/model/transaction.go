package model

type Transaction struct {
	Base
	UserId   uint64  `json:"user_id"`
	WalletId uint64  `json:"wallet_id"`
	Amount   float64 `json:"amount"`
}
