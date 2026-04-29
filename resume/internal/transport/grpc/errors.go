package grpc

import (
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	ErrCodeInvalidInput = "INVALID_INPUT"
	ErrCodeUnauthorized = "UNAUTHORIZED"
	ErrCodeNotFound     = "NOT_FOUND"
	ErrCodeInternal     = "INTERNAL_ERROR"

	// errorDomain identifies which service surface produced the error. Clients
	// that aggregate errors from multiple services use Domain to disambiguate
	// the same Reason coming from different sources.
	errorDomain = "resume.service.v1"
)

// newError returns a gRPC status with a structured ErrorInfo detail. Clients
// read `Reason` (our errCode) and `Domain` instead of parsing JSON out of the
// message field — the canonical gRPC convention for machine-readable errors.
//
// If attaching the detail somehow fails (proto marshaling — should not happen
// for ErrorInfo), we fall back to a plain status with the human message so
// the caller still gets the right code.
func newError(grpcCode codes.Code, errCode, message string) error {
	st := status.New(grpcCode, message)
	withDetails, err := st.WithDetails(&errdetails.ErrorInfo{
		Reason: errCode,
		Domain: errorDomain,
	})
	if err != nil {
		return st.Err()
	}
	return withDetails.Err()
}
