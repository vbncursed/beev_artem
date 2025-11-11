package jwt

import (
    "context"
    "time"

    "github.com/golang-jwt/jwt/v5"

    "github.com/artem13815/hr/pkg/auth"
)

type Generator struct {
    secret []byte
    issuer string
    ttl    time.Duration
}

func NewGenerator(secret, issuer string, ttl time.Duration) *Generator {
    return &Generator{secret: []byte(secret), issuer: issuer, ttl: ttl}
}

func (g *Generator) Generate(ctx context.Context, user auth.User) (string, error) {
    now := time.Now().UTC()
    claims := jwt.RegisteredClaims{
        Issuer:    g.issuer,
        Subject:   user.ID.String(),
        IssuedAt:  jwt.NewNumericDate(now),
        ExpiresAt: jwt.NewNumericDate(now.Add(g.ttl)),
    }
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString(g.secret)
}


