import type {
  BatchIngestResult,
  Candidate,
  CandidateWithResume,
  Resume,
} from '@/domain/resume/types'

export type CreateCandidateFromResumeInput = {
  vacancyId: string
  /** Raw file bytes — gateway will base64-encode for the backend. */
  file: File
}

export type IngestResumeBatchInput = {
  vacancyId: string
  /**
   * Files to ingest in one request. Order is preserved — each item's
   * external_id is its index in this array, so the caller can map results
   * back to the original `File` for per-file UI feedback.
   */
  files: File[]
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
  ingestResumeBatch(input: IngestResumeBatchInput): Promise<BatchIngestResult>
  getCandidate(candidateId: string): Promise<Candidate>
  getResume(resumeId: string): Promise<Resume>
  downloadResume(resumeId: string): Promise<ResumeFile>
  deleteCandidate(candidateId: string): Promise<void>
}
