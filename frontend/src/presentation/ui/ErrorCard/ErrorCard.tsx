import type { ReactNode } from 'react'
import { useI18n } from '@/app/providers/I18nProvider'
import { BadgePill } from '@/presentation/ui/BadgePill'
import { Card } from '@/presentation/ui/Card'

/**
 * Generic "something went wrong" card used by listing pages when a
 * fetch fails. Three variants of the same shape lived inline in
 * VacanciesPage, AdminPage, and elsewhere — collapsed here to one.
 *
 * `title` defaults to the localised `common.error` label; pass it
 * explicitly when the page needs a more specific framing
 * ("We couldn't load vacancies", etc.).
 */
type Props = {
  message: string
  title?: ReactNode
  /** Override the down-tone "Error" badge label. */
  badge?: ReactNode
  /** Extra body slot rendered after the message — e.g. a retry button. */
  children?: ReactNode
}

export function ErrorCard({ message, title, badge, children }: Props) {
  const { t } = useI18n()
  return (
    <Card variant="feature" className="text-center">
      <div className="mx-auto flex max-w-[420px] flex-col items-center gap-3 py-12">
        <BadgePill tone="down">{badge ?? t('common.error')}</BadgePill>
        {title && <h3 className="text-title-lg">{title}</h3>}
        <p className="text-body-md text-body break-words">{message}</p>
        {children}
      </div>
    </Card>
  )
}
