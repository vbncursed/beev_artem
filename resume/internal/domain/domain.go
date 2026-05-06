package domain

import "time"

type Candidate struct {
	ID          string
	VacancyID   string
	OwnerUserID uint64
	FullName    string
	Email       string
	Phone       string
	Source      string
	Comment     string
	CreatedAt   time.Time
}

type Resume struct {
	ID            string
	CandidateID   string
	FileName      string
	FileType      string
	FileSizeBytes uint64
	StoragePath   string
	ExtractedText string
	CreatedAt     time.Time
}

type CreateCandidateInput struct {
	RequestUserID uint64
	VacancyID     string
	FullName      string
	Email         string
	Phone         string
	Source        string
	Comment       string
}

type GetCandidateInput struct {
	RequestUserID uint64
	IsAdmin       bool
	CandidateID   string
}

type UploadResumeInput struct {
	RequestUserID uint64
	IsAdmin       bool
	CandidateID   string
	FileName      string
	FileType      string
	ExtractedText string
	Data          []byte
}

type GetResumeInput struct {
	RequestUserID uint64
	IsAdmin       bool
	ResumeID      string
}

// DownloadResumeInput drives the file-bytes fetch path. Same authz shape
// as GetResume; we keep it as a separate type so the heavy `Data` payload
// is never returned by the lighter GetResume read.
type DownloadResumeInput struct {
	RequestUserID uint64
	IsAdmin       bool
	ResumeID      string
}

// ResumeFile is the binary payload returned by DownloadResume — only the
// bits needed to write a file response (filename + content-type + bytes).
// Distinct from Resume so callers can't accidentally surface the full
// extracted text alongside the blob.
type ResumeFile struct {
	FileName string
	FileType string
	Data     []byte
}

// DeleteCandidateInput drives the destructive remove path. The DB schema
// declares ON DELETE CASCADE on resumes.candidate_id, so a single DELETE
// against `candidates` clears both rows in one statement. Authorization is
// the same shape as GetCandidate.
type DeleteCandidateInput struct {
	RequestUserID uint64
	IsAdmin       bool
	CandidateID   string
}

type CreateCandidateFromResumeInput struct {
	RequestUserID uint64
	VacancyID     string
	FileData      []byte
}

// NewResumeData carries the resume file payload without an owning CandidateID.
// Used by storage's CreateCandidateWithResume where the candidate ID is
// generated inside the same transaction.
type NewResumeData struct {
	FileName      string
	FileType      string
	ExtractedText string
	Data          []byte
}

type CandidateResumeResult struct {
	Candidate *Candidate
	Resume    *Resume
}

type BatchIngestResumeInput struct {
	RequestUserID uint64
	VacancyID     string
	Files         []ResumeIntakeFile
}

type ResumeIntakeFile struct {
	ExternalID string
	FileData   []byte
}

type BatchIngestResumeResult struct {
	Results []BatchIngestResumeItemResult
}

type BatchIngestResumeItemResult struct {
	ExternalID string
	Candidate  *Candidate
	Resume     *Resume
	Error      string
}
