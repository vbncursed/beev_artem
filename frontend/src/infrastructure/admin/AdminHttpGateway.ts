import type { AdminGateway } from '@/application/admin/ports'
import type { AdminUserView, SystemStats } from '@/domain/admin/types'
import type { HttpClient } from '@/infrastructure/http/client'

type StatsDto = {
  usersTotal?: string | number
  adminsTotal?: string | number
  vacanciesTotal?: string | number
  candidatesTotal?: string | number
  analysesTotal?: string | number
  analysesDone?: string | number
  analysesFailed?: string | number
}

type OverviewResponse = { stats?: StatsDto }

type UserDto = {
  id?: string | number
  email?: string
  role?: string
  createdAt?: string
  vacanciesOwned?: string | number
  candidatesUploaded?: string | number
}

type ListUsersResponse = { users?: UserDto[] }

export class AdminHttpGateway implements AdminGateway {
  private readonly http: HttpClient

  constructor(http: HttpClient) {
    this.http = http
  }

  async getOverview(): Promise<SystemStats> {
    const dto = await this.http.get<OverviewResponse>('/api/v1/admin/overview')
    const s = dto.stats ?? {}
    return {
      usersTotal: toNum(s.usersTotal),
      adminsTotal: toNum(s.adminsTotal),
      vacanciesTotal: toNum(s.vacanciesTotal),
      candidatesTotal: toNum(s.candidatesTotal),
      analysesTotal: toNum(s.analysesTotal),
      analysesDone: toNum(s.analysesDone),
      analysesFailed: toNum(s.analysesFailed),
    }
  }

  async listUsers(): Promise<AdminUserView[]> {
    const dto = await this.http.get<ListUsersResponse>('/api/v1/admin/users')
    return (dto.users ?? []).map((u) => ({
      id: toNum(u.id),
      email: u.email ?? '',
      role: u.role ?? 'user',
      createdAt: u.createdAt ?? '',
      vacanciesOwned: toNum(u.vacanciesOwned),
      candidatesUploaded: toNum(u.candidatesUploaded),
    }))
  }

  async promoteUser(userId: number): Promise<void> {
    await this.http.post(
      `/api/v1/admin/users/${userId}/promote`,
      { userId },
    )
  }

  async demoteUser(userId: number): Promise<void> {
    await this.http.post(
      `/api/v1/admin/users/${userId}/demote`,
      { userId },
    )
  }
}

function toNum(v: unknown): number {
  if (typeof v === 'number') return v
  if (typeof v === 'string') {
    const n = Number(v)
    return Number.isFinite(n) ? n : 0
  }
  return 0
}
