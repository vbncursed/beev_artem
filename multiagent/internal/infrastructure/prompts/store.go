// Package prompts implements the usecase.PromptStore port by serving
// role-specific prompt templates baked into the binary at compile time.
//
// Adding a new role is a one-file commit: drop assets/prompts/<role>.txt,
// rebuild. No schema migration, no config change, no proto bump.
package prompts

import (
	"embed"
	"path"
	"slices"
	"strings"
)

// embedded is the prompt set baked into the binary. The directory layout
// is intentionally flat — `<role>.txt` — so the lookup is a single Open
// call, no globbing.
//
//go:embed templates/*.txt
var embedded embed.FS

// Store reads role prompts from the embedded FS, with a default fallback
// when the requested role has no dedicated prompt. Stateless and safe to
// share across goroutines — embed.FS is read-only.
type Store struct {
	defaultPrompt string
	byRole        map[string]string
}

// New eagerly loads all prompt files from assets/prompts/ into memory.
// Eager so every Get is a map hit (no file I/O on the hot path) and a
// missing default.txt fails the boot loudly instead of silently falling
// back at request time.
func New() (*Store, error) {
	def, err := readPrompt("default")
	if err != nil {
		return nil, err
	}
	entries, err := embedded.ReadDir("templates")
	if err != nil {
		return nil, err
	}
	by := make(map[string]string, len(entries))
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".txt") {
			continue
		}
		role := strings.TrimSuffix(name, ".txt")
		body, err := readPrompt(role)
		if err != nil {
			return nil, err
		}
		by[role] = body
	}
	return &Store{defaultPrompt: def, byRole: by}, nil
}

func readPrompt(role string) (string, error) {
	b, err := embedded.ReadFile(path.Join("templates", role+".txt"))
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// Get returns the prompt for the requested role. Empty role and unknown
// roles both fall back to default — the contract usecase relies on. Role
// matching is case-insensitive so "Programmer" / "PROGRAMMER" / "programmer"
// all hit the same template.
func (s *Store) Get(role string) string {
	if role == "" {
		return s.defaultPrompt
	}
	if p, ok := s.byRole[strings.ToLower(strings.TrimSpace(role))]; ok {
		return p
	}
	return s.defaultPrompt
}

// ListRoles returns the registered role names sorted alphabetically. The
// "default" fallback is excluded — it's not a real classification target,
// just the catch-all when no specific role applies. The classifier embeds
// this list into its system prompt so adding a new templates/<role>.txt
// is enough to extend the model's vocabulary.
func (s *Store) ListRoles() []string {
	roles := make([]string, 0, len(s.byRole))
	for r := range s.byRole {
		if r == "default" {
			continue
		}
		roles = append(roles, r)
	}
	slices.Sort(roles)
	return roles
}
