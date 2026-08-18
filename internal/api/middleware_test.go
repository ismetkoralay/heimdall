package api

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ismetkoralay/heimdall/internal/auth"
	"github.com/stretchr/testify/assert"
)

type mockAuthService struct {
	validateAPIKey func(ctx context.Context, key string) (auth.APIKey, error)
}

func (m *mockAuthService) ValidateAPIKey(ctx context.Context, key string) (auth.APIKey, error) {
	return m.validateAPIKey(ctx, key)
}

func TestAuthMiddleware(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	fixedCreatedAt := time.Date(2026, time.January, 1, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name                    string
		authHeader              string
		authService             AuthService
		expectedStatusCode      int
		expectedBody            string
		expectNextCalled        bool
		expectedAPIKeyID        string
		expectedRPMLimit        int
		expectedDailyTokenQuota int
		expectedCreatedAt       time.Time
	}{
		{
			name:       "missing authorization header",
			authHeader: "",
			authService: &mockAuthService{
				validateAPIKey: func(ctx context.Context, key string) (auth.APIKey, error) {
					t.Fatal("ValidateAPIKey should not have been called")
					return auth.APIKey{}, nil
				},
			},
			expectedStatusCode: http.StatusUnauthorized,
			expectedBody:       "invalid token\n",
		},
		{
			name:       "authorization header missing bearer prefix",
			authHeader: "Basic dXNlcjpwYXNz",
			authService: &mockAuthService{
				validateAPIKey: func(ctx context.Context, key string) (auth.APIKey, error) {
					t.Fatal("ValidateAPIKey should not have been called")
					return auth.APIKey{}, nil
				},
			},
			expectedStatusCode: http.StatusUnauthorized,
			expectedBody:       "invalid token\n",
		},
		{
			name:       "authorization header malformed - has only prefix",
			authHeader: "Bearerxxx",
			authService: &mockAuthService{
				validateAPIKey: func(ctx context.Context, key string) (auth.APIKey, error) {
					t.Fatal("ValidateAPIKey should not have been called")
					return auth.APIKey{}, nil
				},
			},
			expectedStatusCode: http.StatusUnauthorized,
			expectedBody:       "invalid token\n",
		},
		{
			name:       "authorization header malformed - has prefix but no rest",
			authHeader: "Bearer",
			authService: &mockAuthService{
				validateAPIKey: func(ctx context.Context, key string) (auth.APIKey, error) {
					t.Fatal("ValidateAPIKey should not have been called")
					return auth.APIKey{}, nil
				},
			},
			expectedStatusCode: http.StatusUnauthorized,
			expectedBody:       "invalid token\n",
		},
		{
			name:       "auth service returns AuthError",
			authHeader: "Bearer valid-token",
			authService: &mockAuthService{
				validateAPIKey: func(ctx context.Context, key string) (auth.APIKey, error) {
					assert.Equal(t, "valid-token", key)
					return auth.APIKey{}, &auth.AuthError{
						StatusCode: http.StatusUnauthorized,
						Err:        errors.New("invalid api key"),
					}
				},
			},
			expectedStatusCode: http.StatusUnauthorized,
			expectedBody:       "invalid api key\n",
		},
		{
			name:       "auth service returns unexpected error",
			authHeader: "Bearer valid-token",
			authService: &mockAuthService{
				validateAPIKey: func(ctx context.Context, key string) (auth.APIKey, error) {
					return auth.APIKey{}, errors.New("database unavailable")
				},
			},
			expectedStatusCode: http.StatusInternalServerError,
			expectedBody:       "something went wrong\n",
		},
		{
			name:       "valid token calls next handler with populated context",
			authHeader: "Bearer valid-token",
			authService: &mockAuthService{
				validateAPIKey: func(ctx context.Context, key string) (auth.APIKey, error) {
					assert.Equal(t, "valid-token", key)
					return auth.APIKey{
						ID:              "key-1",
						RPMLimit:        60,
						DailyTokenQuota: 1000,
						CreatedAt:       fixedCreatedAt,
					}, nil
				},
			},
			expectedStatusCode:      http.StatusOK,
			expectedBody:            "ok",
			expectNextCalled:        true,
			expectedAPIKeyID:        "key-1",
			expectedRPMLimit:        60,
			expectedDailyTokenQuota: 1000,
			expectedCreatedAt:       fixedCreatedAt,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			nextCalled := false
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				nextCalled = true

				assert.Equal(t, tt.expectedAPIKeyID, r.Context().Value(apiKeyIDContextKey))
				assert.Equal(t, tt.expectedRPMLimit, r.Context().Value(apiKeyRPMLimitContextKey))
				assert.Equal(t, tt.expectedDailyTokenQuota, r.Context().Value(apiKeyDailyTokenLimitContextKey))
				assert.Equal(t, tt.expectedCreatedAt, r.Context().Value(apiKeyCreatedAtContextKey))

				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("ok"))
			})

			h := AuthMiddleware(next, logger, tt.authService)

			request := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.authHeader != "" {
				request.Header.Set("Authorization", tt.authHeader)
			}
			recorder := httptest.NewRecorder()

			// Act
			h.ServeHTTP(recorder, request)

			response := recorder.Result()
			defer func() {
				_ = response.Body.Close()
			}()

			body, err := io.ReadAll(response.Body)
			assert.NoError(t, err)

			// Assert
			assert.Equal(t, tt.expectedStatusCode, response.StatusCode)
			assert.Equal(t, tt.expectedBody, string(body))
			assert.Equal(t, tt.expectNextCalled, nextCalled)
		})
	}
}
