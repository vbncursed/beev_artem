export type ResumeId = string
export type CandidateId = string

export type Candidate = {
  id: CandidateId
  vacancyId: string
  fullName: string
  email: string
  phone: string
  source: string
  comment: string
  createdAt: string
}

export type Resume = {
  id: ResumeId
  candidateId: CandidateId
  fileName: string
  fileType: string
  fileSizeBytes: number
  createdAt: string
}

export type CandidateWithResume = {
  candidate: Candidate
  resume: Resume
}

/** Files we accept in the upload zone — extension hint, not validation. */
export const ACCEPTED_RESUME_TYPES = '.pdf,.doc,.docx,.txt'
export const MAX_RESUME_BYTES = 10 * 1024 * 1024 // 10 MB
