package auth_service_api

import (
	"encoding/json"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	ErrCodeInvalidEmail    = "INVALID_EMAIL"
	ErrCodeInvalidPassword = "INVALID_PASSWORD"
	ErrCodeInvalidInput    = "INVALID_INPUT"
	ErrCodeMissingField    = "MISSING_FIELD"

	ErrCodeInvalidCredentials = "INVALID_CREDENTIALS"
	ErrCodeInvalidToken       = "INVALID_TOKEN"
	ErrCodeTokenExpired       = "TOKEN_EXPIRED"
	ErrCodeTokenRevoked       = "TOKEN_REVOKED"
	ErrCodeSessionExpired     = "SESSION_EXPIRED"
	ErrCodeSessionRevoked     = "SESSION_REVOKED"

	ErrCodeEmailAlreadyExists = "EMAIL_ALREADY_EXISTS"

	ErrCodeRateLimitExceeded = "RATE_LIMIT_EXCEEDED"

	ErrCodeUnauthorized       = "UNAUTHORIZED"
	ErrCodeForbidden          = "FORBIDDEN"
	ErrCodeInternal           = "INTERNAL_ERROR"
	ErrCodeServiceUnavailable = "SERVICE_UNAVAILABLE"
)

type ErrorDetail struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Field   string            `json:"field,omitempty"`
	Meta    map[string]string `json:"meta,omitempty"`
}

func newError(grpcCode codes.Code, errCode, message string) error {
	detail := ErrorDetail{
		Code:    errCode,
		Message: message,
	}
	jsonBytes, err := json.Marshal(detail)
	if err != nil {
		return status.Error(grpcCode, message)
	}
	return status.Error(grpcCode, string(jsonBytes))
}

func newFieldError(grpcCode codes.Code, errCode, field, message string) error {
	detail := ErrorDetail{
		Code:    errCode,
		Message: message,
		Field:   field,
	}
	jsonBytes, err := json.Marshal(detail)
	if err != nil {
		return status.Error(grpcCode, message)
	}
	return status.Error(grpcCode, string(jsonBytes))
}

