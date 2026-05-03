import { useI18n } from '@/app/providers/I18nProvider'
import { BadgePill } from '@/presentation/ui'
import type { VacancyStatus } from '@/domain/vacancy/types'

/**
 * Single source of truth for vacancy-status pills. Was duplicated inline
 * in VacancyCard and VacancyDetailsPage — same switch, same tones.
 * Returns `null` for `unknown` so callers don't need to special-case
 * empty render.
 */
export function VacancyStatusBadge({ status }: { status: VacancyStatus }) {
  const { t } = useI18n()
  switch (status) {
    case 'open':
      return <BadgePill tone="up">{t('status.open')}</BadgePill>
    case 'archived':
      return <BadgePill tone="down">{t('status.archived')}</BadgePill>
    case 'draft':
      return <BadgePill>{t('status.draft')}</BadgePill>
    default:
      return null
  }
}
