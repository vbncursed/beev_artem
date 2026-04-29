package domain

// CandidateProfile is the value object the profile extractor produces from a
// raw resume text + filename. It lives in domain because it represents a
// concept the use case manipulates by name (FullName, Email, Phone) — the
// extractor adapter merely fills it in.
type CandidateProfile struct {
	FullName string
	Email    string
	Phone    string
}

// SourceFor returns the canonical "source" string for a candidate created
// from a resume upload. The rule (vacancy → resume_auto, no vacancy →
// resume_pool) is a domain rule, not a storage concern, so it lives here
// instead of inside use case as a free function.
func SourceFor(vacancyID string) string {
	if vacancyID == "" {
		return "resume_pool"
	}
	return "resume_auto"
}

// BelongsTo reports whether this candidate is owned by userID, or whether
// the caller has admin privileges. Encapsulates the access-control rule so
// transport / use case never need to inspect OwnerUserID + role separately.
func (c *Candidate) BelongsTo(userID uint64, isAdmin bool) bool {
	if c == nil {
		return false
	}
	return isAdmin || c.OwnerUserID == userID
}
