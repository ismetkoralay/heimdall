package service

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/ismetkoralay/heimdall/internal/auth"
	"github.com/ismetkoralay/heimdall/internal/utils"
	"github.com/stretchr/testify/assert"
)

type MockRepository struct {
	createAPIKey func(ctx context.Context, apiKey auth.APIKey) (auth.APIKey, error)
	captured     auth.APIKey
}

func (r *MockRepository) CreateAPIKey(ctx context.Context, apiKey auth.APIKey) (auth.APIKey, error) {
	r.captured = apiKey
	return r.createAPIKey(ctx, apiKey)
}

func TestGenerateAndSaveAPIKey(t *testing.T) {
	var (
		now    = time.Now().UTC()
		key    = "generated-key"
		apiKey = auth.APIKey{
			ID:              "api-key-id",
			HashedKey:       hashKey(key),
			Name:            "name",
			RPMLimit:        1,
			DailyTokenQuota: 2,
			Revoked:         false,
			CreatedAt:       now,
		}
	)
	tests := []struct {
		name              string
		repository        *MockRepository
		nowFn             func() time.Time
		generateKeyFn     func() (string, error)
		apiKey            auth.APIKey
		expectedHashedKey string
		expectedRes       string
		expectedErr       error
	}{
		{
			name: "successful call",
			repository: &MockRepository{
				createAPIKey: func(ctx context.Context, apiKey auth.APIKey) (auth.APIKey, error) {
					return apiKey, nil
				},
			},
			nowFn: func() time.Time {
				return now
			},
			generateKeyFn: func() (string, error) {
				return key, nil
			},
			apiKey:            apiKey,
			expectedHashedKey: hashKey(key),
			expectedRes:       key,
			expectedErr:       nil,
		},
		{
			name: "fails when generate key fails",
			repository: &MockRepository{
				createAPIKey: func(ctx context.Context, apiKey auth.APIKey) (auth.APIKey, error) {
					return apiKey, nil
				},
			},
			nowFn: func() time.Time {
				return now
			},
			generateKeyFn: func() (string, error) {
				return "", fmt.Errorf("%w: %w", utils.ErrGeneratingKey, errors.New("failed"))
			},
			apiKey:      apiKey,
			expectedErr: utils.ErrGeneratingKey,
		},
		{
			name: "fails when saving to db",
			repository: &MockRepository{
				createAPIKey: func(ctx context.Context, apiKey auth.APIKey) (auth.APIKey, error) {
					return auth.APIKey{}, ErrSavingKey
				},
			},
			nowFn: func() time.Time {
				return now
			},
			generateKeyFn: func() (string, error) {
				return key, nil
			},
			apiKey:      apiKey,
			expectedErr: ErrSavingKey,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			s := New(tt.repository, tt.nowFn, tt.generateKeyFn)

			// Act
			res, err := s.GenerateAndSaveAPIKey(context.Background(), tt.apiKey.Name, tt.apiKey.RPMLimit, tt.apiKey.DailyTokenQuota)

			// Assert
			assert.ErrorIs(t, err, tt.expectedErr)
			assert.Equal(t, tt.expectedRes, res)

			if tt.expectedHashedKey != "" {
				assert.Equal(t, tt.expectedHashedKey, tt.repository.captured.HashedKey)
			}
		})
	}
}
