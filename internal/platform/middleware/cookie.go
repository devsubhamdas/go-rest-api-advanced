package middleware

import (
	"net/http"
	"time"
)

const sessionCookieName = "session_token"

// SetSessionCookie writes a secure session cookie. Call this from your
// login handler after issuing a token.
func SetSessionCookie(w http.ResponseWriter, token string, ttl time.Duration) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,                 // blocks JS access — mitigates XSS token theft
		Secure:   true,                 // HTTPS only; set false only for local http dev
		SameSite: http.SameSiteLaxMode, // use Strict if you don't need cross-site nav
		MaxAge:   int(ttl.Seconds()),
	})
}

// ClearSessionCookie expires the session cookie immediately. Call this
// from your logout handler.
func ClearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1, // deletes the cookie
	})
}

// ExtractSessionToken pulls the raw token out of the request cookie.
// It does no validation — that's the auth layer's job. Returns an error
// if the cookie is missing or empty so callers can 401 without caring
// about cookie mechanics.
func ExtractSessionToken(r *http.Request) (string, error) {
	cookie, err := r.Cookie(sessionCookieName)
	if err != nil {
		return "", err
	}
	if cookie.Value == "" {
		return "", http.ErrNoCookie
	}
	return cookie.Value, nil
}
