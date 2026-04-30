package http

import (
	"fmt"
	"net/http"
)

// SwaggerHandler returns a sub-mux that serves the merged OpenAPI v2 spec
// at /swagger.json and a Scalar-rendered docs page at /docs. The spec is
// pre-rendered at boot (see bootstrap.InitSwaggerSpec) so each request is
// just a static read — no per-request file I/O.
func SwaggerHandler(spec []byte) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/swagger.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(spec)
	})
	mux.HandleFunc("/docs", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, scalarDocsHTML)
	})
	return mux
}

// scalarDocsHTML renders the OpenAPI spec via Scalar's CDN-hosted
// reference UI. Pinned to @latest for now — flip to a fixed major if a
// breaking change ever surfaces in their JS bundle.
const scalarDocsHTML = `<!doctype html>
<html>
  <head>
    <title>Gateway API Reference</title>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
    <style>body { margin: 0; }</style>
  </head>
  <body>
    <script id="api-reference" data-url="/swagger.json"></script>
    <script src="https://cdn.jsdelivr.net/npm/@scalar/api-reference@latest"></script>
  </body>
</html>`
