export type AgentResultDto = {
  agentName?: string
  summary?: string
  structuredJson?: string
  confidence?: number
}

export type AIDecisionDto = {
  hrRecommendation?: string
  confidence?: number
  hrRationale?: string
  candidateFeedback?: string
  softSkillsNotes?: string
  agentResults?: AgentResultDto[]
}

export type CandidateProfileDto = {
  skills?: string[]
  yearsExperience?: number
  positions?: string[]
  technologies?: string[]
  education?: string[]
  summary?: string
}

export type ScoreBreakdownDto = {
  matchedSkills?: string[]
  missingSkills?: string[]
  extraSkills?: string[]
  baseScore?: number
  mustHavePenalty?: number
  niceToHaveBonus?: number
  explanation?: string
}

export type AnalysisDto = {
  id: string
  vacancyId?: string
  candidateId?: string
  resumeId?: string
  vacancyVersion?: number
  status?: number | string
  matchScore?: number
  profile?: CandidateProfileDto
  breakdown?: ScoreBreakdownDto
  ai?: AIDecisionDto
  errorMessage?: string
}

export type CandidateWithAnalysisDto = {
  candidateId?: string
  fullName?: string
  email?: string
  phone?: string
  matchScore?: number
  analysisId?: string
  analysisStatus?: number | string
  createdAt?: string
}

export type StartAnalysisResponse = { analysisId: string; status?: number }
export type AnalysisResponse = { analysis: AnalysisDto }
export type ListCandidatesResponse = {
  candidates?: CandidateWithAnalysisDto[]
  page?: { limit?: number; offset?: number; total?: string }
}
