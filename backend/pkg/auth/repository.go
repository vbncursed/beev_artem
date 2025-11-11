package auth

import (
	"context"
	"errors"
)

// Common errors used by repository/use cases
var (
	ErrNotFound          = errors.New("not found")
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
)

// UserRepository abstracts persistence concerns from the domain layer.
// Implementations may be in-memory, SQL, NoSQL, etc.
type UserRepository interface {
	Create(ctx context.Context, user User) error
	GetByEmail(ctx context.Context, email string) (User, error)
}


