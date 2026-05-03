// Package domain holds transport-independent types for the admin service.
// Aggregates here mirror the wire shapes from admin_model.proto but stay
// free of pb imports so usecase code never reaches into transport.
package domain

import "time"

// SystemStats captures the dashboard-headline counters: how many of each
// entity exist platform-wide. All fields are uint64 because the underlying
// COUNT(*) returns BIGINT and we never decrement.
type SystemStats struct {
	UsersTotal       uint64
	AdminsTotal      uint64
	VacanciesTotal   uint64
	CandidatesTotal  uint64
	AnalysesTotal    uint64
	AnalysesDone     uint64
	AnalysesFailed   uint64
}

// AdminUserView is one row of the user-management table the dashboard
// renders. VacanciesOwned and CandidatesUploaded require cross-table
// joins, hence the dedicated type instead of reusing auth.User.
type AdminUserView struct {
	ID                 uint64
	Email              string
	Role               string
	CreatedAt          time.Time
	VacanciesOwned     uint64
	CandidatesUploaded uint64
}

// UpdateRoleInput is the use-case input for promote/demote. We accept
// the raw role string so the same path serves both directions; the
// usecase validates against domain.RoleAdmin / RoleUser before calling
// the auth client.
type UpdateRoleInput struct {
	CallerUserID uint64 // bookkeeping for future audit log
	IsAdmin      bool
	TargetUserID uint64
	NewRole      string
}

const (
	RoleAdmin = "admin"
	RoleUser  = "user"
)
