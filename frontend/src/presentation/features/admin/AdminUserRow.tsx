import { useState } from 'react'
import { useGateways } from '@/app/providers/GatewaysProvider'
import { useI18n } from '@/app/providers/I18nProvider'
import { ApiError } from '@/infrastructure/http/errors'
import {
  AssetIconCircular,
  BadgePill,
  Button,
  PriceCell,
} from '@/presentation/ui'
import type { AdminUserView } from '@/domain/admin/types'
import { formatDate, initials } from '@/shared/lib/format'

/**
 * One row in the admin user table. Owns its own promote/demote busy +
 * error state so a slow auth call on one row doesn't lock the rest of
 * the table. Calls `onRefresh` on success so the parent re-fetches and
 * the badge + button flip in unison.
 */
type Props = {
  user: AdminUserView
  onRefresh: () => Promise<void>
}

export function AdminUserRow({ user, onRefresh }: Props) {
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
      if (isAdmin) await admin.demoteUser(user.id)
      else await admin.promoteUser(user.id)
      await onRefresh()
    } catch (cause) {
      const message =
        cause instanceof ApiError ? cause.message : t('admin.errors.role')
      setErr(message)
    } finally {
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
