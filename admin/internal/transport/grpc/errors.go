package grpc

import (
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// errdetails ErrorInfo Reason values surfaced over the wire so frontend
// can switch on them without parsing free-form messages.
const (
	ErrCodeUnauthorized = "UNAUTHORIZED"
	ErrCodeForbidden    = "FORBIDDEN"
	ErrCodeInvalidInput = "INVALID_INPUT"
	ErrCodeNotFound     = "NOT_FOUND"
	ErrCodeInternal     = "INTERNAL"

	errDomain = "admin.service.v1"
)

func newError(code codes.Code, reason, msg string) error {
	st := status.New(code, msg)
	withDetails, err := st.WithDetails(&errdetails.ErrorInfo{
		Reason: reason,
		Domain: errDomain,
	})
	if err != nil {
		return st.Err()
	}
	return withDetails.Err()
}
