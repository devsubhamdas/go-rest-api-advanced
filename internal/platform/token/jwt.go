package token

import (
	"errors"
	"fmt"
	"time"

	"github.com/devsubhamdas/go-rest-api-advanced/internal/platform/errorsx"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type AccessPayload struct {
	ID    uuid.UUID
	Name  string
	Email string
}

type RefreshPayload struct {
	ID uuid.UUID
}

type Type string

const (
	Access  Type = "access"
	Refresh Type = "refresh"
)

type Claims struct {
	Type  Type   `json:"typ"`
	Name  string `json:"name,omitempty"`  // for access token only
	Email string `json:"email,omitempty"` // for access token only
	jwt.RegisteredClaims
}

type Config struct {
	Issuer        string
	AccessSecret  []byte
	RefreshSecret []byte
	AccessTTL     time.Duration // e.g. 15 * time.Minute
	RefreshTTL    time.Duration // e.g. 7 * 24 * time.Hour
}

type Manager struct {
	cfg *Config
	now func() time.Time // injectable for tests
}

func NewManager(cfg *Config) (*Manager, error) {
	if len(cfg.AccessSecret) < 32 || len(cfg.RefreshSecret) < 32 {
		return nil, errors.New("token: secrets must be at least 32 bytes")
	}
	if cfg.AccessTTL <= 0 || cfg.RefreshTTL <= 0 {
		return nil, errors.New("token: TTLs must be positive")
	}
	return &Manager{cfg: cfg, now: time.Now}, nil
}

// GenerateAccess returns a short-lived access token.
func (m *Manager) GenerateAccess(p AccessPayload) (string, time.Time, error) {
	c := Claims{
		Type:  Access,
		Name:  p.Name,
		Email: p.Email,
	}
	c.Subject = p.ID.String()
	return m.sign(c, Access, m.cfg.AccessSecret, m.cfg.AccessTTL)
}

// GenerateRefresh returns a long-lived refresh token and its expiry
// (use the expiry for the cookie's Expires/MaxAge).
func (m *Manager) GenerateRefresh(p RefreshPayload) (string, time.Time, error) {
	c := Claims{
		Type: Refresh,
	}
	c.Subject = p.ID.String()
	return m.sign(c, Refresh, m.cfg.RefreshSecret, m.cfg.RefreshTTL)
}

func (m *Manager) sign(c Claims, typ Type, secret []byte, ttl time.Duration) (string, time.Time, error) {
	now := m.now()
	expiresAt := now.Add(ttl)

	c.Issuer = m.cfg.Issuer
	c.ID = uuid.NewString() // jti: lets you add revocation/rotation later
	c.IssuedAt = jwt.NewNumericDate(now)
	c.NotBefore = jwt.NewNumericDate(now)
	c.ExpiresAt = jwt.NewNumericDate(expiresAt)

	signed, err := jwt.NewWithClaims(jwt.SigningMethodHS256, c).SignedString(secret)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("token: sign %s token: %w", typ, err)
	}
	return signed, expiresAt, nil
}

func (m *Manager) Parse(raw string, typ Type) (*Claims, error) {
	secret := m.cfg.AccessSecret
	if typ == Refresh {
		secret = m.cfg.RefreshSecret
	}

	claims := &Claims{}
	_, err := jwt.ParseWithClaims(
		raw,
		claims,
		func(*jwt.Token) (any, error) {
			return secret, nil
		},
		jwt.WithValidMethods([]string{"HS256"}),
		jwt.WithIssuer(m.cfg.Issuer),
		jwt.WithExpirationRequired(),
	)
	if err != nil || claims.Type != typ {
		return nil, errorsx.ErrInvalidToken
	}
	return claims, nil
}
