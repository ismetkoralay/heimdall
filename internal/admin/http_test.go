package admin

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

type MockAuthService struct {
	generateAndSaveAPIKeyFn func(ctx context.Context, name string, rpmLimit, dailyTokenQuota int) (string, error)
	capturedName            string
	capturedRPMLimit        int
	capturedDailyTokenQuota int
}

func (s *MockAuthService) GenerateAndSaveAPIKey(ctx context.Context, name string, rpmLimit int, dailyTokenQuota int) (string, error) {
	s.capturedName = name
	s.capturedRPMLimit = rpmLimit
	s.capturedDailyTokenQuota = dailyTokenQuota
	return s.generateAndSaveAPIKeyFn(ctx, name, rpmLimit, dailyTokenQuota)
}

func TestCreateKey(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	var (
		generatedKey = "generated-key"
		validRequest = CreateKeyRequest{
			Name:            "name",
			RPMLimit:        1,
			DailyTokenQuota: 2,
		}
		validResponse = CreateKeyResponse{
			Key: generatedKey,
		}
		invalidRequest  = []byte("invalid-request")
		invalidResponse = CreateKeyErrorResponse{
			Error: "invalid request",
		}
		missingNameRequest = CreateKeyRequest{
			RPMLimit:        1,
			DailyTokenQuota: 2,
		}
		missingNameResponse = CreateKeyErrorResponse{
			Error: "name cannot be empty",
		}
		missingRPMLimitRequest = CreateKeyRequest{
			Name:            "name",
			DailyTokenQuota: 2,
		}
		missingRPMLimitResponse = CreateKeyErrorResponse{
			Error: "request per minute limit must be greater than 0",
		}
		missingDailyTokenQuotaRequest = CreateKeyRequest{
			Name:     "name",
			RPMLimit: 1,
		}
		missingDailyTokenQuotaResponse = CreateKeyErrorResponse{
			Error: "daily token quota must be greater than 0",
		}
		authServiceFailsResponse = CreateKeyErrorResponse{
			Error: "something went wrong",
		}
	)

	validRequestJson, err := json.Marshal(validRequest)
	assert.NoError(t, err)

	validResponseJson, err := json.Marshal(validResponse)
	assert.NoError(t, err)
	validResponseJson = append(validResponseJson, []byte("\n")...)

	invalidResponseJson, err := json.Marshal(invalidResponse)
	assert.NoError(t, err)
	invalidResponseJson = append(invalidResponseJson, []byte("\n")...)

	missingNameRequestJson, err := json.Marshal(missingNameRequest)
	assert.NoError(t, err)

	missingNameResponseJson, err := json.Marshal(missingNameResponse)
	assert.NoError(t, err)
	missingNameResponseJson = append(missingNameResponseJson, []byte("\n")...)

	missingRPMLimitRequestJson, err := json.Marshal(missingRPMLimitRequest)
	assert.NoError(t, err)

	missingRPMLimitResponseJson, err := json.Marshal(missingRPMLimitResponse)
	assert.NoError(t, err)
	missingRPMLimitResponseJson = append(missingRPMLimitResponseJson, []byte("\n")...)

	missingDailyTokenQuotaRequestJson, err := json.Marshal(missingDailyTokenQuotaRequest)
	assert.NoError(t, err)

	missingDailyTokenQuotaResponseJson, err := json.Marshal(missingDailyTokenQuotaResponse)
	assert.NoError(t, err)
	missingDailyTokenQuotaResponseJson = append(missingDailyTokenQuotaResponseJson, []byte("\n")...)

	authServiceFailsResponseJson, err := json.Marshal(authServiceFailsResponse)
	assert.NoError(t, err)
	authServiceFailsResponseJson = append(authServiceFailsResponseJson, []byte("\n")...)

	tests := []struct {
		name                 string
		authService          *MockAuthService
		method               string
		requestBody          []byte
		expectedResponseBody []byte
		expectedStatusCode   int
	}{
		{
			name: "successful call",
			authService: &MockAuthService{
				generateAndSaveAPIKeyFn: func(ctx context.Context, name string, rpmLimit, dailyTokenQuota int) (string, error) {
					return generatedKey, nil
				},
			},
			method:               http.MethodPost,
			requestBody:          validRequestJson,
			expectedResponseBody: validResponseJson,
			expectedStatusCode:   http.StatusOK,
		},
		{
			name:                 "fails when request is invalid",
			authService:          &MockAuthService{},
			method:               http.MethodPost,
			requestBody:          invalidRequest,
			expectedResponseBody: invalidResponseJson,
			expectedStatusCode:   http.StatusBadRequest,
		},
		{
			name:                 "fails when name is missing in request",
			authService:          &MockAuthService{},
			method:               http.MethodPost,
			requestBody:          missingNameRequestJson,
			expectedResponseBody: missingNameResponseJson,
			expectedStatusCode:   http.StatusBadRequest,
		},
		{
			name:                 "fails when rpm limit is missing in request",
			authService:          &MockAuthService{},
			method:               http.MethodPost,
			requestBody:          missingRPMLimitRequestJson,
			expectedResponseBody: missingRPMLimitResponseJson,
			expectedStatusCode:   http.StatusBadRequest,
		},
		{
			name:                 "fails when daily token quota is missing in request",
			authService:          &MockAuthService{},
			method:               http.MethodPost,
			requestBody:          missingDailyTokenQuotaRequestJson,
			expectedResponseBody: missingDailyTokenQuotaResponseJson,
			expectedStatusCode:   http.StatusBadRequest,
		},
		{
			name: "fails when auth service returns error",
			authService: &MockAuthService{
				generateAndSaveAPIKeyFn: func(ctx context.Context, name string, rpmLimit, dailyTokenQuota int) (string, error) {
					return "", errors.New("internal error")
				},
			},
			method:               http.MethodPost,
			requestBody:          validRequestJson,
			expectedResponseBody: authServiceFailsResponseJson,
			expectedStatusCode:   http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			h := CreateKeyHandler(logger, tt.authService)

			request := httptest.NewRequest(tt.method, "/", bytes.NewReader(tt.requestBody))
			recorder := httptest.NewRecorder()

			h.ServeHTTP(recorder, request)

			// Act
			response := recorder.Result()
			defer func() {
				_ = response.Body.Close()
			}()

			responseBody, err := io.ReadAll(response.Body)
			assert.NoError(t, err)

			// Assert
			assert.Equal(t, tt.expectedResponseBody, responseBody)
			assert.Equal(t, tt.expectedStatusCode, response.StatusCode)
		})
	}
}
