package http

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/artem13815/hr/gateway/internal/pb/auth_api"
	pb_models "github.com/artem13815/hr/gateway/internal/pb/models"
)

// authValidationTimeout caps an upstream ValidateAccessToken RPC. Each
// gated request pays this latency once at the edge.
const authValidationTimeout = 3 * time.Second

// WithAuthContext is the gateway-side fast-fail auth check. It calls
// auth.ValidateAccessToken before letting a request reach a backend so a
// bad token returns 401 in one hop instead of four.
//
// Backend services validate the JWT themselves too (defense in depth) —
// gateway only gates, it does NOT inject identity headers. The bearer
// flows through to backends via grpc-gateway's metadata propagation
// (Authorization is in IncomingHeaderMatcher's allowlist).
func WithAuthContext(authClient auth_api.AuthServiceClient, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !requiresAuth(r.Method, r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}

		token := extractBearerToken(r.Header.Get("Authorization"))
		if token == "" {
			writeUnauthorized(w, "missing bearer token")
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), authValidationTimeout)
		defer cancel()

		res, err := authClient.ValidateAccessToken(ctx, &pb_models.ValidateAccessTokenRequest{
			AccessToken: token,
		})
		if err != nil || !res.GetValid() || res.GetUserId() == 0 {
			writeUnauthorized(w, "invalid access token")
			return
		}

		next.ServeHTTP(w, r)
	})
}

// requiresAuth lists the paths that need a valid bearer at the edge.
// Auth-issuing endpoints (/login, /register, /refresh) MUST stay public —
// listing them here would create a chicken-and-egg problem.
func requiresAuth(method, path string) bool {
	switch {
	case strings.HasPrefix(path, "/api/v1/vacancies"):
		return true
	case strings.HasPrefix(path, "/api/v1/candidates"):
		return true
	case strings.HasPrefix(path, "/api/v1/resumes"):
		return true
	case strings.HasPrefix(path, "/api/v1/analyses"):
		return true
	case strings.HasPrefix(path, "/api/v1/admin"):
		return true
	case method == http.MethodGet && path == "/api/v1/auth/me":
		return true
	case method == http.MethodPost && (path == "/api/v1/auth/logout" || path == "/api/v1/auth/logout-all"):
		return true
	}
	return false
}

// extractBearerToken pulls the JWT out of an Authorization header. Falls
// back to treating the header as a raw token if no scheme prefix is
// present — historically some clients sent the token directly.
func extractBearerToken(authHeader string) string {
	if authHeader == "" {
		return ""
	}
	if scheme, value, ok := strings.Cut(authHeader, " "); ok && strings.EqualFold(scheme, "bearer") {
		return strings.TrimSpace(value)
	}
	return strings.TrimSpace(authHeader)
}

// writeUnauthorized returns a JSON 401 with a structured error code
// matching the convention backend services use (Reason+Domain via
// errdetails). The gateway writes plain JSON because the client at this
// stage hasn't begun a gRPC interaction.
func writeUnauthorized(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"code":    "UNAUTHORIZED",
		"message": message,
	})
}
