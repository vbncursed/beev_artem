package vacancy_service_api

import (
	"encoding/json"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	ErrCodeInvalidInput = "INVALID_INPUT"
	ErrCodeUnauthorized = "UNAUTHORIZED"
	ErrCodeNotFound     = "NOT_FOUND"
	ErrCodeInternal     = "INTERNAL_ERROR"
)

type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Field   string `json:"field,omitempty"`
}

func newError(grpcCode codes.Code, errCode, message string) error {
	detail := ErrorDetail{Code: errCode, Message: message}
	jsonBytes, err := json.Marshal(detail)
	if err != nil {
		return status.Error(grpcCode, message)
	}
	return status.Error(grpcCode, string(jsonBytes))
}
