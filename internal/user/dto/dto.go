package dto

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/devsubhamdas/go-rest-api-advanced/internal/platform/errorsx"
	"github.com/devsubhamdas/go-rest-api-advanced/internal/platform/response"
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

	CreatedAt time.Time `json:"created_at,omitzero"`
	UpdatedAt time.Time `json:"updated_at,omitzero"`
}

// CreateUserInput Validation
func (i *CreateUserInput) Validate() error {
	vErrDetails := []errorsx.ValidationErrorDetails{}

	// Empty fields validation
	if strings.TrimSpace(i.Name) == "" {
		vErrDetails = append(
			vErrDetails,
			errorsx.ValidationErrorDetails{
				Field:   "name",
				Message: errorsx.ErrEmptyField.Error(),
			},
		)
	}

	if strings.TrimSpace(i.Email) == "" {
		vErrDetails = append(
			vErrDetails,
			errorsx.ValidationErrorDetails{
				Field:   "email",
				Message: errorsx.ErrEmptyField.Error(),
			},
		)
	}

	if strings.TrimSpace(i.Password) == "" {
		vErrDetails = append(
			vErrDetails,
			errorsx.ValidationErrorDetails{
				Field:   "password",
				Message: errorsx.ErrEmptyField.Error(),
			},
		)
	}

	// Validate Name
	if !isValidName(i.Name) {
		vErrDetails = append(vErrDetails, errorsx.ValidationErrorDetails{
			Field: "name",
			Message: fmt.Sprintf(
				"%s: only alphabets with no trailing spaces allowed and no numbers or special characters except '.'",
				errorsx.ErrInvalidName.Error(),
			),
		})
	}

	// Validate Email
	if !isValidEmail(i.Email) {
		vErrDetails = append(vErrDetails, errorsx.ValidationErrorDetails{
			Field:   "email",
			Message: errorsx.ErrInvalidEmail.Error(),
		})
	}

	// Validate Password
	minPwdLen := 4
	if len(i.Password) < minPwdLen {
		vErrDetails = append(vErrDetails, errorsx.ValidationErrorDetails{
			Field:   "password",
			Message: fmt.Sprintf("password length must be greater than %d", minPwdLen),
		})
	}

	if len(vErrDetails) > 0 {
		return &errorsx.ValidationError{
			Code:    response.CodeUnprocessableEntity,
			Details: vErrDetails,
		}
	}

	return nil
}

// UpdateUserInput Validation
func (i *UpdateUserInput) Validate() error {
	vErrDetails := []errorsx.ValidationErrorDetails{}

	// Empty fields validation
	if strings.TrimSpace(i.Name) == "" {
		vErrDetails = append(
			vErrDetails,
			errorsx.ValidationErrorDetails{
				Field:   "name",
				Message: errorsx.ErrEmptyField.Error(),
			},
		)
	}

	if strings.TrimSpace(i.Email) == "" {
		vErrDetails = append(
			vErrDetails,
			errorsx.ValidationErrorDetails{
				Field:   "email",
				Message: errorsx.ErrEmptyField.Error(),
			},
		)
	}

	// Validate Name
	if !isValidName(i.Name) {
		vErrDetails = append(vErrDetails, errorsx.ValidationErrorDetails{
			Field: "name",
			Message: fmt.Sprintf(
				"%s: only alphabets with no trailing spaces allowed and no numbers or special characters except '.'",
				errorsx.ErrInvalidName.Error(),
			),
		})
	}

	// Validate Email
	if !isValidEmail(i.Email) {
		vErrDetails = append(vErrDetails, errorsx.ValidationErrorDetails{
			Field:   "email",
			Message: errorsx.ErrInvalidEmail.Error(),
		})
	}

	if len(vErrDetails) > 0 {
		return &errorsx.ValidationError{
			Code:    response.CodeUnprocessableEntity,
			Details: vErrDetails,
		}
	}

	return nil
}

func isValidName(name string) bool {
	nameRegex := regexp.MustCompile(`^[a-zA-Z]+\.?(\s+[a-zA-Z]+\.?)*$`)
	return nameRegex.MatchString(name)
}

func isValidEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}
