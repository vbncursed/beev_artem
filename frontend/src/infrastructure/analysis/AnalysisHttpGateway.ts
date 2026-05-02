import type {
  AnalysisGateway,
  StartAnalysisInput,
} from '@/application/analysis/ports'
import {
  ANALYSIS_STATUS_BY_CODE,
  type AIDecision,
  type Analysis,
  type AgentResult,
  type CandidateProfile,
  type CandidateWithAnalysis,
  type ListCandidatesPage,
  type ListCandidatesParams,
  type ScoreBreakdown,
} from '@/domain/analysis/types'
import type { HttpClient } from '@/infrastructure/http/client'

type AgentResultDto = {
  agentName?: string
  summary?: string
  structuredJson?: string
  confidence?: number
}

type AIDecisionDto = {
  hrRecommendation?: string
  confidence?: number
  hrRationale?: string
  candidateFeedback?: string
  softSkillsNotes?: string
  agentResults?: AgentResultDto[]
}

type CandidateProfileDto = {
  skills?: string[]
  yearsExperience?: number
  positions?: string[]
  technologies?: string[]
  education?: string[]
  summary?: string
}

type ScoreBreakdownDto = {
  matchedSkills?: string[]
  missingSkills?: string[]
  extraSkills?: string[]
  baseScore?: number
  mustHavePenalty?: number
  niceToHaveBonus?: number
  explanation?: string
}

type AnalysisDto = {
  id: string
  vacancyId?: string
  candidateId?: string
  resumeId?: string
  vacancyVersion?: number
  status?: number
  matchScore?: number
  profile?: CandidateProfileDto
  breakdown?: ScoreBreakdownDto
  ai?: AIDecisionDto
  errorMessage?: string
}

type CandidateWithAnalysisDto = {
  candidateId?: string
  fullName?: string
  email?: string
  phone?: string
  matchScore?: number
  analysisId?: string
  analysisStatus?: number
  createdAt?: string
}

type StartAnalysisResponse = { analysisId: string; status?: number }
type AnalysisResponse = { analysis: AnalysisDto }
type ListCandidatesResponse = {
  candidates?: CandidateWithAnalysisDto[]
  page?: { limit?: number; offset?: number; total?: string }
}

export class AnalysisHttpGateway implements AnalysisGateway {
  private readonly http: HttpClient

  constructor(http: HttpClient) {
    this.http = http
  }

  async start(
    input: StartAnalysisInput,
  ): Promise<{ analysisId: string }> {
    const dto = await this.http.post<StartAnalysisResponse>(
      `/api/v1/resumes/${encodeURIComponent(input.resumeId)}/analyze`,
      {
        resumeId: input.resumeId,
        vacancyId: input.vacancyId,
        useLlm: input.useLlm ?? true,
      },
    )
    return { analysisId: dto.analysisId }
  }

  async get(analysisId: string): Promise<Analysis> {
    const dto = await this.http.get<AnalysisResponse>(
      `/api/v1/analyses/${encodeURIComponent(analysisId)}`,
    )
    return toAnalysis(dto.analysis)
  }

  async listCandidates(
    params: ListCandidatesParams,
  ): Promise<ListCandidatesPage> {
    const search = new URLSearchParams()
    if (params.limit !== undefined)
      search.set('page.limit', String(params.limit))
    if (params.offset !== undefined)
      search.set('page.offset', String(params.offset))
    if (params.minScore !== undefined)
      search.set('minScore', String(params.minScore))
    if (params.requiredSkill) search.set('requiredSkill', params.requiredSkill)
    if (params.scoreOrder) search.set('scoreOrder', params.scoreOrder)

    const qs = search.toString()
    const path = `/api/v1/vacancies/${encodeURIComponent(params.vacancyId)}/candidates${qs ? `?${qs}` : ''}`
    const dto = await this.http.get<ListCandidatesResponse>(path)

    return {
      candidates: (dto.candidates ?? []).map(toCandidateWithAnalysis),
      total: parseInt(dto.page?.total ?? '0', 10) || 0,
      limit: dto.page?.limit ?? params.limit ?? 50,
      offset: dto.page?.offset ?? params.offset ?? 0,
    }
  }
}

function toAnalysis(dto: AnalysisDto): Analysis {
  return {
    id: dto.id,
    vacancyId: dto.vacancyId ?? '',
    candidateId: dto.candidateId ?? '',
    resumeId: dto.resumeId ?? '',
    vacancyVersion: dto.vacancyVersion ?? 0,
    status: ANALYSIS_STATUS_BY_CODE[dto.status ?? 0] ?? 'unknown',
    matchScore: dto.matchScore ?? 0,
    profile: dto.profile ? toProfile(dto.profile) : undefined,
    breakdown: dto.breakdown ? toBreakdown(dto.breakdown) : undefined,
    ai: dto.ai ? toAi(dto.ai) : undefined,
    errorMessage: dto.errorMessage,
  }
}

function toProfile(dto: CandidateProfileDto): CandidateProfile {
  return {
    skills: dto.skills ?? [],
    yearsExperience: dto.yearsExperience ?? 0,
    positions: dto.positions ?? [],
    technologies: dto.technologies ?? [],
    education: dto.education ?? [],
    summary: dto.summary ?? '',
  }
}

function toBreakdown(dto: ScoreBreakdownDto): ScoreBreakdown {
  return {
    matchedSkills: dto.matchedSkills ?? [],
    missingSkills: dto.missingSkills ?? [],
    extraSkills: dto.extraSkills ?? [],
    baseScore: dto.baseScore ?? 0,
    mustHavePenalty: dto.mustHavePenalty ?? 0,
    niceToHaveBonus: dto.niceToHaveBonus ?? 0,
    explanation: dto.explanation ?? '',
  }
}

function toAi(dto: AIDecisionDto): AIDecision {
  return {
    hrRecommendation: dto.hrRecommendation ?? '',
    confidence: dto.confidence ?? 0,
    hrRationale: dto.hrRationale ?? '',
    candidateFeedback: dto.candidateFeedback ?? '',
    softSkillsNotes: dto.softSkillsNotes ?? '',
    agentResults: (dto.agentResults ?? []).map(toAgentResult),
  }
}

function toAgentResult(dto: AgentResultDto): AgentResult {
  return {
    agentName: dto.agentName ?? '',
    summary: dto.summary ?? '',
    structuredJson: dto.structuredJson ?? '',
    confidence: dto.confidence ?? 0,
  }
}

function toCandidateWithAnalysis(
  dto: CandidateWithAnalysisDto,
): CandidateWithAnalysis {
  return {
    candidateId: dto.candidateId ?? '',
    fullName: dto.fullName ?? '',
    email: dto.email ?? '',
    phone: dto.phone ?? '',
    matchScore: dto.matchScore ?? 0,
    analysisId: dto.analysisId ?? '',
    analysisStatus:
      ANALYSIS_STATUS_BY_CODE[dto.analysisStatus ?? 0] ?? 'unknown',
    createdAt: dto.createdAt ?? '',
  }
}
