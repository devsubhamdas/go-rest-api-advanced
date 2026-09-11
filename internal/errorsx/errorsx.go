package errorsx

import "errors"

var (
	ErrEmailAlreadyExists = errors.New("email already exists")
	ErrDuplicateKey       = errors.New("duplicate key error")
	ErrNotFound           = errors.New("not found")
)
