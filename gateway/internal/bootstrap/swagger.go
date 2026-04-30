package bootstrap

import (
	"encoding/json"
	"fmt"
	"os"

	"go.yaml.in/yaml/v4"
)

// openAPISpecPath is the merged OpenAPI 3.0 document gnostic emits during
// code generation. A single protoc-gen-openapi invocation across all four
// service protos produces this one file already merged — so this layer
// just reads it and converts YAML to JSON for the /swagger.json endpoint.
const openAPISpecPath = "./internal/pb/openapi/openapi.yaml"

// InitSwaggerSpec loads the embedded merged OpenAPI v3 document and
// returns it as JSON bytes. Performed once at boot so the per-request
// /swagger.json handler is just a static read.
//
// We do YAML -> JSON here (rather than committing JSON) because gnostic's
// canonical output format is YAML and forcing a hand conversion at
// generate time would drift on every regeneration.
func InitSwaggerSpec() ([]byte, error) {
	raw, err := os.ReadFile(openAPISpecPath)
	if err != nil {
		return nil, fmt.Errorf("read openapi spec %s: %w", openAPISpecPath, err)
	}
	var doc map[string]any
	if err := yaml.Unmarshal(raw, &doc); err != nil {
		return nil, fmt.Errorf("parse openapi spec %s: %w", openAPISpecPath, err)
	}
	out, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal openapi to json: %w", err)
	}
	return out, nil
}
