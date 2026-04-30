package http

import "net/http"

// HealthHandler returns a 200 OK with no body. compose/k8s probe this
// endpoint to learn the gateway HTTP listener is up. A more elaborate
// readiness check could verify the auth-client connection, but that
// would couple liveness to a downstream — which is exactly what
// k8s readiness probes warn against.
func HealthHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}
