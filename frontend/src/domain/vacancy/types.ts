export type VacancyId = string

/**
 * Backend exposes status as a numeric enum. The vacancy proto uses:
 *   0 — DRAFT, 1 — OPEN, 2 — ARCHIVED  (cf. beev/vacancy/api/models)
 * If the schema evolves we keep an `unknown` fallback so the UI never crashes.
 */
export type VacancyStatus = 'draft' | 'open' | 'archived' | 'unknown'

export const VACANCY_STATUS_BY_CODE: Record<number, VacancyStatus> = {
  0: 'draft',
  1: 'open',
  2: 'archived',
}

/**
 * grpc-gateway emits proto enums as their string name by default
 * (e.g. `"VACANCY_STATUS_OPEN"`). Accept both shapes so the frontend
 * survives any future marshaler tweak on the gateway.
 */
const VACANCY_STATUS_BY_NAME: Record<string, VacancyStatus> = {
  VACANCY_STATUS_DRAFT: 'draft',
  VACANCY_STATUS_OPEN: 'open',
  VACANCY_STATUS_ARCHIVED: 'archived',
  VACANCY_STATUS_UNSPECIFIED: 'unknown',
}

export function parseVacancyStatus(raw: unknown): VacancyStatus {
  if (typeof raw === 'number') {
    return VACANCY_STATUS_BY_CODE[raw] ?? 'unknown'
  }
  if (typeof raw === 'string') {
    return VACANCY_STATUS_BY_NAME[raw] ?? 'unknown'
  }
  return 'unknown'
}

/**
 * Free-form role string. multiagent looks up `assets/prompts/<role>.txt`
 * and falls back to `default`. Known roles drive prompt selection but the
 * backend accepts any string.
 */
export type VacancyRole = string

/**
 * Source of truth: vacancy/internal/usecase/role_detector.go +
 * multiagent/internal/infrastructure/prompts/templates/<role>.txt.
 * Adding a new role on the backend (drop a new prompt template + extend
 * roleKeywords) → mirror it here.
 *
 * Order matches the detector priority: more specific first, generic last.
 */
export const KNOWN_ROLES: ReadonlyArray<{ value: string; label: string }> = [
  { value: 'accountant', label: 'Accountant' },
  { value: 'doctor', label: 'Doctor' },
  { value: 'electrician', label: 'Electrician' },
  { value: 'analyst', label: 'Analyst' },
  { value: 'manager', label: 'Manager' },
  { value: 'programmer', label: 'Programmer' },
  { value: 'default', label: 'Other' },
]

export const KNOWN_ROLE_VALUES: ReadonlySet<string> = new Set(
  KNOWN_ROLES.map((r) => r.value),
)

export type SkillWeight = {
  name: string
  weight: number
  mustHave?: boolean
  niceToHave?: boolean
}

export type Vacancy = {
  id: VacancyId
  ownerUserId: string
  title: string
  description: string
  role: VacancyRole
  skills: SkillWeight[]
  status: VacancyStatus
  version: number
  createdAt: string
  updatedAt: string
}

export type VacancyPage = {
  vacancies: Vacancy[]
  total: number
  limit: number
  offset: number
}

export type ListVacanciesParams = {
  query?: string
  limit?: number
  offset?: number
}

export type CreateVacancyInput = {
  title: string
  description: string
  role?: string
  skills: SkillWeight[]
}

export type UpdateVacancyInput = {
  vacancyId: VacancyId
  title?: string
  description?: string
  role?: string
  skills?: SkillWeight[]
}
