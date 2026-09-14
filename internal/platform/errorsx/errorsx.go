package errorsx

import "errors"

var (
	ErrInvalidID          = errors.New("invalid id")
	ErrInvalidEmail       = errors.New("invalid email")
	ErrEmailAlreadyExists = errors.New("email already exists")
	ErrDuplicateKey       = errors.New("duplicate key error")
	ErrNotFound           = errors.New("not found")
	ErrUnknown            = errors.New("unknown error")
)
