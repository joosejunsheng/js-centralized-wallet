package utils

import (
	"context"
	"js-centralized-wallet/pkg/model"
	"js-centralized-wallet/pkg/utils/middlewares"
)

func GetUserIdFromCtx(ctx context.Context) (uint64, error) {
	val := ctx.Value(middlewares.USER_ID_KEY)
	userId, ok := val.(uint64)
	if !ok {
		return 0, model.ErrUnauthorized
	}
	return userId, nil
}
