package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/devsubhamdas/go-rest-api-advanced/internal/platform/response"
)

type ctxKey int

const (
	userIDCtxKey ctxKey = iota
)

type SessionValidator interface {
	Validate(string) (string, error)
}

func RequireAuth(sv SessionValidator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			raw, ok := bearerToken(r)
			if !ok {
				_ = response.WriteErrorFromCode(w, response.CodeUnauthorized, "unauthorized user access")
				return
			}

			userID, err := sv.Validate(raw)
			if err != nil || userID == "" {
				_ = response.WriteErrorFromCode(w, response.CodeUnauthorized, "invalid authorization token")
				return
			}

			ctx := context.WithValue(r.Context(), userIDCtxKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetUserIDFromContext(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(userIDCtxKey).(string)
	return id, ok && id != ""
}

func bearerToken(r *http.Request) (string, bool) {
	const prefix = "Bearer "

	h := r.Header.Get("Authorization")
	if len(h) <= len(prefix) || !strings.EqualFold(h[:len(prefix)], prefix) {
		return "", false
	}

	raw := strings.TrimSpace(h[len(prefix):])

	return raw, raw != ""
}
