/**
 * AnalysisStatus enum (analysis/api/models/*.proto):
 *   0 UNSPECIFIED · 1 QUEUED · 2 RUNNING · 3 DONE · 4 FAILED
 */
export type AnalysisStatus =
  | 'queued'
  | 'running'
  | 'done'
  | 'failed'
  | 'unknown'

export const ANALYSIS_STATUS_BY_CODE: Record<number, AnalysisStatus> = {
  0: 'unknown',
  1: 'queued',
  2: 'running',
  3: 'done',
  4: 'failed',
}

/**
 * grpc-gateway by default serializes proto enums as their **string name**
 * (e.g. `"ANALYSIS_STATUS_DONE"`) — not as the integer code. We accept
 * both shapes here so the frontend keeps working whether the gateway
 * marshaler is reconfigured later or not.
 */
const ANALYSIS_STATUS_BY_NAME: Record<string, AnalysisStatus> = {
  ANALYSIS_STATUS_UNSPECIFIED: 'unknown',
  ANALYSIS_STATUS_QUEUED: 'queued',
  ANALYSIS_STATUS_RUNNING: 'running',
  ANALYSIS_STATUS_DONE: 'done',
  ANALYSIS_STATUS_FAILED: 'failed',
}

export function parseAnalysisStatus(raw: unknown): AnalysisStatus {
  if (typeof raw === 'number') {
    return ANALYSIS_STATUS_BY_CODE[raw] ?? 'unknown'
  }
  if (typeof raw === 'string') {
    return ANALYSIS_STATUS_BY_NAME[raw] ?? 'unknown'
  }
  return 'unknown'
}

export type CandidateProfile = {
  skills: string[]
  yearsExperience: number
  positions: string[]
  technologies: string[]
  education: string[]
  summary: string
}

export type ScoreBreakdown = {
  matchedSkills: string[]
  missingSkills: string[]
  extraSkills: string[]
  baseScore: number
  mustHavePenalty: number
  niceToHaveBonus: number
  explanation: string
}

export type AgentResult = {
  agentName: string
  summary: string
  structuredJson: string
  confidence: number
}

export type AIDecision = {
  hrRecommendation: string
  confidence: number
  hrRationale: string
  candidateFeedback: string
  softSkillsNotes: string
  agentResults: AgentResult[]
}

export type Analysis = {
  id: string
  vacancyId: string
  candidateId: string
  resumeId: string
  vacancyVersion: number
  status: AnalysisStatus
  matchScore: number
  profile?: CandidateProfile
  breakdown?: ScoreBreakdown
  ai?: AIDecision
  errorMessage?: string
}

export type CandidateWithAnalysis = {
  candidateId: string
  fullName: string
  email: string
  phone: string
  matchScore: number
  analysisId: string
  analysisStatus: AnalysisStatus
  createdAt: string
}

/**
 * common.v1.SortOrder — backend accepts the proto enum name as the
 * query value (grpc-gateway), not the lowercase short form.
 */
export type SortOrder = 'SORT_ORDER_ASC' | 'SORT_ORDER_DESC'

export type ListCandidatesParams = {
  vacancyId: string
  limit?: number
  offset?: number
  minScore?: number
  requiredSkill?: string
  scoreOrder?: SortOrder
}

export type ListCandidatesPage = {
  candidates: CandidateWithAnalysis[]
  total: number
  limit: number
  offset: number
}
