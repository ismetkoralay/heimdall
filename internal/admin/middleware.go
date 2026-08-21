package admin

import (
	"crypto/subtle"
	"net/http"
	"strings"
)

func AdminMiddleware(
	adminKey string,
	next http.Handler,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		subs := strings.SplitN(authHeader, " ", 2)
		if len(subs) < 2 {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		header := strings.TrimPrefix(authHeader, "Bearer ")
		if subtle.ConstantTimeCompare([]byte(header), []byte(adminKey)) != 1 {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}
