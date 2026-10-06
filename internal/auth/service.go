package auth

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/devsubhamdas/go-rest-api-advanced/internal/auth/dto"
	"github.com/devsubhamdas/go-rest-api-advanced/internal/platform/errorsx"
	"github.com/devsubhamdas/go-rest-api-advanced/internal/platform/hash"
	"github.com/devsubhamdas/go-rest-api-advanced/internal/platform/token"
	"github.com/devsubhamdas/go-rest-api-advanced/internal/user"
	"github.com/google/uuid"
)

type TokenManager interface {
	GenerateAccess(token.AccessPayload) (string, time.Time, error)
	GenerateRefresh(token.RefreshPayload) (string, time.Time, error)
	Parse(raw string, typ token.Type) (*token.Claims, error)
}

type Tokens struct {
	Access     string
	AccessExp  time.Time
	Refresh    string
	RefreshExp time.Time
}

type UserReader interface {
	GetByID(context.Context, uuid.UUID) (*user.User, error)
	GetByEmail(context.Context, string) (*user.User, error)
}

type UserWriter interface {
	Create(context.Context, *user.User) (*user.User, error)
}

type service struct {
	userReader   UserReader
	userWriter   UserWriter
	tokenManager TokenManager
}

func NewService(r UserReader, w UserWriter, tm TokenManager) Service {
	return &service{
		userReader:   r,
		userWriter:   w,
		tokenManager: tm,
	}
}

func (s *service) Login(ctx context.Context, input *dto.LoginUserInput) (*dto.LoginResponseData, error) {
	input.Email = strings.TrimSpace(input.Email)

	// validate input
	if err := input.Validate(); err != nil {
		return nil, err
	}

	u, err := s.userReader.GetByEmail(ctx, input.Email)
	if err != nil {
		if errors.Is(err, errorsx.ErrNotFound) {
			return nil, errorsx.ErrInvalidEmail
		}
		return nil, err
	}

	// compare password with hash
	if err := hash.CompareWithPassword(input.Password, u.Password); err != nil {
		return nil, errorsx.ErrInvalidPassword
	}

	// generate jwt token
	t, err := s.issue(u)
	if err != nil {
		return nil, err
	}

	return &dto.LoginResponseData{
		AccessToken:     t.Access,
		AccessTokenExp:  t.AccessExp,
		RefreshToken:    t.Refresh,
		RefreshTokenExp: t.RefreshExp,
	}, nil
}

func (s *service) Refresh(ctx context.Context, rt string) (*dto.RefreshResponseData, error) {
	claims, err := s.tokenManager.Parse(rt, token.Refresh)

	if err != nil {
		return nil, errorsx.ErrInvalidToken
	}

	id, err := uuid.Parse(claims.Subject)
	if err != nil {
		return nil, errorsx.ErrInvalidToken
	}

	// re-load: fresh name/email, catches deleted users
	u, err := s.userReader.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, errorsx.ErrNotFound) {
			return nil, errorsx.ErrInvalidCredentials
		}
		return nil, err
	}

	// issues a new refresh token too (rotation)
	t, err := s.issue(u)
	if err != nil {
		return nil, err
	}

	return &dto.RefreshResponseData{
		AccessToken:     t.Access,
		AccessTokenExp:  t.AccessExp,
		RefreshToken:    t.Refresh,
		RefreshTokenExp: t.RefreshExp,
	}, nil
}

func (s *service) Singup(ctx context.Context, input *dto.SignupUserInput) (*dto.SignupResponseData, error) {
	// Sanitize input
	input.Name = strings.TrimSpace(input.Name)
	input.Email = strings.ToLower(strings.TrimSpace(input.Email))

	// Validate input
	if err := input.Validate(); err != nil {
		return nil, err
	}

	// Hash password
	hashPwd, err := hash.GenerateFromPassword(input.Password)
	if err != nil {
		return nil, err
	}

	payload := &user.User{
		Name:     input.Name,
		Email:    input.Email,
		Password: hashPwd,
	}

	u, err := s.userWriter.Create(ctx, payload)
	if err != nil {
		return nil, err
	}

	return &dto.SignupResponseData{
		ID:        u.ID,
		Name:      u.Name,
		Email:     u.Email,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}, nil
}

func (s *service) issue(u *user.User) (*Tokens, error) {
	access, accessExp, err := s.tokenManager.GenerateAccess(token.AccessPayload{
		ID: u.ID, Name: u.Name, Email: u.Email,
	})
	if err != nil {
		return nil, err
	}
	refresh, refreshExp, err := s.tokenManager.GenerateRefresh(token.RefreshPayload{ID: u.ID})
	if err != nil {
		return nil, err
	}
	return &Tokens{
		access,
		accessExp,
		refresh,
		refreshExp,
	}, nil
}
