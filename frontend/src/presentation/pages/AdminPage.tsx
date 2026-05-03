import { useEffect, useState } from 'react'
import { useGateways } from '@/app/providers/GatewaysProvider'
import { useI18n } from '@/app/providers/I18nProvider'
import {
  AssetIconCircular,
  BadgePill,
  Button,
  Card,
  PriceCell,
  Spinner,
} from '@/presentation/ui'
import type { AdminUserView, SystemStats } from '@/domain/admin/types'
import { ApiError } from '@/infrastructure/http/errors'

/**
 * Admin dashboard at /admin. Two sections:
 *   1. Stats card grid — top-line counters from GetOverview
 *   2. User table — every HR account with role + activity + promote/demote
 *
 * Page-level loading + error states are handled per-section so a failed
 * stats fetch doesn't hide the user list, and vice versa.
 */
export function AdminPage() {
  const { t } = useI18n()
  const { admin } = useGateways()
  const [stats, setStats] = useState<SystemStats | null>(null)
  const [users, setUsers] = useState<AdminUserView[]>([])
  const [statsErr, setStatsErr] = useState<string | null>(null)
  const [usersErr, setUsersErr] = useState<string | null>(null)
  const [loading, setLoading] = useState(true)

  const refresh = async () => {
    setLoading(true)
    setStatsErr(null)
    setUsersErr(null)
    const [overviewRes, usersRes] = await Promise.allSettled([
      admin.getOverview(),
      admin.listUsers(),
    ])
    if (overviewRes.status === 'fulfilled') {
      setStats(overviewRes.value)
    } else {
      setStatsErr(messageOf(overviewRes.reason, t('admin.errors.stats')))
    }
    if (usersRes.status === 'fulfilled') {
      setUsers(usersRes.value)
    } else {
      setUsersErr(messageOf(usersRes.reason, t('admin.errors.users')))
    }
    setLoading(false)
  }

  useEffect(() => {
    void refresh()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  return (
    <>
      <section className="bg-canvas">
        <div className="mx-auto max-w-[1200px] px-6 pt-[96px] pb-12">
          <div className="flex items-end justify-between gap-4">
            <div>
              <BadgePill tone="inverse">{t('admin.eyebrow')}</BadgePill>
              <h1 className="text-display-md mt-4">{t('admin.title')}</h1>
              <p className="text-body-md text-body mt-3 max-w-[520px]">
                {t('admin.subtitle')}
              </p>
            </div>
            {loading && <Spinner size={20} />}
          </div>
        </div>
      </section>

      <section className="bg-surface-soft">
        <div className="mx-auto max-w-[1200px] px-6 py-12">
          {statsErr ? (
            <ErrorCard message={statsErr} />
          ) : (
            <StatsGrid stats={stats} />
          )}
        </div>
      </section>

      <section className="bg-canvas">
        <div className="mx-auto max-w-[1200px] px-6 py-12 pb-[96px]">
          <div className="flex items-end justify-between pb-6">
            <div>
              <p className="text-caption-strong text-muted uppercase">
                {t('admin.users.eyebrow')}
              </p>
              <p className="text-title-lg mt-1">
                {users.length}{' '}
                <span className="text-body text-body-md font-normal">
                  {t('admin.users.subtitle')}
                </span>
              </p>
            </div>
          </div>

          {usersErr ? (
            <ErrorCard message={usersErr} />
          ) : (
            <UsersTable users={users} onRefresh={refresh} />
          )}
        </div>
      </section>
    </>
  )
}

function StatsGrid({ stats }: { stats: SystemStats | null }) {
  const { t } = useI18n()
  const cards = [
    {
      key: 'users',
      label: t('admin.stats.users'),
      value: stats?.usersTotal ?? 0,
      sub: stats
        ? t('admin.stats.usersSub', { admins: stats.adminsTotal })
        : '',
    },
    {
      key: 'vacancies',
      label: t('admin.stats.vacancies'),
      value: stats?.vacanciesTotal ?? 0,
      sub: '',
    },
    {
      key: 'candidates',
      label: t('admin.stats.candidates'),
      value: stats?.candidatesTotal ?? 0,
      sub: '',
    },
    {
      key: 'analyses',
      label: t('admin.stats.analyses'),
      value: stats?.analysesTotal ?? 0,
      sub: stats
        ? t('admin.stats.analysesSub', {
            done: stats.analysesDone,
            failed: stats.analysesFailed,
          })
        : '',
    },
  ]
  return (
    <ul className="grid grid-cols-1 gap-6 md:grid-cols-2 xl:grid-cols-4">
      {cards.map((c) => (
        <li key={c.key}>
          <Card variant="feature" className="flex h-full flex-col gap-3">
            <p className="text-caption-strong text-muted uppercase">
              {c.label}
            </p>
            <p className="text-display-sm tabular-nums">{c.value}</p>
            {c.sub && (
              <p className="text-caption text-muted mt-auto">{c.sub}</p>
            )}
          </Card>
        </li>
      ))}
    </ul>
  )
}

function UsersTable({
  users,
  onRefresh,
}: {
  users: AdminUserView[]
  onRefresh: () => Promise<void>
}) {
  const { t } = useI18n()
  if (users.length === 0) {
    return (
      <Card variant="feature" className="text-center">
        <p className="text-body-md text-body py-12">{t('admin.users.empty')}</p>
      </Card>
    )
  }
  return (
    <Card variant="feature" className="p-2">
      <ul className="flex flex-col divide-y divide-hairline">
        {users.map((u) => (
          <li key={u.id}>
            <UserRow user={u} onRefresh={onRefresh} />
          </li>
        ))}
      </ul>
    </Card>
  )
}

function UserRow({
  user,
  onRefresh,
}: {
  user: AdminUserView
  onRefresh: () => Promise<void>
}) {
  const { t } = useI18n()
  const { admin } = useGateways()
  const [busy, setBusy] = useState(false)
  const [err, setErr] = useState<string | null>(null)

  const isAdmin = user.role === 'admin'

  const toggle = async () => {
    if (busy) return
    setBusy(true)
    setErr(null)
    try {
      if (isAdmin) {
        await admin.demoteUser(user.id)
      } else {
        await admin.promoteUser(user.id)
      }
      await onRefresh()
    } catch (cause) {
      setErr(messageOf(cause, t('admin.errors.role')))
      setBusy(false)
    }
  }

  return (
    <div className="flex flex-wrap items-center gap-4 px-3 py-3">
      <AssetIconCircular>{initials(user.email)}</AssetIconCircular>
      <div className="min-w-0 flex-1">
        <p className="text-title-sm text-ink truncate">{user.email}</p>
        <p className="text-caption text-muted">
          ID {user.id} · {formatDate(user.createdAt)}
        </p>
      </div>
      <div className="hidden flex-col items-end sm:flex">
        <PriceCell tone="neutral" value={user.vacanciesOwned} />
        <span className="text-caption text-muted">
          {t('admin.users.vacancies')}
        </span>
      </div>
      <div className="hidden flex-col items-end sm:flex">
        <PriceCell tone="neutral" value={user.candidatesUploaded} />
        <span className="text-caption text-muted">
          {t('admin.users.candidates')}
        </span>
      </div>
      {isAdmin ? (
        <BadgePill tone="inverse">{t('admin.users.role.admin')}</BadgePill>
      ) : (
        <BadgePill>{t('admin.users.role.user')}</BadgePill>
      )}
      <Button
        variant={isAdmin ? 'secondary-light' : 'primary'}
        size="md"
        onClick={() => void toggle()}
        loading={busy}
      >
        {isAdmin ? t('admin.users.demote') : t('admin.users.promote')}
      </Button>
      {err && (
        <p
          role="alert"
          className="text-caption text-semantic-down basis-full text-right"
        >
          {err}
        </p>
      )}
    </div>
  )
}

function ErrorCard({ message }: { message: string }) {
  const { t } = useI18n()
  return (
    <Card variant="feature">
      <BadgePill tone="down">{t('common.error')}</BadgePill>
      <p className="text-body-md text-body mt-3">{message}</p>
    </Card>
  )
}

function initials(email: string): string {
  if (!email) return '?'
  return email[0]?.toUpperCase() ?? '?'
}

function formatDate(iso: string): string {
  if (!iso) return ''
  const d = new Date(iso)
  if (Number.isNaN(d.getTime())) return ''
  return d.toLocaleDateString(undefined, {
    day: '2-digit',
    month: 'short',
    year: 'numeric',
  })
}

function messageOf(cause: unknown, fallback: string): string {
  if (cause instanceof ApiError) return cause.message || fallback
  return fallback
}
