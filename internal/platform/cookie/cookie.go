package cookie

import (
	"net/http"
	"time"
)

const sessionCookieName = "session_token"

type Config struct {
	Path     string
	HttpOnly bool
	Secure   bool
	SameSite http.SameSite
}

// SetSessionToken writes a secure session cookie. Call this from your
// login handler after issuing a token.
func SetSessionToken(cfg *Config, w http.ResponseWriter, token string, exp time.Time) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    token,
		Path:     cfg.Path,
		HttpOnly: cfg.HttpOnly, // blocks JS access — mitigates XSS token theft
		Secure:   cfg.Secure,   // HTTPS only; set false only for local http dev
		SameSite: cfg.SameSite, // use Strict if you don't need cross-site nav
		Expires:  exp,
		MaxAge:   int(time.Until(exp).Seconds()),
	})
}

// ClearSessionToken expires the session cookie immediately. Call this
// from your logout handler.
func ClearSessionToken(cfg *Config, w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     cfg.Path,
		HttpOnly: cfg.HttpOnly,
		Secure:   cfg.Secure,
		SameSite: cfg.SameSite,
		Expires:  time.Unix(0, 0), // fallback for old browser
		MaxAge:   -1,              // deletes the cookie
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
