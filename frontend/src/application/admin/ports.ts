import type { AdminUserView, SystemStats } from '@/domain/admin/types'

export interface AdminGateway {
  getOverview(): Promise<SystemStats>
  listUsers(): Promise<AdminUserView[]>
  promoteUser(userId: number): Promise<void>
  demoteUser(userId: number): Promise<void>
}
