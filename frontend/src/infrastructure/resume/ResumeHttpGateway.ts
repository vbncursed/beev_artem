import type {
  CreateCandidateFromResumeInput,
  ResumeFile,
  ResumeGateway,
} from '@/application/resume/ports'
import type {
  Candidate,
  CandidateWithResume,
  Resume,
} from '@/domain/resume/types'
import type { HttpClient } from '@/infrastructure/http/client'

type CandidateDto = {
  id: string
  vacancyId?: string
  fullName?: string
  email?: string
  phone?: string
  source?: string
  comment?: string
  createdAt?: string
}

type ResumeDto = {
  id: string
  candidateId?: string
  fileName?: string
  fileType?: string
  fileSizeBytes?: string
  storagePath?: string
  extractedText?: string
  createdAt?: string
}

type CandidateResumeResponse = {
  candidate: CandidateDto
  resume: ResumeDto
}

type CandidateResponse = { candidate: CandidateDto }
type ResumeResponse = { resume: ResumeDto }

export class ResumeHttpGateway implements ResumeGateway {
  private readonly http: HttpClient

  constructor(http: HttpClient) {
    this.http = http
  }

  async createCandidateFromResume(
    input: CreateCandidateFromResumeInput,
  ): Promise<CandidateWithResume> {
    const fileData = await fileToBase64(input.file)
    const dto = await this.http.post<CandidateResumeResponse>(
      `/api/v1/vacancies/${encodeURIComponent(input.vacancyId)}/candidates/from-resume`,
      { vacancyId: input.vacancyId, fileData },
    )
    return {
      candidate: toCandidate(dto.candidate),
      resume: toResume(dto.resume),
    }
  }

  async getCandidate(candidateId: string): Promise<Candidate> {
    const dto = await this.http.get<CandidateResponse>(
      `/api/v1/candidates/${encodeURIComponent(candidateId)}`,
    )
    return toCandidate(dto.candidate)
  }

  async getResume(resumeId: string): Promise<Resume> {
    const dto = await this.http.get<ResumeResponse>(
      `/api/v1/resumes/${encodeURIComponent(resumeId)}`,
    )
    return toResume(dto.resume)
  }

  async downloadResume(resumeId: string): Promise<ResumeFile> {
    const dto = await this.http.get<{
      fileData?: string
      fileName?: string
      fileType?: string
    }>(`/api/v1/resumes/${encodeURIComponent(resumeId)}/download`)
    return {
      fileName: dto.fileName ?? 'resume',
      fileType: dto.fileType ?? '',
      data: base64ToBytes(dto.fileData ?? ''),
    }
  }

  async deleteCandidate(candidateId: string): Promise<void> {
    await this.http.delete(
      `/api/v1/candidates/${encodeURIComponent(candidateId)}`,
    )
  }
}

/**
 * grpc-gateway serializes proto `bytes` as base64 in JSON. We decode in
 * chunks via atob so very large resumes don't blow the call stack.
 */
function base64ToBytes(b64: string): Uint8Array {
  if (!b64) return new Uint8Array(0)
  const binary = atob(b64)
  const bytes = new Uint8Array(binary.length)
  for (let i = 0; i < binary.length; i++) bytes[i] = binary.charCodeAt(i)
  return bytes
}

function toCandidate(dto: CandidateDto): Candidate {
  return {
    id: dto.id,
    vacancyId: dto.vacancyId ?? '',
    fullName: dto.fullName ?? '',
    email: dto.email ?? '',
    phone: dto.phone ?? '',
    source: dto.source ?? '',
    comment: dto.comment ?? '',
    createdAt: dto.createdAt ?? '',
  }
}

function toResume(dto: ResumeDto): Resume {
  return {
    id: dto.id,
    candidateId: dto.candidateId ?? '',
    fileName: dto.fileName ?? '',
    fileType: dto.fileType ?? '',
    fileSizeBytes: parseInt(dto.fileSizeBytes ?? '0', 10) || 0,
    createdAt: dto.createdAt ?? '',
  }
}

/**
 * Backend takes `fileData` as a base64 string inside JSON (proto bytes →
 * base64 in grpc-gateway). We avoid loading the whole binary as a
 * data-URL: read as ArrayBuffer, encode with Uint8Array → btoa over chunks.
 */
async function fileToBase64(file: File): Promise<string> {
  const buffer = await file.arrayBuffer()
  const bytes = new Uint8Array(buffer)
  let binary = ''
  const chunk = 0x8000
  for (let i = 0; i < bytes.length; i += chunk) {
    binary += String.fromCharCode(...bytes.subarray(i, i + chunk))
  }
  return btoa(binary)
}
