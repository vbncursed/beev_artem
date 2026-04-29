package domain

import "time"

const (
	StatusActive   = "active"
	StatusArchived = "archived"
)

type SkillWeight struct {
	Name       string
	Weight     float32
	MustHave   bool
	NiceToHave bool
}

type Vacancy struct {
	ID          string
	OwnerUserID uint64
	Title       string
	Description string
	Skills      []SkillWeight
	Status      string
	Version     uint32
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type CreateVacancyInput struct {
	OwnerUserID uint64
	Title       string
	Description string
	Skills      []SkillWeight
}

type GetVacancyInput struct {
	VacancyID   string
	OwnerUserID uint64
	IsAdmin     bool
}

type ListVacanciesInput struct {
	OwnerUserID uint64
	Limit       uint32
	Offset      uint32
	Query       string
	IsAdmin     bool
}

type UpdateVacancyInput struct {
	VacancyID   string
	OwnerUserID uint64
	IsAdmin     bool
	Title       string
	Description string
	Skills      []SkillWeight
}

type ArchiveVacancyInput struct {
	VacancyID   string
	OwnerUserID uint64
	IsAdmin     bool
}

type ListVacanciesResult struct {
	Vacancies []Vacancy
	Total     uint64
}
