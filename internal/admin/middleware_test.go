package admin

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAdminMiddleware(t *testing.T) {
	tests := []struct {
		name               string
		adminKey           string
		header             string
		expectedStatusCode int
		expectedBody       string
		expectNextCalled   bool
	}{
		{
			name:               "successful call",
			adminKey:           "admin-key",
			header:             "Bearer admin-key",
			expectedStatusCode: http.StatusOK,
			expectedBody:       "ok",
			expectNextCalled:   true,
		},
		{
			name:               "fails when auth header is missing",
			adminKey:           "admin-key",
			expectedStatusCode: http.StatusUnauthorized,
			expectedBody:       "unauthorized\n",
		},
		{
			name:               "fails when auth header is not starting with bearer",
			adminKey:           "admin-key",
			header:             "invalid test",
			expectedStatusCode: http.StatusUnauthorized,
			expectedBody:       "unauthorized\n",
		},
		{
			name:               "fails when auth header only has bearer",
			adminKey:           "admin-key",
			header:             "Bearer",
			expectedStatusCode: http.StatusUnauthorized,
			expectedBody:       "unauthorized\n",
		},
		{
			name:               "fails when auth header has only prefix - invalid",
			adminKey:           "admin-key",
			header:             "Bearerxxx",
			expectedStatusCode: http.StatusUnauthorized,
			expectedBody:       "unauthorized\n",
		},
		{
			name:               "fails when auth header has only prefix - invalid",
			adminKey:           "admin-key",
			header:             "adminkey",
			expectedStatusCode: http.StatusUnauthorized,
			expectedBody:       "unauthorized\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			nextCalled := false
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				nextCalled = true

				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("ok"))
			})

			h := AdminMiddleware(tt.adminKey, next)

			request := httptest.NewRequest(http.MethodPost, "/", nil)
			if tt.header != "" {
				request.Header.Set("Authorization", tt.header)
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
