package domain

// RoleClassifyRequest carries the raw vacancy text the LLM should classify.
// Kept minimal on purpose — the classifier only needs the human-readable
// description; the structured fields (skills, salary, …) live in the
// vacancy service and add no signal for role detection.
type RoleClassifyRequest struct {
	Title       string
	Description string
}

// RoleClassifyResponse holds the role label resolved by the LLM. The usecase
// guarantees the value is one of PromptStore.ListRoles() ∪ {"default"} — the
// transport layer can pass it through as-is.
type RoleClassifyResponse struct {
	Role string
}
