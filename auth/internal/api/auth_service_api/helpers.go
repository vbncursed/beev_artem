package auth_service_api

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net"
	"strings"

	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
)

// Metadata header that the gateway (or any upstream proxy) sets to forward the
// real client IP. Without it, peer.FromContext only sees the gateway container's
// address, which would collapse all traffic into a single rate-limit bucket.
const mdClientIP = "x-client-ip"

func isDatabaseError(err error) bool {
	if err == nil {
		return false
	}

	errStr := strings.ToLower(err.Error())

	if _, ok := errors.AsType[net.Error](err); ok {
		return true
	}

	databaseErrors := []string{
		"connection refused",
		"connection reset",
		"no such host",
		"network is unreachable",
		"timeout",
		"dial tcp",
		"connection closed",
		"broken pipe",
		"database",
		"postgres",
		"pgx",
		"pool",
		"unable to connect",
		"connection failed",
	}

	for _, dbErr := range databaseErrors {
		if strings.Contains(errStr, dbErr) {
			return true
		}
	}

	return false
}

func clientMeta(ctx context.Context) (userAgent, ip string) {
	md, _ := metadata.FromIncomingContext(ctx)

	if md != nil {
		if ua := md.Get("user-agent"); len(ua) > 0 {
			userAgent = ua[0]
		}
		if forwarded := md.Get(mdClientIP); len(forwarded) > 0 {
			ip = strings.TrimSpace(forwarded[0])
		}
	}

	if ip == "" {
		ip = peerIP(ctx)
	}

	if ip == "" {
		ip = "unknown"
	}

	return userAgent, ip
}

func peerIP(ctx context.Context) string {
	p, ok := peer.FromContext(ctx)
	if !ok || p.Addr == nil {
		return ""
	}

	addr := p.Addr.String()

	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		if idx := strings.LastIndex(addr, ":"); idx != -1 {
			host = addr[:idx]
		} else {
			host = addr
		}
	}

	return strings.Trim(host, "[]")
}

// emailRateKey returns a stable, low-cardinality bucket key for per-email
// rate limiting. We hash so that we don't store raw email addresses in Redis
// keys (which would be PII visible in any monitoring dashboard).
func emailRateKey(email string) string {
	normalized := strings.ToLower(strings.TrimSpace(email))
	if normalized == "" {
		return "empty"
	}
	sum := sha256.Sum256([]byte(normalized))
	return hex.EncodeToString(sum[:8])
}
