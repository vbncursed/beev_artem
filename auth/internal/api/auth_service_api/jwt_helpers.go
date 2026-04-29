package auth_service_api

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc/metadata"
)

type tokenClaims struct {
	UserID uint64
	Email  string
	Role   string
}

func parseTokenClaims(tokenString, jwtSecret string) (*tokenClaims, error) {
	claims, err := parseJWTToken(tokenString, jwtSecret)
	if err != nil {
		return nil, err
	}
	return parseClaims(claims)
}

func parseClaims(claims jwt.MapClaims) (*tokenClaims, error) {
	userID, err := extractUserIDFromClaims(claims)
	if err != nil {
		return nil, err
	}

	email, _ := claims["email"].(string)
	role, _ := claims["role"].(string)

	return &tokenClaims{
		UserID: userID,
		Email:  email,
		Role:   role,
	}, nil
}

func extractUserIDFromClaims(claims jwt.MapClaims) (uint64, error) {
	userID, err := parseUintClaim(claims, "user_id")
	if err == nil {
		return userID, nil
	}

	// Backward compatibility with old tokens that only contain "sub".
	userID, subErr := parseUintClaim(claims, "sub")
	if subErr != nil {
		return 0, errors.New("user_id not found in token claims")
	}
	return userID, nil
}

func parseUintClaim(claims jwt.MapClaims, key string) (uint64, error) {
	raw, ok := claims[key]
	if !ok {
		return 0, fmt.Errorf("claim %q not found", key)
	}

	var userID uint64
	switch v := raw.(type) {
	case float64:
		userID = uint64(v)
	case int64:
		userID = uint64(v)
	case uint64:
		userID = v
	case string:
		parsed, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("invalid user_id format in claim %q: %w", key, err)
		}
		userID = parsed
	default:
		return 0, fmt.Errorf("unexpected user_id type in claim %q: %T", key, v)
	}

	if userID == 0 {
		return 0, errors.New("user_id cannot be zero")
	}

	return userID, nil
}

func extractTokenFromContext(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", errors.New("metadata not found in context")
	}

	var tokenString string
	if authHeaders := md.Get("authorization"); len(authHeaders) > 0 {
		authHeader := authHeaders[0]
		if scheme, value, ok := strings.Cut(authHeader, " "); ok && strings.EqualFold(scheme, "bearer") {
			tokenString = value
		} else {
			tokenString = authHeader
		}
	}

	if tokenString == "" {
		return "", errors.New("authorization token not found in context")
	}

	return tokenString, nil
}

// jwtParser is built once and reused: it pins the algorithm to HS256 (defense
// against alg-confusion / "none" attacks even if the keyfunc is bypassed) and
// requires `exp` to be present so a token without an expiry is rejected outright
// instead of being accepted as eternal.
var jwtParser = jwt.NewParser(
	jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
	jwt.WithExpirationRequired(),
)

func parseJWTToken(tokenString, jwtSecret string) (jwt.MapClaims, error) {
	token, err := jwtParser.Parse(tokenString, func(token *jwt.Token) (any, error) {
		// Belt-and-braces: WithValidMethods already enforces HS256, but we
		// double-check so a future relaxation of the parser config can't
		// silently accept a different family.
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(jwtSecret), nil
	})
	if err != nil {
		return nil, fmt.Errorf("parse token: %w", err)
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}

	return claims, nil
}
