package auth

import "context"

// TokenGenerator abstracts token creation (e.g., JWT).
// It allows use cases to stay framework-agnostic.
type TokenGenerator interface {
    Generate(ctx context.Context, user User) (string, error)
}


