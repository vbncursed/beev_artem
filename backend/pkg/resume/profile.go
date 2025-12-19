package resume

import (
	"time"

	"github.com/google/uuid"
)

// Profile — структурированное представление резюме (MVP).
type Profile struct {
	Summary    string           `json:"summary"`
	Skills     []string         `json:"skills"`
	Experience []ExperienceItem `json:"experience"`
	Education  []EducationItem  `json:"education"`
}

type ExperienceItem struct {
	Company     string `json:"company"`
	Role        string `json:"role"`
	Start       string `json:"start"` // YYYY-MM or free text
	End         string `json:"end"`   // YYYY-MM or "present"
	Description string `json:"description"`
}

type EducationItem struct {
	Institution string `json:"institution"`
	Degree      string `json:"degree"`
	Start       string `json:"start"`
	End         string `json:"end"`
}

type ProfileStatus string

const (
	ProfileStatusPending ProfileStatus = "pending"
	ProfileStatusOK      ProfileStatus = "ok"
	ProfileStatusFailed  ProfileStatus = "failed"
)

// ProfileRecord — то, что мы храним в БД по резюме.
type ProfileRecord struct {
	ResumeID  uuid.UUID     `json:"resumeId"`
	Status   ProfileStatus `json:"status"`
	Model    string        `json:"model"`
	Error    string        `json:"error"`
	Profile  Profile       `json:"profile"`
	UpdatedAt time.Time    `json:"updatedAt"`
}


