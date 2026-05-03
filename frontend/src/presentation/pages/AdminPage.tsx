import { useI18n } from '@/app/providers/I18nProvider'
import { BadgePill, ErrorCard, Spinner } from '@/presentation/ui'
import { AdminStatsGrid } from '@/presentation/features/admin/AdminStatsGrid'
import { AdminUsersTable } from '@/presentation/features/admin/AdminUsersTable'
import { useAdminDashboard } from '@/presentation/features/admin/useAdminDashboard'

/**
 * Admin dashboard at `/admin`. Pure layout — data fetching is delegated
 * to `useAdminDashboard`, the stats grid + user table to their own
 * feature components. Adding a new section reduces to a new feature
 * file + one block here.
 */
export function AdminPage() {
  const { t } = useI18n()
  const { stats, users, statsErr, usersErr, loading, refresh } =
    useAdminDashboard()

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
            <AdminStatsGrid stats={stats} />
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
            <AdminUsersTable users={users} onRefresh={refresh} />
          )}
        </div>
      </section>
    </>
  )
}
