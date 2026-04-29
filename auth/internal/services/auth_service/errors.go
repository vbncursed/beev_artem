package auth_service

import "errors"

var (
	ErrInvalidEmail        = errors.New("invalid email")
	ErrInvalidPassword     = errors.New("invalid password")
	ErrInvalidArgument     = errors.New("invalid argument")
	ErrEmailAlreadyExists  = errors.New("email already exists")
	ErrInvalidCredentials  = errors.New("invalid credentials")
	ErrInvalidRefreshToken = errors.New("invalid refresh token")
	ErrSessionExpired      = errors.New("session expired")
	ErrSessionRevoked      = errors.New("session revoked")
	ErrSessionNotFound     = errors.New("session not found")
	ErrInvalidRole         = errors.New("invalid role")
	ErrPermissionDenied    = errors.New("permission denied")
	ErrCannotChangeOwnRole = errors.New("cannot change own role")
	ErrUserNotFound        = errors.New("user not found")
)
