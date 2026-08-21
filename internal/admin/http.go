package admin

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
)

type AuthService interface {
	GenerateAndSaveAPIKey(ctx context.Context, name string, rpmLimit int, dailyTokenQuota int) (string, error)
}

func CreateKeyHandler(
	logger *slog.Logger,
	authService AuthService,
) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			_ = r.Body.Close()
		}()

		w.Header().Set("Content-Type", "application/json")

		var request CreateKeyRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			errResponse := CreateKeyErrorResponse{
				Error: "invalid request",
			}
			_ = json.NewEncoder(w).Encode(errResponse)
			return
		}

		if strings.TrimSpace(request.Name) == "" {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(CreateKeyErrorResponse{
				Error: "name cannot be empty",
			})
			return
		}

		if request.RPMLimit < 1 {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(CreateKeyErrorResponse{
				Error: "request per minute limit must be greater than 0",
			})
			return
		}

		if request.DailyTokenQuota < 1 {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(CreateKeyErrorResponse{
				Error: "daily token quota must be greater than 0",
			})
			return
		}

		key, err := authService.GenerateAndSaveAPIKey(r.Context(), request.Name, request.RPMLimit, request.DailyTokenQuota)
		if err != nil {
			logger.Error("error generating key", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(CreateKeyErrorResponse{
				Error: "something went wrong",
			})
			return
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(CreateKeyResponse{
			Key: key,
		})
	})

	return mux
}

type CreateKeyRequest struct {
	Name            string `json:"name"`
	RPMLimit        int    `json:"request_per_minute_limit"`
	DailyTokenQuota int    `json:"daily_token_quota"`
}

type CreateKeyResponse struct {
	Key string `json:"key"`
}

type CreateKeyErrorResponse struct {
	Error string `json:"error"`
}
