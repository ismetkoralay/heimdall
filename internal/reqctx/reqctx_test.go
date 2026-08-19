package reqctx

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAPIKeyIDContext(t *testing.T) {
	ctx := context.Background()

	res, err := GetAPIKeyID(ctx)
	assert.Empty(t, res)
	assert.ErrorIs(t, err, ErrAPIKeyIDMissing)

	ctx = SetAPIKeyID(ctx, "test-id")

	res, err = GetAPIKeyID(ctx)
	assert.Equal(t, "test-id", res)
	assert.NoError(t, err)
}

func TestAPIKeyRPMLimitContext(t *testing.T) {
	ctx := context.Background()

	res, err := GetAPIKeyRPMLimit(ctx)
	assert.Equal(t, 0, res)
	assert.ErrorIs(t, err, ErrAPIKeyRPMLimitMissing)

	ctx = SetAPIKeyRPMLimit(ctx, 10)

	res, err = GetAPIKeyRPMLimit(ctx)
	assert.Equal(t, 10, res)
	assert.NoError(t, err)
}

func TestAPIKeyDailyTokenQuotaContext(t *testing.T) {
	ctx := context.Background()

	res, err := GetAPIKeyDailyTokenQuota(ctx)
	assert.Equal(t, 0, res)
	assert.ErrorIs(t, err, ErrAPIKeyDailyTokenQuotaMissing)

	ctx = SetAPIKeyDailyTokenQuota(ctx, 20)

	res, err = GetAPIKeyDailyTokenQuota(ctx)
	assert.Equal(t, 20, res)
	assert.NoError(t, err)
}
