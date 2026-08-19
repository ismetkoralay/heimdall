package api

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/ismetkoralay/heimdall/internal/auth"
	"github.com/ismetkoralay/heimdall/internal/reqctx"
)

type AuthService interface {
	ValidateAPIKey(ctx context.Context, key string) (auth.APIKey, error)
}

func AuthMiddleware(
	next http.Handler,
	logger *slog.Logger,
	authService AuthService,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.Info("auth middleware begin")
		authHeader := r.Header.Get("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer") {
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}

		subs := strings.SplitN(authHeader, " ", 2)
		if len(subs) < 2 {
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}

		token := subs[1]
		res, err := authService.ValidateAPIKey(r.Context(), token)
		var authErr *auth.AuthError
		if errors.As(err, &authErr) {
			http.Error(w, authErr.Unwrap().Error(), authErr.StatusCode)
			return
		} else if err != nil {
			logger.ErrorContext(r.Context(), err.Error())
			http.Error(w, "something went wrong", http.StatusInternalServerError)
			return
		}

		ctx := reqctx.SetAPIKeyID(r.Context(), res.ID)
		ctx = reqctx.SetAPIKeyRPMLimit(ctx, res.RPMLimit)
		ctx = reqctx.SetAPIKeyDailyTokenQuota(ctx, res.DailyTokenQuota)

		next.ServeHTTP(w, r.WithContext(ctx))
		logger.Info("auth middleware done")
	})
}
