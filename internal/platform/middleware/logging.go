package middleware

import (
	"log"
	"net/http"
	"time"
)

// statusRecorder wraps ResponseWriter so we can capture the status code,
// since http.ResponseWriter doesn't expose it after WriteHeader is called.
type statusRecorder struct {
	status int
	http.ResponseWriter
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)

}

// Logging logs method, path, status code, and duration for every request.
func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		rec := &statusRecorder{
			ResponseWriter: w,
			status:         http.StatusOK,
		}

		next.ServeHTTP(rec, r)

		log.Printf(
			"%s %s %d %s",
			r.Method,
			r.URL.Path,
			rec.status,
			time.Since(start),
		)
	})
}
