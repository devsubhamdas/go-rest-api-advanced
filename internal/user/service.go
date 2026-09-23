package user

import (
	"context"
	"fmt"
	"strings"

	"github.com/devsubhamdas/go-rest-api-advanced/internal/platform/errorsx"
	"github.com/devsubhamdas/go-rest-api-advanced/internal/platform/hash"
	"github.com/devsubhamdas/go-rest-api-advanced/internal/user/dto"
	"github.com/google/uuid"
)

type Repository interface {
	Create(context.Context, *User) (*User, error)
	Update(context.Context, *User) (*User, error)
	Delete(context.Context, uuid.UUID) error
	GetByID(context.Context, uuid.UUID) (*User, error)
	GetByEmail(context.Context, string) (*User, error)
	GetMany(context.Context) ([]User, error)
}

type service struct {
	repo Repository
}

func NewService(r Repository) Service {
	return &service{
		repo: r,
	}
}

func (s *service) Create(ctx context.Context, input *dto.CreateUserInput) (*dto.UserResponseData, error) {
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

	payload := &User{
		Name:     input.Name,
		Email:    input.Email,
		Password: hashPwd,
	}

	u, err := s.repo.Create(ctx, payload)
	if err != nil {
		return nil, err
	}

	return &dto.UserResponseData{
		ID:        u.ID,
		Name:      u.Name,
		Email:     u.Email,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}, nil
}

func (s *service) Update(ctx context.Context, id string, input *dto.UpdateUserInput) (*dto.UserResponseData, error) {
	pID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errorsx.ErrInvalidID, err)
	}

	// Sanitize input
	input.Name = strings.TrimSpace(input.Name)
	input.Email = strings.ToLower(strings.TrimSpace(input.Email))

	// Validate input
	if err := input.Validate(); err != nil {
		return nil, err
	}

	payload := &User{
		ID:    pID,
		Name:  input.Name,
		Email: input.Email,
	}

	u, err := s.repo.Update(ctx, payload)
	if err != nil {
		return nil, err
	}

	return &dto.UserResponseData{
		ID:        u.ID,
		Name:      u.Name,
		Email:     u.Email,
		UpdatedAt: u.UpdatedAt,
	}, nil
}

func (s *service) Delete(ctx context.Context, id string) error {
	// TODO: validate input

	pID, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("%w: %w", errorsx.ErrInvalidID, err)
	}
	return s.repo.Delete(ctx, pID)
}

func (s *service) GetByID(ctx context.Context, id string) (*dto.UserResponseData, error) {
	// TODO: validate input

	pID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errorsx.ErrInvalidID, err)
	}

	u, err := s.repo.GetByID(ctx, pID)
	if err != nil {
		return nil, err
	}

	return &dto.UserResponseData{
		ID:        u.ID,
		Name:      u.Name,
		Email:     u.Email,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}, nil
}

func (s *service) GetByEmail(ctx context.Context, email string) (*dto.UserResponseData, error) {
	// TODO: validate input

	u, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	return &dto.UserResponseData{
		ID:        u.ID,
		Name:      u.Name,
		Email:     u.Email,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}, nil
}

func (s *service) GetMany(ctx context.Context) ([]dto.UserResponseData, error) {
	users, err := s.repo.GetMany(ctx)
	if err != nil {
		return nil, err
	}

	data := make([]dto.UserResponseData, 0, len(users))

	for _, u := range users {
		data = append(data, dto.UserResponseData{
			ID:        u.ID,
			Name:      u.Name,
			Email:     u.Email,
			CreatedAt: u.CreatedAt,
			UpdatedAt: u.UpdatedAt,
		})
	}

	return data, nil
}
