import { Link } from 'react-router-dom'
import { useI18n } from '@/app/providers/I18nProvider'
import { BadgePill, Card } from '@/presentation/ui'
import { KNOWN_ROLE_VALUES, type Vacancy } from '@/domain/vacancy/types'
import { formatDate } from '@/shared/lib/format'
import { VacancyStatusBadge } from './VacancyStatusBadge'

export function VacancyCard({ vacancy }: { vacancy: Vacancy }) {
  const { t, locale } = useI18n()
  const skillsCount = vacancy.skills.length
  const mustHave = vacancy.skills.filter((s) => s.mustHave).length
  const roleLabel = KNOWN_ROLE_VALUES.has(vacancy.role)
    ? t(`roles.${vacancy.role}`)
    : vacancy.role

  return (
    <Link
      to={`/vacancies/${vacancy.id}`}
      className="group block cursor-pointer rounded-xl transition-colors focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary"
    >
      <Card
        variant="feature"
        className="flex h-full flex-col gap-6 transition-colors group-hover:border-ink"
      >
        <div className="flex items-start justify-between gap-3">
          <BadgePill>{roleLabel}</BadgePill>
          <VacancyStatusBadge status={vacancy.status} />
        </div>

        <div className="flex flex-col gap-2">
          <h3 className="text-title-md text-ink line-clamp-2">
            {vacancy.title}
          </h3>
          {vacancy.description && (
            <p className="text-body-sm text-body line-clamp-3">
              {vacancy.description}
            </p>
          )}
        </div>

        <div className="mt-auto flex items-end justify-between gap-3">
          <div>
            <p className="text-caption-strong text-muted uppercase">
              {t('skills.legend')}
            </p>
            <p className="text-number-display text-ink mt-1">
              {skillsCount}
              {mustHave > 0 && (
                <span className="text-caption text-muted ml-2">
                  · {mustHave} {t('skills.must').toLowerCase()}
                </span>
              )}
            </p>
          </div>
          <span className="text-caption text-muted">
            {formatDate(vacancy.createdAt, locale)}
          </span>
        </div>
      </Card>
    </Link>
  )
}
