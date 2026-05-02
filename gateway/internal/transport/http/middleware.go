package http

import (
	"log/slog"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
)

// IncomingHeaderMatcher tells grpc-gateway which incoming HTTP headers to
// propagate as gRPC metadata. We allow `authorization` (bearer token —
// downstream services validate it themselves) and `x-client-ip`
// (downstream rate-limit bucketing). Everything else falls back to
// grpc-gateway's default matcher (which keeps `x-*` and a few standard
// headers).
//
// Historical note: this used to also forward x-user-id / x-user-role /
// x-user-email — populated by withAuthContext after the gateway pre-
// validated the JWT. Downstream services no longer trust those headers
// (they call auth.ValidateAccessToken themselves), so propagating them
// would be misleading at best and a spoofing vector at worst.
func IncomingHeaderMatcher(key string) (string, bool) {
	switch strings.ToLower(key) {
	case "authorization", "x-client-ip":
		return key, true
	default:
		return runtime.DefaultHeaderMatcher(key)
	}
}

// WithCORS handles CORS preflight (OPTIONS) and decorates real responses
// with Access-Control-Allow-* headers when the request Origin matches the
// allowlist. allowedOrigins=["*"] enables a permissive mode that echoes
// any origin BUT cannot be combined with credentials per the CORS spec —
// we do not set Access-Control-Allow-Credentials because the frontend
// does not send cookies (auth uses Authorization: Bearer).
//
// Empty allowedOrigins disables CORS (handler is a no-op pass-through).
func WithCORS(allowedOrigins []string) func(http.Handler) http.Handler {
	if len(allowedOrigins) == 0 {
		return func(next http.Handler) http.Handler { return next }
	}
	wildcard := len(allowedOrigins) == 1 && allowedOrigins[0] == "*"
	allowed := make(map[string]struct{}, len(allowedOrigins))
	for _, o := range allowedOrigins {
		allowed[strings.ToLower(o)] = struct{}{}
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if origin != "" {
				if wildcard || originAllowed(origin, allowed) {
					if wildcard {
						w.Header().Set("Access-Control-Allow-Origin", "*")
					} else {
						w.Header().Set("Access-Control-Allow-Origin", origin)
					}
					w.Header().Add("Vary", "Origin")
					w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
					w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type, X-Client-IP, X-Requested-With")
					w.Header().Set("Access-Control-Max-Age", "600")
				}
			}
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func originAllowed(origin string, allowed map[string]struct{}) bool {
	_, ok := allowed[strings.ToLower(origin)]
	return ok
}

// WithJSONContentType backfills Content-Type: application/json on writes
// that arrive without one. grpc-gateway rejects bodies it can't parse,
// and curl-style callers routinely forget the header.
func WithJSONContentType(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Type") == "" && (r.Method == http.MethodPost || r.Method == http.MethodPatch || r.Method == http.MethodPut) {
			r.Header.Set("Content-Type", "application/json")
		}
		next.ServeHTTP(w, r)
	})
}

// WithLogging records one access-log line per request. Centralised so
// handlers don't need to emit slog calls for things derivable from
// method+path+duration.
func WithLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startedAt := time.Now()
		next.ServeHTTP(w, r)
		slog.Info("gateway request", "method", r.Method, "path", r.URL.Path, "duration", time.Since(startedAt).String())
	})
}

// trustForwardedFor toggles X-Forwarded-For trust. Flip to true ONLY when
// the gateway sits behind a known proxy / LB that strips and sets the
// header. In direct-edge deployments leaving it false avoids a trivial
// client-IP spoofing vector — the rate-limit bucket downstream uses
// X-Client-IP, so a spoofed value would let an attacker bypass per-IP
// limits.
const trustForwardedFor = false

// WithClientIP populates X-Client-IP from r.RemoteAddr (or X-Forwarded-For
// if trustForwardedFor is true). Downstream services consume this header
// for rate-limit bucketing — without it every request carries the gateway
// container's IP and shares one bucket.
func WithClientIP(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := remoteIP(r.RemoteAddr)
		if trustForwardedFor {
			if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
				if first, _, ok := strings.Cut(xff, ","); ok {
					ip = strings.TrimSpace(first)
				} else {
					ip = strings.TrimSpace(xff)
				}
			}
		}
		if ip != "" {
			r.Header.Set("X-Client-IP", ip)
		}
		next.ServeHTTP(w, r)
	})
}

func remoteIP(addr string) string {
	if addr == "" {
		return ""
	}
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return addr
	}
	return host
}
