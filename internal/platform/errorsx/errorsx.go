package errorsx

import (
	"errors"

	"github.com/devsubhamdas/go-rest-api-advanced/internal/platform/response"
)

var (
	ErrEmptyField         = errors.New("empty field not allowed")
	ErrInvalidID          = errors.New("invalid id")
	ErrInvalidEmail       = errors.New("invalid email")
	ErrInvalidName        = errors.New("invalid name")
	ErrInvalidPassword    = errors.New("invalid password")
	ErrEmailAlreadyExists = errors.New("email already exists")
	ErrDuplicateKey       = errors.New("duplicate key error")
	ErrNotFound           = errors.New("not found")
	ErrInvalidToken       = errors.New("invalid jwt token")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUnknown            = errors.New("unknown error")
)

type ValidationErrorDetails struct {
	Field   string `json:"field,omitempty"`
	Message string `json:"message,omitempty"`
}

type ValidationError struct {
	Code    response.Code
	Message string                   `json:"message,omitempty"`
	Details []ValidationErrorDetails `json:"details,omitempty"`
}

func (ve *ValidationError) Error() string {
	return "validation error"
}
