package persistence

import (
	"context"

	"github.com/artem13815/hr/resume/internal/domain"
)

// DeleteCandidate removes a candidate row. The resumes FK has ON DELETE
// CASCADE, so dependent resume rows disappear in the same statement.
// Returns ErrNotFound when no row matched (either missing or owned by a
// different user).
func (s *ResumeStorage) DeleteCandidate(
	ctx context.Context,
	candidateID string,
	requestUserID uint64,
	isAdmin bool,
) error {
	tag, err := s.db.Exec(ctx, `
DELETE FROM candidates
WHERE id = $1 AND ($2 OR owner_user_id = $3)
`, candidateID, isAdmin, requestUserID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}
