package usecase

import (
	"testing"

	"gotest.tools/v3/assert"
)

// TestDetectRole pins the title/description -> role mapping. The detector is
// the source of truth for what prompt multiagent will load, so each role's
// happy path plus the fallback gets a row here. Not a suite — pure function,
// no fixtures needed.
func TestDetectRole(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		title       string
		description string
		want        string
	}{
		{
			name:  "russian programmer title",
			title: "Senior Go разработчик",
			want:  "programmer",
		},
		{
			name:        "english backend engineer",
			title:       "Backend Engineer",
			description: "We build distributed systems in Go.",
			want:        "programmer",
		},
		{
			name:  "frontend",
			title: "Frontend Developer (React)",
			want:  "programmer",
		},
		{
			name:  "russian manager",
			title: "Менеджер по продажам",
			want:  "manager",
		},
		{
			name:  "english team lead",
			title: "Engineering Team Lead",
			want:  "manager",
		},
		{
			name:        "tim-lid wins over engineer keyword",
			title:       "Тимлид backend команды",
			description: "Руководит инженерами.",
			want:        "manager",
		},
		{
			name:  "russian accountant",
			title: "Главный бухгалтер",
			want:  "accountant",
		},
		{
			name:  "english accountant",
			title: "Senior Accountant",
			want:  "accountant",
		},
		{
			name:  "data analyst beats programmer",
			title: "Data Analyst",
			want:  "analyst",
		},
		{
			name:  "russian analyst",
			title: "Бизнес-аналитик",
			want:  "analyst",
		},
		{
			name:  "doctor",
			title: "Терапевт",
			want:  "doctor",
		},
		{
			name:  "electrician",
			title: "Электрик 4 разряда",
			want:  "electrician",
		},
		{
			name:  "fallback when nothing matches",
			title: "Уникальная позиция",
			want:  "default",
		},
		{
			name:  "empty input",
			title: "",
			want:  "default",
		},
		{
			name:        "case insensitive",
			title:       "PROGRAMMING ENGINEER",
			description: "WE WRITE GOLANG.",
			want:        "programmer",
		},
		{
			name:        "description-only signal",
			title:       "Senior position",
			description: "Looking for an experienced Python developer.",
			want:        "programmer",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DetectRole(tt.title, tt.description)
			assert.Equal(t, got, tt.want)
		})
	}
}
