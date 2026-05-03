import { useCallback, useEffect, useState } from 'react'
import { useGateways } from '@/app/providers/GatewaysProvider'
import { useI18n } from '@/app/providers/I18nProvider'
import { ApiError } from '@/infrastructure/http/errors'
import type { AdminUserView, SystemStats } from '@/domain/admin/types'

type DashboardState = {
  stats: SystemStats | null
  users: AdminUserView[]
  statsErr: string | null
  usersErr: string | null
  loading: boolean
  refresh: () => Promise<void>
}

/**
 * Fetches the two top-level admin endpoints in parallel via
 * Promise.allSettled so a 500 on one of them doesn't hide the other.
 * Per-section error strings are surfaced separately.
 */
export function useAdminDashboard(): DashboardState {
  const { admin } = useGateways()
  const { t } = useI18n()
  const [stats, setStats] = useState<SystemStats | null>(null)
  const [users, setUsers] = useState<AdminUserView[]>([])
  const [statsErr, setStatsErr] = useState<string | null>(null)
  const [usersErr, setUsersErr] = useState<string | null>(null)
  const [loading, setLoading] = useState(true)

  const refresh = useCallback(async () => {
    setLoading(true)
    setStatsErr(null)
    setUsersErr(null)
    const [overviewRes, usersRes] = await Promise.allSettled([
      admin.getOverview(),
      admin.listUsers(),
    ])
    if (overviewRes.status === 'fulfilled') setStats(overviewRes.value)
    else setStatsErr(messageOf(overviewRes.reason, t('admin.errors.stats')))

    if (usersRes.status === 'fulfilled') setUsers(usersRes.value)
    else setUsersErr(messageOf(usersRes.reason, t('admin.errors.users')))

    setLoading(false)
  }, [admin, t])

  useEffect(() => {
    void refresh()
  }, [refresh])

  return { stats, users, statsErr, usersErr, loading, refresh }
}

function messageOf(cause: unknown, fallback: string): string {
  if (cause instanceof ApiError) return cause.message || fallback
  return fallback
}
