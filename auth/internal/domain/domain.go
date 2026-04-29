package domain

import "time"

const (
	RoleUser  = "user"
	RoleAdmin = "admin"
)

type User struct {
	ID           uint64
	Email        string
	PasswordHash string
	Role         string
	CreatedAt    time.Time
}

type Session struct {
	UserID      uint64
	RefreshHash []byte
	ExpiresAt   time.Time
	CreatedAt   time.Time
	UserAgent   string
	IP          string
}

type AuthInput struct {
	Email     string
	Password  string
	UserAgent string
	IP        string
}

type RegisterInput = AuthInput

type LoginInput = AuthInput

type RefreshInput struct {
	RefreshToken string
	UserAgent    string
	IP           string
}

type AuthInfo struct {
	UserID       uint64
	AccessToken  string
	RefreshToken string
}
