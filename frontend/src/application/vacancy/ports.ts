import type {
  CreateVacancyInput,
  ListVacanciesParams,
  UpdateVacancyInput,
  Vacancy,
  VacancyPage,
} from '@/domain/vacancy/types'

export interface VacancyGateway {
  list(params: ListVacanciesParams): Promise<VacancyPage>
  get(id: string): Promise<Vacancy>
  create(input: CreateVacancyInput): Promise<Vacancy>
  update(input: UpdateVacancyInput): Promise<Vacancy>
  archive(id: string): Promise<void>
}
