package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/ismetkoralay/heimdall/internal/auth"
	"github.com/ismetkoralay/heimdall/internal/utils"
	"github.com/stretchr/testify/assert"
)

type MockRepository struct {
	createAPIKey         func(ctx context.Context, apiKey auth.APIKey) (auth.APIKey, error)
	getAPIKeyByHashedKey func(ctx context.Context, hashedKey string) (auth.APIKey, error)
	captured             auth.APIKey
	capturedHashedKey    string
}

func (r *MockRepository) CreateAPIKey(ctx context.Context, apiKey auth.APIKey) (auth.APIKey, error) {
	r.captured = apiKey
	return r.createAPIKey(ctx, apiKey)
}

func (r *MockRepository) GetAPIKeyByHashedKey(ctx context.Context, hashedKey string) (auth.APIKey, error) {
	r.capturedHashedKey = hashedKey
	return r.getAPIKeyByHashedKey(ctx, hashedKey)
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

func TestValidateAPIKey(t *testing.T) {
	var (
		now    = time.Now().UTC()
		key    = "hd_lpOhZAmsYy8ag76ftCAhfisnkgtakkQh4zK_wR57UQg"
		apiKey = auth.APIKey{
			ID:              "api-key-id",
			HashedKey:       hashKey(key),
			Name:            "name",
			RPMLimit:        1,
			DailyTokenQuota: 2,
			Revoked:         false,
			CreatedAt:       now,
		}
		revokedAPIKey = auth.APIKey{
			ID:              "api-key-id",
			HashedKey:       hashKey(key),
			Name:            "name",
			RPMLimit:        1,
			DailyTokenQuota: 2,
			Revoked:         true,
			CreatedAt:       now,
		}
	)
	tests := []struct {
		name              string
		repository        *MockRepository
		nowFn             func() time.Time
		requestedKey      string
		expectedHashedKey string
		expectedRes       auth.APIKey
		expectedErr       error
		expectedAuthErr   *auth.AuthError
	}{
		{
			name: "successful call",
			repository: &MockRepository{
				getAPIKeyByHashedKey: func(ctx context.Context, hashedKey string) (auth.APIKey, error) {
					return apiKey, nil
				},
			},
			nowFn: func() time.Time {
				return now
			},
			requestedKey:      key,
			expectedHashedKey: hashKey(key),
			expectedRes:       apiKey,
		},
		{
			name: "fails when key is not found",
			repository: &MockRepository{
				getAPIKeyByHashedKey: func(ctx context.Context, key string) (auth.APIKey, error) {
					return auth.APIKey{}, &auth.AuthError{
						StatusCode: http.StatusNotFound,
						Err:        errors.New("key not found"),
					}
				},
			},
			nowFn: func() time.Time {
				return now
			},
			requestedKey:      "non-existing",
			expectedHashedKey: hashKey("non-existing"),
			expectedRes:       auth.APIKey{},
			expectedErr:       errors.New("key not found"),
			expectedAuthErr: &auth.AuthError{
				StatusCode: http.StatusNotFound,
				Err:        errors.New("key not found"),
			},
		},
		{
			name: "fails when repo fails with unexpected error",
			repository: &MockRepository{
				getAPIKeyByHashedKey: func(ctx context.Context, key string) (auth.APIKey, error) {
					return auth.APIKey{}, errors.New("unexpected error")
				},
			},
			nowFn: func() time.Time {
				return now
			},
			requestedKey:      key,
			expectedHashedKey: hashKey(key),
			expectedRes:       auth.APIKey{},
			expectedErr:       errors.New("unexpected error"),
		},
		{
			name: "fails when the key is revoked",
			repository: &MockRepository{
				getAPIKeyByHashedKey: func(ctx context.Context, key string) (auth.APIKey, error) {
					return revokedAPIKey, nil
				},
			},
			nowFn: func() time.Time {
				return now
			},
			requestedKey:      key,
			expectedHashedKey: hashKey(key),
			expectedRes:       auth.APIKey{},
			expectedErr:       errors.New("invalid token"),
			expectedAuthErr: &auth.AuthError{
				StatusCode: http.StatusUnauthorized,
				Err:        errors.New("invalid token"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			s := New(tt.repository, tt.nowFn, func() (string, error) {
				return key, nil
			})

			// Act
			res, err := s.ValidateAPIKey(context.Background(), tt.requestedKey)

			// Assert
			assert.Equal(t, tt.expectedHashedKey, tt.repository.capturedHashedKey)
			assert.Equal(t, tt.expectedRes, res)

			if tt.expectedErr != nil {
				assert.ErrorContains(t, err, tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
			}

			if tt.expectedAuthErr != nil {
				var authErr *auth.AuthError
				assert.ErrorAs(t, err, &authErr)
				assert.Equal(t, tt.expectedAuthErr.StatusCode, authErr.StatusCode)
				assert.ErrorContains(t, authErr.Err, tt.expectedAuthErr.Unwrap().Error())
			}
		})
	}
}
