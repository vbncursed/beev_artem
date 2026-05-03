import {
  parseAnalysisStatus,
  type AIDecision,
  type Analysis,
  type AgentResult,
  type CandidateProfile,
  type CandidateWithAnalysis,
  type ScoreBreakdown,
} from '@/domain/analysis/types'
import type {
  AgentResultDto,
  AIDecisionDto,
  AnalysisDto,
  CandidateProfileDto,
  CandidateWithAnalysisDto,
  ScoreBreakdownDto,
} from './dto'

export function toAnalysis(dto: AnalysisDto): Analysis {
  return {
    id: dto.id,
    vacancyId: dto.vacancyId ?? '',
    candidateId: dto.candidateId ?? '',
    resumeId: dto.resumeId ?? '',
    vacancyVersion: dto.vacancyVersion ?? 0,
    status: parseAnalysisStatus(dto.status),
    matchScore: dto.matchScore ?? 0,
    profile: dto.profile ? toProfile(dto.profile) : undefined,
    breakdown: dto.breakdown ? toBreakdown(dto.breakdown) : undefined,
    ai: dto.ai ? toAi(dto.ai) : undefined,
    errorMessage: dto.errorMessage,
  }
}

export function toProfile(dto: CandidateProfileDto): CandidateProfile {
  return {
    skills: dto.skills ?? [],
    yearsExperience: dto.yearsExperience ?? 0,
    positions: dto.positions ?? [],
    technologies: dto.technologies ?? [],
    education: dto.education ?? [],
    summary: dto.summary ?? '',
  }
}

export function toBreakdown(dto: ScoreBreakdownDto): ScoreBreakdown {
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

export function toAi(dto: AIDecisionDto): AIDecision {
  return {
    hrRecommendation: dto.hrRecommendation ?? '',
    confidence: dto.confidence ?? 0,
    hrRationale: dto.hrRationale ?? '',
    candidateFeedback: dto.candidateFeedback ?? '',
    softSkillsNotes: dto.softSkillsNotes ?? '',
    agentResults: (dto.agentResults ?? []).map(toAgentResult),
  }
}

export function toAgentResult(dto: AgentResultDto): AgentResult {
  return {
    agentName: dto.agentName ?? '',
    summary: dto.summary ?? '',
    structuredJson: dto.structuredJson ?? '',
    confidence: dto.confidence ?? 0,
  }
}

export function toCandidateWithAnalysis(
  dto: CandidateWithAnalysisDto,
): CandidateWithAnalysis {
  return {
    candidateId: dto.candidateId ?? '',
    fullName: dto.fullName ?? '',
    email: dto.email ?? '',
    phone: dto.phone ?? '',
    matchScore: dto.matchScore ?? 0,
    analysisId: dto.analysisId ?? '',
    analysisStatus: parseAnalysisStatus(dto.analysisStatus),
    createdAt: dto.createdAt ?? '',
  }
}
