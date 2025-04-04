package model

type Wallet struct {
	Base
	UserId  uint64  `json:"user_id"`
	Balance float64 `json:"balance"`
}
