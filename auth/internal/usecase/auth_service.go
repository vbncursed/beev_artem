package usecase

//go:generate go run github.com/gojuno/minimock/v3/cmd/minimock@v3.4.7 -i AuthStorage,SessionStorage,TokenIssuer -o ./mocks -s _mock.go -g

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

// TokenIssuer is the access/refresh token minting driven port. Implemented
// by infrastructure/jwt. Keeping the JWT library outside the use case lets
// us swap algorithms (HS256 → RS256, opaque tokens) without touching
// business logic.
type TokenIssuer interface {
	IssueAccess(userID uint64, email, role string) (string, error)
	IssueRefresh() (string, error)
	HashRefresh(token string) []byte
}

type AuthService struct {
	authStorage    AuthStorage
	sessionStorage SessionStorage
	tokenIssuer    TokenIssuer

	refreshTTL time.Duration
	bcryptCost int
}

// NewAuthService wires the use case with three driven ports + two business
// knobs (refresh-session lifetime, bcrypt cost). Access-token TTL lives
// inside the TokenIssuer because it's a JWT-format detail; refresh-session
// lifetime stays here because it controls the database row TTL.
//
// bcryptCost == 0 falls back to bcrypt.DefaultCost so callers don't have to
// know that constant.
func NewAuthService(
	authStorage AuthStorage,
	sessionStorage SessionStorage,
	tokenIssuer TokenIssuer,
	refreshTTL time.Duration,
	bcryptCost int,
) *AuthService {
	return &AuthService{
		authStorage:    authStorage,
		sessionStorage: sessionStorage,
		tokenIssuer:    tokenIssuer,
		refreshTTL:     refreshTTL,
		bcryptCost:     bcryptCost,
	}
}
