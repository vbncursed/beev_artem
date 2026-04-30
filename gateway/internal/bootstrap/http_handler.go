package bootstrap

import (
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"

	"github.com/artem13815/hr/gateway/internal/pb/auth_api"
	transport_http "github.com/artem13815/hr/gateway/internal/transport/http"
)

// InitHTTPHandler composes the root handler tree:
//
//	withLogging
//	  └─ withClientIP
//	        └─ withJSONContentType
//	              └─ rootMux:
//	                    /healthz       -> HealthHandler
//	                    /swagger.json  -> SwaggerHandler
//	                    /docs          -> SwaggerHandler
//	                    /              -> withAuthContext(grpc-gateway mux)
//
// Order of decoration matters: logging is outermost so 401-out-at-edge
// requests still get an access log line; clientIP runs before auth so
// the auth check sees the populated header in case it ever needs IP-
// scoped rate limiting.
func InitHTTPHandler(authClient auth_api.AuthServiceClient, gwMux *runtime.ServeMux, swaggerSpec []byte) http.Handler {
	root := http.NewServeMux()

	swaggerH := transport_http.SwaggerHandler(swaggerSpec)
	root.Handle("/swagger.json", swaggerH)
	root.Handle("/docs", swaggerH)
	root.Handle("/healthz", transport_http.HealthHandler())
	root.Handle("/", transport_http.WithAuthContext(authClient, gwMux))

	return transport_http.WithLogging(
		transport_http.WithClientIP(
			transport_http.WithJSONContentType(root),
		),
	)
}
