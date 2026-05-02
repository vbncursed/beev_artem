import type { VacancyGateway } from '@/application/vacancy/ports'
import {
  VACANCY_STATUS_BY_CODE,
  type CreateVacancyInput,
  type ListVacanciesParams,
  type SkillWeight,
  type UpdateVacancyInput,
  type Vacancy,
  type VacancyPage,
} from '@/domain/vacancy/types'
import type { HttpClient } from '@/infrastructure/http/client'

type SkillWeightDto = {
  name: string
  weight: number
  mustHave?: boolean
  niceToHave?: boolean
}

type VacancyDto = {
  id: string
  ownerUserId?: string
  title: string
  description: string
  role?: string
  skills?: SkillWeightDto[]
  status?: number
  version?: number
  createdAt?: string
  updatedAt?: string
}

type ListResponseDto = {
  vacancies?: VacancyDto[]
  page?: { limit?: number; offset?: number; total?: string }
}

type SingleResponseDto = { vacancy: VacancyDto }

const PATH = {
  base: '/api/v1/vacancies',
  byId: (id: string) => `/api/v1/vacancies/${encodeURIComponent(id)}`,
  archive: (id: string) =>
    `/api/v1/vacancies/${encodeURIComponent(id)}/archive`,
} as const

export class VacancyHttpGateway implements VacancyGateway {
  private readonly http: HttpClient

  constructor(http: HttpClient) {
    this.http = http
  }

  async list(params: ListVacanciesParams): Promise<VacancyPage> {
    const search = new URLSearchParams()
    if (params.query) search.set('query', params.query)
    if (params.limit !== undefined)
      search.set('page.limit', String(params.limit))
    if (params.offset !== undefined)
      search.set('page.offset', String(params.offset))

    const qs = search.toString()
    const url = qs ? `${PATH.base}?${qs}` : PATH.base
    const dto = await this.http.get<ListResponseDto>(url)

    return {
      vacancies: (dto.vacancies ?? []).map(toVacancy),
      total: parseTotal(dto.page?.total),
      limit: dto.page?.limit ?? params.limit ?? 20,
      offset: dto.page?.offset ?? params.offset ?? 0,
    }
  }

  async get(id: string): Promise<Vacancy> {
    const dto = await this.http.get<SingleResponseDto>(PATH.byId(id))
    return toVacancy(dto.vacancy)
  }

  async create(input: CreateVacancyInput): Promise<Vacancy> {
    const dto = await this.http.post<SingleResponseDto>(PATH.base, {
      title: input.title,
      description: input.description,
      role: input.role,
      skills: input.skills.map(toSkillDto),
    })
    return toVacancy(dto.vacancy)
  }

  async update(input: UpdateVacancyInput): Promise<Vacancy> {
    const dto = await this.http.patch<SingleResponseDto>(
      PATH.byId(input.vacancyId),
      {
        vacancyId: input.vacancyId,
        title: input.title,
        description: input.description,
        role: input.role,
        skills: input.skills?.map(toSkillDto),
      },
    )
    return toVacancy(dto.vacancy)
  }

  async archive(id: string): Promise<void> {
    await this.http.post(PATH.archive(id), { vacancyId: id })
  }
}

function toVacancy(dto: VacancyDto): Vacancy {
  return {
    id: dto.id,
    ownerUserId: dto.ownerUserId ?? '',
    title: dto.title,
    description: dto.description,
    role: dto.role ?? 'default',
    skills: (dto.skills ?? []).map(toSkill),
    status: VACANCY_STATUS_BY_CODE[dto.status ?? 0] ?? 'unknown',
    version: dto.version ?? 0,
    createdAt: dto.createdAt ?? '',
    updatedAt: dto.updatedAt ?? '',
  }
}

function toSkill(dto: SkillWeightDto): SkillWeight {
  return {
    name: dto.name,
    weight: typeof dto.weight === 'number' ? dto.weight : 0,
    mustHave: dto.mustHave,
    niceToHave: dto.niceToHave,
  }
}

function toSkillDto(skill: SkillWeight): SkillWeightDto {
  return {
    name: skill.name,
    weight: skill.weight,
    mustHave: skill.mustHave,
    niceToHave: skill.niceToHave,
  }
}

function parseTotal(raw: string | undefined): number {
  if (!raw) return 0
  const n = Number(raw)
  return Number.isFinite(n) ? n : 0
}
