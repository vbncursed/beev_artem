import { useI18n } from '@/app/providers/I18nProvider'
import { Card } from '@/presentation/ui'
import type { AdminUserView } from '@/domain/admin/types'
import { AdminUserRow } from './AdminUserRow'

/**
 * Container card for the admin user list. Shows an empty-state when the
 * server returns zero users (rare but defensive — happens on a brand-new
 * deploy before the first registration).
 */
type Props = {
  users: AdminUserView[]
  onRefresh: () => Promise<void>
}

export function AdminUsersTable({ users, onRefresh }: Props) {
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
            <AdminUserRow user={u} onRefresh={onRefresh} />
          </li>
        ))}
      </ul>
    </Card>
  )
}
