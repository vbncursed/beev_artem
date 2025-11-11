package auth

import (
	"time"

	"github.com/google/uuid"
)

// User is a domain entity representing a system user.
type User struct {
	ID           uuid.UUID
	Email        string
	PasswordHash string
	CreatedAt    time.Time
}


