import type {
  AnalysisGateway,
  StartAnalysisInput,
} from '@/application/analysis/ports'
import type {
  Analysis,
  ListCandidatesPage,
  ListCandidatesParams,
} from '@/domain/analysis/types'
import type { HttpClient } from '@/infrastructure/http/client'
import type {
  AnalysisResponse,
  ListCandidatesResponse,
  StartAnalysisResponse,
} from './dto'
import { toAnalysis, toCandidateWithAnalysis } from './mappers'

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
