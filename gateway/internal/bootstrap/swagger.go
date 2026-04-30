package bootstrap

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	transport_http "github.com/artem13815/hr/gateway/internal/transport/http"
)

// openAPIDir holds one YAML file per backend service. The protoc-gen-openapi
// (gnostic) plugin writes <svc>.yaml per generation and the gateway serves
// each at /openapi/<svc>.yaml so Scalar UI can render a dropdown over them.
const openAPIDir = "./internal/pb/openapi"

// titleOverrides give each known slug a human-readable title in the Scalar
// dropdown. Unknown slugs fall back to "<Slug> API" — drop a new file in
// the dir and it appears automatically with a sane default.
var titleOverrides = map[string]string{
	"auth":     "Auth API",
	"vacancy":  "Vacancy API",
	"resume":   "Resume API",
	"analysis": "Analysis API",
}

// InitSwaggerSpecs eagerly loads every per-service OpenAPI YAML in the
// openapi dir. Eager so a missing/unreadable file fails the boot loudly
// instead of erroring on the first /openapi/<svc>.yaml request, and so
// the per-request handler is just a static byte write.
//
// Sort order is alphabetical by slug — gives the dropdown a stable order
// across boots regardless of filesystem listing order.
func InitSwaggerSpecs() ([]transport_http.SwaggerSpec, error) {
	entries, err := os.ReadDir(openAPIDir)
	if err != nil {
		return nil, fmt.Errorf("read openapi dir %s: %w", openAPIDir, err)
	}
	specs := make([]transport_http.SwaggerSpec, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasSuffix(name, ".yaml") {
			continue
		}
		slug := strings.TrimSuffix(name, ".yaml")
		body, err := os.ReadFile(filepath.Join(openAPIDir, name))
		if err != nil {
			return nil, fmt.Errorf("read openapi spec %s: %w", name, err)
		}
		specs = append(specs, transport_http.SwaggerSpec{
			Slug:    slug,
			Title:   titleFor(slug),
			Content: body,
		})
	}
	sort.Slice(specs, func(i, j int) bool { return specs[i].Slug < specs[j].Slug })
	if len(specs) == 0 {
		return nil, fmt.Errorf("no openapi specs found in %s — did you run `make generate-api`?", openAPIDir)
	}
	return specs, nil
}

func titleFor(slug string) string {
	if t, ok := titleOverrides[slug]; ok {
		return t
	}
	if slug == "" {
		return "API"
	}
	return strings.ToUpper(slug[:1]) + slug[1:] + " API"
}
