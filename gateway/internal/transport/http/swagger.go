package http

import (
	"bytes"
	"html/template"
	"log/slog"
	"net/http"
)

// SwaggerSpec is one service's OpenAPI 3.0 document loaded into memory at
// boot. Slug is the URL path component (/openapi/<slug>.yaml); Title is
// the human-readable label Scalar shows in its source dropdown.
type SwaggerSpec struct {
	Slug    string
	Title   string
	Content []byte
}

// SwaggerHandler returns a handler tree that serves:
//   - /openapi/<slug>.yaml  for each spec — raw YAML bytes
//   - /docs                 — Scalar UI initialised with a dropdown over
//     every spec
//
// Specs are loaded once at boot (see bootstrap.InitSwaggerSpecs) so each
// request is a static write; per-request file I/O is avoided.
func SwaggerHandler(specs []SwaggerSpec) http.Handler {
	mux := http.NewServeMux()

	// Each spec gets its own URL so Scalar can fetch them independently.
	for _, s := range specs {
		spec := s // capture for the closure
		mux.HandleFunc("/openapi/"+spec.Slug+".yaml", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/yaml")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(spec.Content)
		})
	}

	docsHTML, err := renderDocsHTML(specs)
	if err != nil {
		// Should never happen — template is a constant and specs are
		// validated at boot. If it does, fall back to a minimal page so
		// /openapi/<slug>.yaml still works for direct API consumption.
		slog.Error("render docs html", "err", err)
		docsHTML = []byte("<!doctype html><body><pre>docs unavailable, see /openapi/&lt;svc&gt;.yaml</pre></body>")
	}
	mux.HandleFunc("/docs", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(docsHTML)
	})

	return mux
}

// docsTemplate is a Scalar UI shell. The Sources slice lands inside a
// <script> block — html/template's context-aware escaping treats the JS
// payload as JS data, so the Go slice marshals to a safe JS literal.
//
// CDN-hosted Scalar; pinned to @latest because we don't ship the bundle
// in our image. Flip to a fixed major if a breaking change ever surfaces.
var docsTemplate = template.Must(template.New("docs").Parse(`<!DOCTYPE html>
<html>
<head>
    <title>Gateway API Documentation</title>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
    <style>body { margin: 0; }</style>
</head>
<body>
    <div id="app"></div>
    <script src="https://cdn.jsdelivr.net/npm/@scalar/api-reference"></script>
    <script>
        Scalar.createApiReference('#app', {
            sources: {{ .Sources }}
        })
    </script>
</body>
</html>`))

// docsSource mirrors what Scalar's `sources` array expects on the JS side.
// Title is the dropdown label; URL is where Scalar fetches the OpenAPI
// document — we point it at our /openapi/<slug>.yaml.
type docsSource struct {
	Title string `json:"title"`
	URL   string `json:"url"`
}

func renderDocsHTML(specs []SwaggerSpec) ([]byte, error) {
	sources := make([]docsSource, 0, len(specs))
	for _, s := range specs {
		sources = append(sources, docsSource{
			Title: s.Title,
			URL:   "/openapi/" + s.Slug + ".yaml",
		})
	}
	var buf bytes.Buffer
	if err := docsTemplate.Execute(&buf, struct{ Sources []docsSource }{sources}); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
