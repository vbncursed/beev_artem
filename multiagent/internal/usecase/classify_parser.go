package usecase

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"
)

// roleClassification is the wire-shape of the classifier's JSON output.
// Decoupled from domain.RoleClassifyResponse so a chatty model that adds
// extra fields doesn't break the response contract — the parser only
// extracts the field it cares about.
type roleClassification struct {
	Role string `json:"role"`
}

// parseRoleClassification turns the model's text into a validated role.
// Returns ErrLLMInvalidResponse when the JSON is malformed or the role
// is outside the allowed set. The allowed set is ListRoles() ∪ {"default"} —
// "default" is always permitted because it's the explicit "unsure" answer.
//
// Role matching is case-insensitive and trimmed: "Programmer", "  PROGRAMMER ",
// and "programmer" all normalize to the same canonical form.
func parseRoleClassification(completion string, allowedRoles []string) (string, error) {
	raw := stripJSONFences(completion)
	var d roleClassification
	if err := json.Unmarshal([]byte(raw), &d); err != nil {
		return "", fmt.Errorf("%w: classify unmarshal: %v", ErrLLMInvalidResponse, err)
	}

	role := strings.ToLower(strings.TrimSpace(d.Role))
	if role == "" {
		return "", fmt.Errorf("%w: empty role", ErrLLMInvalidResponse)
	}
	if role == defaultRole {
		return defaultRole, nil
	}
	if slices.Contains(allowedRoles, role) {
		return role, nil
	}
	return "", fmt.Errorf("%w: unknown role %q", ErrLLMInvalidResponse, d.Role)
}
