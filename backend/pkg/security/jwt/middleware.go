package jwt

import (
	"net/http"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

// NewAuthMiddleware returns a Fiber middleware that validates Bearer JWT (HS256).
// On success sets user id (subject) into c.Locals("userId").
func NewAuthMiddleware(secret, expectedIssuer string) fiber.Handler {
	secretBytes := []byte(secret)
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"message": "missing Authorization header"})
		}
		// Support both "Bearer <token>" and "<token>" (no prefix).
		var tokenStr string
		if strings.Contains(authHeader, " ") {
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
				tokenStr = strings.TrimSpace(parts[1])
			} else {
				// Fallback: treat entire header as token (for non-standard clients)
				tokenStr = strings.TrimSpace(authHeader)
			}
		} else {
			tokenStr = strings.TrimSpace(authHeader)
		}
		if tokenStr == "" {
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"message": "empty token"})
		}
		token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (any, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fiber.ErrUnauthorized
			}
			return secretBytes, nil
		}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Name}))
		if err != nil || !token.Valid {
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"message": "invalid or expired token"})
		}
		claims, ok := token.Claims.(*Claims)
		if !ok {
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"message": "invalid token claims"})
		}
		if expectedIssuer != "" && claims.RegisteredClaims.Issuer != expectedIssuer {
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"message": "invalid token issuer"})
		}
		c.Locals("userId", claims.RegisteredClaims.Subject)
		if claims.IsAdmin {
			c.Locals("isAdmin", true)
		}
		return c.Next()
	}
}
