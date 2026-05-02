import type {
  Analysis,
  ListCandidatesPage,
  ListCandidatesParams,
} from '@/domain/analysis/types'

export type StartAnalysisInput = {
  vacancyId: string
  resumeId: string
  useLlm?: boolean
}

export interface AnalysisGateway {
  start(input: StartAnalysisInput): Promise<{ analysisId: string }>
  get(analysisId: string): Promise<Analysis>
  listCandidates(params: ListCandidatesParams): Promise<ListCandidatesPage>
}
