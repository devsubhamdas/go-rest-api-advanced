package middleware

import "net/http"

// SecurityHeaders sets common hardening headers on every response.
// This is the JSON-API equivalent of XSS protection: since you're not
// server-rendering HTML, the CSP mainly protects any Swagger/docs page
// you might serve from this same server, and the other headers stop
// this API's responses from being framed or MIME-sniffed by a browser.
func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("X-Frame-Options", "DENY")
		h.Set("Referrer-Policy", "strict-origin-when-cross-origin")
		h.Set("Content-Security-Policy", "default-src 'self'")
		h.Set("Strict-Transport-Security", "max-age=63072000; includeSubDomains")
		next.ServeHTTP(w, r)
	})
}
