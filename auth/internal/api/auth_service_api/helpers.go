package auth_service_api

import (
	"context"
	"errors"
	"net"
	"strings"

	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
)

func isDatabaseError(err error) bool {
	if err == nil {
		return false
	}

	errStr := strings.ToLower(err.Error())

	var netErr net.Error
	if errors.As(err, &netErr) {
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
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		ua := md.Get("user-agent")
		if len(ua) > 0 {
			userAgent = ua[0]
		}
	}

	p, ok := peer.FromContext(ctx)
	if ok && p.Addr != nil {
		addr := p.Addr.String()

		host, _, err := net.SplitHostPort(addr)
		if err != nil {
			if idx := strings.LastIndex(addr, ":"); idx != -1 {
				host = addr[:idx]
			} else {
				host = addr
			}
		}

		host = strings.Trim(host, "[]")

		if host != "" {
			ip = host
		}
	}

	if ip == "" {
		ip = "unknown"
	}

	return userAgent, ip
}

