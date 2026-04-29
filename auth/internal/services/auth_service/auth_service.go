package auth_service

//go:generate go run github.com/gojuno/minimock/v3/cmd/minimock@v3.4.7 -i AuthStorage,SessionStorage -o ./mocks -s _mock.go -g

import (
	"context"
	"time"

	"github.com/artem13815/hr/auth/internal/domain"
)

type AuthStorage interface {
	CreateUser(ctx context.Context, email string, passwordHash string) (uint64, error)
	GetUserByEmail(ctx context.Context, email string) (*domain.User, error)
	GetUserByID(ctx context.Context, userID uint64) (*domain.User, error)
	UpdateUserRole(ctx context.Context, userID uint64, role string) error
}

type SessionStorage interface {
	CreateSession(ctx context.Context, userID uint64, refreshHash []byte, expiresAt time.Time, userAgent, ip string) error
	GetSessionByRefreshHash(ctx context.Context, refreshHash []byte) (*domain.Session, error)
	// ConsumeSessionByRefreshHash atomically returns and deletes a session in one
	// round-trip. Used by Refresh to make refresh tokens one-shot — concurrent
	// replays of the same token race for the single existing record; only one
	// wins.
	ConsumeSessionByRefreshHash(ctx context.Context, refreshHash []byte) (*domain.Session, error)
	RevokeSessionByRefreshHash(ctx context.Context, refreshHash []byte) error
	RevokeAllSessionsByUserID(ctx context.Context, userID uint64) error
}

type AuthService struct {
	authStorage    AuthStorage
	sessionStorage SessionStorage

	jwtSecret  string
	accessTTL  time.Duration
	refreshTTL time.Duration
	bcryptCost int
}

// NewAuthService takes durations (not int64 seconds) and the bcrypt cost as
// proper-typed parameters. The bootstrap layer is responsible for converting
// YAML scalars into the right types. bcryptCost == 0 falls back to
// bcrypt.DefaultCost so callers don't have to know that constant.
func NewAuthService(
	authStorage AuthStorage,
	sessionStorage SessionStorage,
	jwtSecret string,
	accessTTL time.Duration,
	refreshTTL time.Duration,
	bcryptCost int,
) *AuthService {
	return &AuthService{
		authStorage:    authStorage,
		sessionStorage: sessionStorage,
		jwtSecret:      jwtSecret,
		accessTTL:      accessTTL,
		refreshTTL:     refreshTTL,
		bcryptCost:     bcryptCost,
	}
}
