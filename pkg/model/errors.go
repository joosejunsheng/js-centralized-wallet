package model

import (
	"fmt"
)

// For http 4xx client errors
type ClientError struct {
	Code string
}

var _ = error(&ClientError{})

func (e *ClientError) Error() string {
	return fmt.Sprintf("[%s]", e.Code)
}

func newClientError(code string) error {
	return fmt.Errorf("client error: %w", &ClientError{code})
}

var (
	ErrBadInput            = newClientError("bad_input")
	ErrInvalidAmount       = newClientError("invalid_amount")
	ErrUnauthorized        = newClientError("unauthorized")
	ErrBalanceInsufficient = newClientError("balance_insufficient")
	ErrSelfTransferInvalid = newClientError("self_transfer_invalid")
)
