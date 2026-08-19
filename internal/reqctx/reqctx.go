package reqctx

import (
	"context"
	"errors"
)

var (
	ErrAPIKeyIDMissing              = errors.New("api key id is missing in context")
	ErrAPIKeyRPMLimitMissing        = errors.New("api key rpm limit is missing in context")
	ErrAPIKeyDailyTokenQuotaMissing = errors.New("api key daily token quota is missing in context")
)

type apiKeyIDContextKey struct{}

func SetAPIKeyID(ctx context.Context, ID string) context.Context {
	return context.WithValue(ctx, apiKeyIDContextKey{}, ID)
}

func GetAPIKeyID(ctx context.Context) (string, error) {
	res := ctx.Value(apiKeyIDContextKey{})
	if res == nil {
		return "", ErrAPIKeyIDMissing
	}

	return res.(string), nil
}

type apiKeyRPMLimitContextKey struct{}

func SetAPIKeyRPMLimit(ctx context.Context, rpmLimit int) context.Context {
	return context.WithValue(ctx, apiKeyRPMLimitContextKey{}, rpmLimit)
}

func GetAPIKeyRPMLimit(ctx context.Context) (int, error) {
	res := ctx.Value(apiKeyRPMLimitContextKey{})
	if res == nil {
		return 0, ErrAPIKeyRPMLimitMissing
	}

	return res.(int), nil
}

type apiKeyDailyTokenQuotaContextKey struct{}

func SetAPIKeyDailyTokenQuota(ctx context.Context, dailyTokenLimit int) context.Context {
	return context.WithValue(ctx, apiKeyDailyTokenQuotaContextKey{}, dailyTokenLimit)
}

func GetAPIKeyDailyTokenQuota(ctx context.Context) (int, error) {
	res := ctx.Value(apiKeyDailyTokenQuotaContextKey{})
	if res == nil {
		return 0, ErrAPIKeyDailyTokenQuotaMissing
	}

	return res.(int), nil
}
