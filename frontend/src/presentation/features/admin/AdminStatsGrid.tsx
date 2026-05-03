import { useI18n } from '@/app/providers/I18nProvider'
import { Card } from '@/presentation/ui'
import type { SystemStats } from '@/domain/admin/types'

/**
 * Top-of-dashboard 4-card grid. `stats` may be null while the request
 * is in flight (or has failed at the page level) — empty cards still
 * render so the layout doesn't jump.
 */
export function AdminStatsGrid({ stats }: { stats: SystemStats | null }) {
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
