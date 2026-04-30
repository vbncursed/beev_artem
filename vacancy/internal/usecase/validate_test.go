package usecase

import (
	"testing"

	"gotest.tools/v3/assert"

	"github.com/artem13815/hr/vacancy/internal/domain"
)

// normalizeSkills is exercised end-to-end through Create/Update tests, but
// its branches deserve direct coverage too — the service treats its return
// value as authoritative input to storage, so wrong distribution here would
// silently corrupt every vacancy without tripping any other assertion.
func TestNormalizeSkills(t *testing.T) {
	tests := []struct {
		name string
		in   []domain.SkillWeight
		want []domain.SkillWeight
	}{
		{
			name: "empty slice returned as is",
			in:   nil,
			want: nil,
		},
		{
			name: "any positive weight short-circuits — input passes through",
			in: []domain.SkillWeight{
				{Name: "Go", Weight: 0.7},
				{Name: "SQL", Weight: 0},
			},
			want: []domain.SkillWeight{
				{Name: "Go", Weight: 0.7},
				{Name: "SQL", Weight: 0},
			},
		},
		{
			name: "all-zero weights get equal 1/N distribution",
			in: []domain.SkillWeight{
				{Name: "Go"},
				{Name: "SQL"},
				{Name: "AWS"},
				{Name: "K8s"},
			},
			want: []domain.SkillWeight{
				{Name: "Go", Weight: 0.25},
				{Name: "SQL", Weight: 0.25},
				{Name: "AWS", Weight: 0.25},
				{Name: "K8s", Weight: 0.25},
			},
		},
		{
			name: "single zero-weight skill becomes 1.0",
			in:   []domain.SkillWeight{{Name: "Go"}},
			want: []domain.SkillWeight{{Name: "Go", Weight: 1.0}},
		},
		{
			name: "preserves MustHave / NiceToHave flags during redistribution",
			in: []domain.SkillWeight{
				{Name: "Go", MustHave: true},
				{Name: "SQL", NiceToHave: true},
			},
			want: []domain.SkillWeight{
				{Name: "Go", Weight: 0.5, MustHave: true},
				{Name: "SQL", Weight: 0.5, NiceToHave: true},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeSkills(tt.in)
			assert.DeepEqual(t, got, tt.want)
		})
	}
}
