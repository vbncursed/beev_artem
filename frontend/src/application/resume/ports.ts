import type {
  Candidate,
  CandidateWithResume,
  Resume,
} from '@/domain/resume/types'

export type CreateCandidateFromResumeInput = {
  vacancyId: string
  /** Raw file bytes — gateway will base64-encode for the backend. */
  file: File
}

export type ResumeFile = {
  fileName: string
  fileType: string
  /** Decoded raw bytes ready to wrap in a Blob. */
  data: Uint8Array
}

export interface ResumeGateway {
  createCandidateFromResume(
    input: CreateCandidateFromResumeInput,
  ): Promise<CandidateWithResume>
  getCandidate(candidateId: string): Promise<Candidate>
  getResume(resumeId: string): Promise<Resume>
  downloadResume(resumeId: string): Promise<ResumeFile>
  deleteCandidate(candidateId: string): Promise<void>
}
