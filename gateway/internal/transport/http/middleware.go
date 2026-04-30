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
