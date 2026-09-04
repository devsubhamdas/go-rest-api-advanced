package header

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

type ctxKey int

const (
	requestIDCtxKey ctxKey = iota
)

const (
	requestIDHeaderKey string = "X-Request-ID"
)

func SetRequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqID := r.Header.Get(requestIDHeaderKey)
		if reqID == "" {
			reqID = uuid.NewString()
		}

		w.Header().Add(requestIDHeaderKey, reqID)
		ctx := context.WithValue(r.Context(), requestIDCtxKey, reqID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetRequestIDFromContext(ctx context.Context) string {
	return ctx.Value(requestIDCtxKey).(string)
}
