package bootstrap

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/artem13815/hr/gateway/config"
	"github.com/artem13815/hr/gateway/internal/pb/analysis_api"
	"github.com/artem13815/hr/gateway/internal/pb/auth_api"
	pb_models "github.com/artem13815/hr/gateway/internal/pb/models"
	"github.com/artem13815/hr/gateway/internal/pb/resume_api"
	"github.com/artem13815/hr/gateway/internal/pb/vacancy_api"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func AppRun(cfg *config.Config) error {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	authConn, err := grpc.NewClient(cfg.Auth.GRPCAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("failed to create auth grpc client: %w", err)
	}
	defer authConn.Close()
	authClient := auth_api.NewAuthServiceClient(authConn)

	gwMux := runtime.NewServeMux(
		runtime.WithIncomingHeaderMatcher(incomingHeaderMatcher),
	)
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}

	if err := auth_api.RegisterAuthServiceHandlerFromEndpoint(ctx, gwMux, cfg.Auth.GRPCAddr, opts); err != nil {
		return fmt.Errorf("failed to register auth gateway handlers: %w", err)
	}
	if err := vacancy_api.RegisterVacancyServiceHandlerFromEndpoint(ctx, gwMux, cfg.Vacancy.GRPCAddr, opts); err != nil {
		return fmt.Errorf("failed to register vacancy gateway handlers: %w", err)
	}
	if err := resume_api.RegisterResumeServiceHandlerFromEndpoint(ctx, gwMux, cfg.Resume.GRPCAddr, opts); err != nil {
		return fmt.Errorf("failed to register resume gateway handlers: %w", err)
	}
	if err := analysis_api.RegisterAnalysisServiceHandlerFromEndpoint(ctx, gwMux, cfg.Analysis.GRPCAddr, opts); err != nil {
		return fmt.Errorf("failed to register analysis gateway handlers: %w", err)
	}

	authSwaggerPath := "./internal/pb/swagger/auth_api/auth.swagger.json"
	vacancySwaggerPath := "./internal/pb/swagger/vacancy_api/vacancy.swagger.json"
	resumeSwaggerPath := "./internal/pb/swagger/resume_api/resume.swagger.json"
	analysisSwaggerPath := "./internal/pb/swagger/analysis_api/analysis.swagger.json"
	mergedSwagger, err := mergeSwaggerFiles(authSwaggerPath, vacancySwaggerPath, resumeSwaggerPath, analysisSwaggerPath)
	if err != nil {
		return fmt.Errorf("failed to merge swagger specs: %w", err)
	}

	rootMux := http.NewServeMux()
	rootMux.HandleFunc("/swagger.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(mergedSwagger)
	})
	rootMux.HandleFunc("/docs", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `<!doctype html>
<html>
  <head>
    <title>Gateway API Reference</title>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
    <style>
      body { margin: 0; }
    </style>
  </head>
  <body>
    <script id="api-reference" data-url="/swagger.json"></script>
    <script src="https://cdn.jsdelivr.net/npm/@scalar/api-reference@latest"></script>
  </body>
</html>`)
	})
	rootMux.Handle("/", withAuthContext(authClient, gwMux))

	handler := withLogging(withClientIP(withJSONContentType(rootMux)))

	srv := &http.Server{
		Addr:              cfg.Server.HTTPAddr,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
	}

	slog.Info(
		"gateway server listening",
		"http_addr", cfg.Server.HTTPAddr,
		"auth_grpc_addr", cfg.Auth.GRPCAddr,
		"vacancy_grpc_addr", cfg.Vacancy.GRPCAddr,
		"resume_grpc_addr", cfg.Resume.GRPCAddr,
		"analysis_grpc_addr", cfg.Analysis.GRPCAddr,
	)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("gateway http server failed: %w", err)
	}

	return nil
}

func withJSONContentType(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Type") == "" && (r.Method == http.MethodPost || r.Method == http.MethodPatch || r.Method == http.MethodPut) {
			r.Header.Set("Content-Type", "application/json")
		}
		next.ServeHTTP(w, r)
	})
}

func withLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startedAt := time.Now()
		next.ServeHTTP(w, r)
		slog.Info("gateway request", "method", r.Method, "path", r.URL.Path, "duration", time.Since(startedAt).String())
	})
}

func incomingHeaderMatcher(key string) (string, bool) {
	switch strings.ToLower(key) {
	case "authorization", "x-user-id", "user-id", "x-user-role", "x-user-email", "x-client-ip":
		return key, true
	default:
		return runtime.DefaultHeaderMatcher(key)
	}
}

// withClientIP populates X-Client-IP from r.RemoteAddr (or X-Forwarded-For if
// the gateway sits behind a trusted proxy). The downstream gRPC services use
// this header for rate-limit bucketing — without it every request carries the
// gateway container's IP and they all share one bucket.
//
// SECURITY: trusting X-Forwarded-For is only safe when the gateway is behind a
// proxy that strips/sets it. In direct-edge deployments, set
// trustForwardedFor=false and rely solely on RemoteAddr.
func withClientIP(next http.Handler) http.Handler {
	const trustForwardedFor = false // flip to true when running behind a known LB/proxy
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

func withAuthContext(authClient auth_api.AuthServiceClient, next http.Handler) http.Handler {
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

		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()

		res, err := authClient.ValidateAccessToken(ctx, &pb_models.ValidateAccessTokenRequest{
			AccessToken: token,
		})
		if err != nil || !res.GetValid() || res.GetUserId() == 0 {
			writeUnauthorized(w, "invalid access token")
			return
		}

		r.Header.Set("x-user-id", strconv.FormatUint(res.GetUserId(), 10))
		r.Header.Set("x-user-role", res.GetRole())
		r.Header.Set("x-user-email", res.GetEmail())

		next.ServeHTTP(w, r)
	})
}

func requiresAuth(method, path string) bool {
	if strings.HasPrefix(path, "/api/v1/vacancies") {
		return true
	}
	if strings.HasPrefix(path, "/api/v1/candidates") || strings.HasPrefix(path, "/api/v1/resumes") {
		return true
	}
	if strings.HasPrefix(path, "/api/v1/analyses") {
		return true
	}
	if method == http.MethodGet && path == "/api/v1/auth/me" {
		return true
	}
	if method == http.MethodPost && (path == "/api/v1/auth/logout" || path == "/api/v1/auth/logout-all") {
		return true
	}
	return false
}

func extractBearerToken(authHeader string) string {
	if authHeader == "" {
		return ""
	}
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) == 2 && strings.EqualFold(parts[0], "bearer") {
		return strings.TrimSpace(parts[1])
	}
	return strings.TrimSpace(authHeader)
}

func writeUnauthorized(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"code":    "UNAUTHORIZED",
		"message": message,
	})
}

func mergeSwaggerFiles(paths ...string) ([]byte, error) {
	merged := map[string]any{
		"swagger":             "2.0",
		"info":                map[string]any{"title": "Gateway API", "version": "1.0.0", "description": "Unified gateway API docs"},
		"paths":               map[string]any{},
		"definitions":         map[string]any{},
		"tags":                []any{},
		"securityDefinitions": map[string]any{},
	}

	pathSet := merged["paths"].(map[string]any)
	defsSet := merged["definitions"].(map[string]any)
	secSet := merged["securityDefinitions"].(map[string]any)
	tagSet := map[string]struct{}{}
	consumes := map[string]struct{}{}
	produces := map[string]struct{}{}

	for _, path := range paths {
		payload, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}

		var spec map[string]any
		if err := json.Unmarshal(payload, &spec); err != nil {
			return nil, err
		}

		mergeMapAny(pathSet, spec["paths"])
		mergeMapAny(defsSet, spec["definitions"])
		mergeMapAny(secSet, spec["securityDefinitions"])

		if tags, ok := spec["tags"].([]any); ok {
			for _, raw := range tags {
				tag, ok := raw.(map[string]any)
				if !ok {
					continue
				}
				name, _ := tag["name"].(string)
				if name == "" {
					continue
				}
				if _, exists := tagSet[name]; exists {
					continue
				}
				tagSet[name] = struct{}{}
				merged["tags"] = append(merged["tags"].([]any), tag)
			}
		}

		collectStringArray(consumes, spec["consumes"])
		collectStringArray(produces, spec["produces"])
	}

	merged["consumes"] = setToArray(consumes)
	merged["produces"] = setToArray(produces)

	return json.MarshalIndent(merged, "", "  ")
}

func mergeMapAny(dst map[string]any, srcRaw any) {
	src, ok := srcRaw.(map[string]any)
	if !ok {
		return
	}
	for k, v := range src {
		dst[k] = v
	}
}

func collectStringArray(set map[string]struct{}, raw any) {
	arr, ok := raw.([]any)
	if !ok {
		return
	}
	for _, item := range arr {
		s, ok := item.(string)
		if !ok || s == "" {
			continue
		}
		set[s] = struct{}{}
	}
}

func setToArray(set map[string]struct{}) []any {
	out := make([]any, 0, len(set))
	for k := range set {
		out = append(out, k)
	}
	return out
}
