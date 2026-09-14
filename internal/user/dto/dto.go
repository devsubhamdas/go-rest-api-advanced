package dto

import (
	"time"

	"github.com/google/uuid"
)

type CreateUserInput struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UpdateUserInput struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type UserResponseData struct {
	ID    uuid.UUID `json:"id"`
	Name  string    `json:"name,omitempty"`
	Email string    `json:"email,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
