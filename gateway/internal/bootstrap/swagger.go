package bootstrap

import (
	"encoding/json"
	"fmt"
	"os"
)

// swaggerSpecPaths lists the per-service OpenAPI v2 specs grpc-gateway
// emits during code generation. Order is irrelevant — mergeSwaggerFiles
// is order-stable for paths/definitions/securityDefinitions.
var swaggerSpecPaths = []string{
	"./internal/pb/swagger/auth_api/auth.swagger.json",
	"./internal/pb/swagger/vacancy_api/vacancy.swagger.json",
	"./internal/pb/swagger/resume_api/resume.swagger.json",
	"./internal/pb/swagger/analysis_api/analysis.swagger.json",
}

// InitSwaggerSpec reads the per-service swagger files and stitches them
// into one document the docs UI can render. Performed once at boot so
// the per-request /swagger.json handler is just a static read.
func InitSwaggerSpec() ([]byte, error) {
	return mergeSwaggerFiles(swaggerSpecPaths...)
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
			return nil, fmt.Errorf("read swagger %s: %w", path, err)
		}

		var spec map[string]any
		if err := json.Unmarshal(payload, &spec); err != nil {
			return nil, fmt.Errorf("parse swagger %s: %w", path, err)
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
