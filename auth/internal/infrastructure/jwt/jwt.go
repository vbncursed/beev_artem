// Package jwt is the infrastructure adapter that owns access/refresh token
// generation and access-token validation. The use case talks to it through
// usecase.TokenIssuer; the transport middleware talks to it through
// usecase.TokenValidator (or directly via the Validator struct). Keeping
// `github.com/golang-jwt/jwt/v5` confined to this package lets us swap to a
// different JWT library — or move to opaque tokens — without touching use
// cases or transport handlers.
package jwt

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"time"

	jwtlib "github.com/golang-jwt/jwt/v5"
)

// Claims is the parsed access-token payload. Plain types so consumers don't
// need to import the JWT library.
type Claims struct {
	UserID uint64
	Email  string
	Role   string
}

// Issuer mints access and refresh tokens. accessTTL is baked in at
// construction so callers don't pass it on every issue.
type Issuer struct {
	secret    string
	accessTTL time.Duration
}

func NewIssuer(secret string, accessTTL time.Duration) *Issuer {
	return &Issuer{secret: secret, accessTTL: accessTTL}
}

// IssueAccess signs a fresh HS256 JWT carrying user_id/email/role + iat/exp.
// Both `sub` and `user_id` are populated for backward compatibility with
// older tokens that only had `sub`.
func (i *Issuer) IssueAccess(userID uint64, email, role string) (string, error) {
	now := time.Now()
	claims := jwtlib.MapClaims{
		"sub":     userID,
		"user_id": userID,
		"email":   email,
		"role":    role,
		"iat":     now.Unix(),
		"exp":     now.Add(i.accessTTL).Unix(),
	}
	token := jwtlib.NewWithClaims(jwtlib.SigningMethodHS256, claims)
	return token.SignedString([]byte(i.secret))
}

// IssueRefresh returns a fresh opaque refresh token. We never sign refresh
// tokens — a 32-byte random string is enough; verification happens by
// hashing and looking up the SessionStorage row.
func (i *Issuer) IssueRefresh() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// HashRefresh produces the SHA-256 fingerprint stored in SessionStorage. The
// raw refresh token is never persisted — only its hash, so a leaked DB dump
// does not let the attacker replay sessions.
//
// Exposed as both a package-level function and an Issuer method so tests can
// compute hashes without instantiating an Issuer (no state involved).
func HashRefresh(token string) []byte {
	h := sha256.Sum256([]byte(token))
	return h[:]
}

func (Issuer) HashRefresh(token string) []byte { return HashRefresh(token) }

// Validator parses and verifies access tokens. It is shared between transport
// middleware (interceptor) and any handler that takes a raw access token (e.g.
// ValidateAccessToken RPC for the gateway).
type Validator struct {
	secret string
	parser *jwtlib.Parser
}

// NewValidator builds the validator with HS256 pinned and `exp` required.
// Pinning the algorithm defends against the classic alg-confusion / alg=none
// attacks. Requiring `exp` rejects any future bug that issues an unbounded
// token.
func NewValidator(secret string) *Validator {
	parser := jwtlib.NewParser(
		jwtlib.WithValidMethods([]string{jwtlib.SigningMethodHS256.Alg()}),
		jwtlib.WithExpirationRequired(),
	)
	return &Validator{secret: secret, parser: parser}
}

// Parse extracts userID/email/role from a signed token, or returns an error
// if the token is invalid (signature, expiry, alg, missing claims).
func (v *Validator) Parse(tokenString string) (*Claims, error) {
	token, err := v.parser.Parse(tokenString, func(t *jwtlib.Token) (any, error) {
		// Belt-and-braces: WithValidMethods already enforces HS256, but we
		// double-check so a future relaxation of the parser config can't
		// silently accept a different family.
		if _, ok := t.Method.(*jwtlib.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(v.secret), nil
	})
	if err != nil {
		return nil, fmt.Errorf("parse token: %w", err)
	}
	if !token.Valid {
		return nil, errors.New("invalid token")
	}
	mapClaims, ok := token.Claims.(jwtlib.MapClaims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}

	userID, err := userIDFromClaims(mapClaims)
	if err != nil {
		return nil, err
	}
	email, _ := mapClaims["email"].(string)
	role, _ := mapClaims["role"].(string)
	return &Claims{UserID: userID, Email: email, Role: role}, nil
}

func userIDFromClaims(c jwtlib.MapClaims) (uint64, error) {
	if v, err := uintClaim(c, "user_id"); err == nil {
		return v, nil
	}
	// Backward compat with old tokens that only had `sub`.
	if v, err := uintClaim(c, "sub"); err == nil {
		return v, nil
	}
	return 0, errors.New("user_id not found in token claims")
}

func uintClaim(c jwtlib.MapClaims, key string) (uint64, error) {
	raw, ok := c[key]
	if !ok {
		return 0, fmt.Errorf("claim %q not found", key)
	}
	var v uint64
	switch x := raw.(type) {
	case float64:
		v = uint64(x)
	case int64:
		v = uint64(x)
	case uint64:
		v = x
	case string:
		parsed, err := strconv.ParseUint(x, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("invalid %q: %w", key, err)
		}
		v = parsed
	default:
		return 0, fmt.Errorf("unexpected %q type: %T", key, x)
	}
	if v == 0 {
		return 0, errors.New("user_id cannot be zero")
	}
	return v, nil
}
