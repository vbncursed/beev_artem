import { useState } from 'react'
import { Link } from 'react-router-dom'
import { useI18n } from '@/app/providers/I18nProvider'
import {
  BadgePill,
  Button,
  Card,
  ErrorCard,
  SearchInput,
  Spinner,
} from '@/presentation/ui'
import { KNOWN_ROLES } from '@/domain/vacancy/types'
import { pluralKey } from '@/shared/i18n/dictionaries'
import { useDebouncedValue } from '@/shared/hooks/useDebouncedValue'
import { useVacancies } from '@/presentation/features/vacancies/useVacancies'
import { VacancyCard } from '@/presentation/features/vacancies/VacancyCard'

const ROLE_FILTERS = [
  { value: 'all', labelKey: 'roles.all' },
  ...KNOWN_ROLES.map((r) => ({ value: r.value, labelKey: `roles.${r.value}` })),
] as const

type RoleFilter = (typeof ROLE_FILTERS)[number]['value']

export function VacanciesPage() {
  const { t, locale } = useI18n()
  const [rawQuery, setRawQuery] = useState('')
  const [role, setRole] = useState<RoleFilter>('all')
  const debouncedQuery = useDebouncedValue(rawQuery, 250)

  const { vacancies, isLoading, error, total } = useVacancies(debouncedQuery)

  const filtered =
    role === 'all'
      ? vacancies
      : vacancies.filter((v) => (v.role || 'default') === role)

  return (
    <>
      <section className="bg-canvas">
        <div className="mx-auto flex max-w-[1200px] flex-col gap-10 px-6 pt-[96px] pb-12">
          <div className="flex flex-col gap-6 md:flex-row md:items-end md:justify-between">
            <div>
              <BadgePill>{t('vacancies.eyebrow')}</BadgePill>
              <h1 className="text-display-lg mt-4">{t('vacancies.title')}</h1>
              <p className="text-body-md text-body mt-3 max-w-[520px]">
                {t('vacancies.subtitle')}
              </p>
            </div>
            <Link to="/vacancies/new" className="self-start md:self-end">
              <Button variant="primary-cta">{t('vacancies.new')}</Button>
            </Link>
          </div>

          <div className="flex flex-col gap-4 md:flex-row md:items-center md:justify-between">
            <SearchInput
              placeholder={t('vacancies.searchPlaceholder')}
              value={rawQuery}
              onChange={(e) => setRawQuery(e.target.value)}
              className="md:w-[420px]"
            />
            <RoleFilterChips value={role} onChange={setRole} />
          </div>
        </div>
      </section>

      <section className="bg-surface-soft">
        <div className="mx-auto max-w-[1200px] px-6 py-12 pb-[96px]">
          <div className="flex items-center justify-between pb-6">
            <p className="text-caption-strong text-muted uppercase">
              {isLoading && total === 0
                ? t('common.loading')
                : t(pluralKey('vacancies.count', total, locale), {
                    n: total,
                  })}
            </p>
            {isLoading && <Spinner size={16} />}
          </div>

          {error ? (
            <ErrorCard
              message={error}
              title={t('vacancies.error.title')}
            />
          ) : filtered.length === 0 && !isLoading ? (
            <EmptyState query={debouncedQuery} role={role} />
          ) : (
            <ul className="grid grid-cols-1 gap-6 md:grid-cols-2 xl:grid-cols-3">
              {filtered.map((v) => (
                <li key={v.id}>
                  <VacancyCard vacancy={v} />
                </li>
              ))}
            </ul>
          )}
        </div>
      </section>
    </>
  )
}

function RoleFilterChips({
  value,
  onChange,
}: {
  value: RoleFilter
  onChange: (next: RoleFilter) => void
}) {
  const { t } = useI18n()
  return (
    <div
      role="group"
      aria-label="Role filter"
      className="flex flex-wrap items-center gap-2"
    >
      {ROLE_FILTERS.map((opt) => {
        const active = opt.value === value
        return (
          <button
            key={opt.value}
            type="button"
            aria-pressed={active}
            onClick={() => onChange(opt.value)}
            className={
              'text-caption-strong inline-flex h-8 cursor-pointer items-center rounded-pill px-3 uppercase tracking-wide transition-colors ' +
              'focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary ' +
              (active
                ? 'bg-surface-dark text-on-dark'
                : 'bg-surface-strong text-ink hover:bg-hairline')
            }
          >
            {t(opt.labelKey)}
          </button>
        )
      })}
    </div>
  )
}

function EmptyState({
  query,
  role,
}: {
  query: string
  role: RoleFilter
}) {
  const { t } = useI18n()
  const isFiltered = Boolean(query) || role !== 'all'
  return (
    <Card variant="feature" className="text-center">
      <div className="mx-auto flex max-w-[420px] flex-col items-center gap-3 py-12">
        <BadgePill>
          {isFiltered
            ? t('vacancies.empty.badgeFiltered')
            : t('vacancies.empty.badge')}
        </BadgePill>
        <h3 className="text-title-lg">
          {isFiltered
            ? t('vacancies.empty.filteredTitle')
            : t('vacancies.empty.title')}
        </h3>
        <p className="text-body-md text-body">
          {isFiltered
            ? t('vacancies.empty.filteredHint')
            : t('vacancies.empty.hint')}
        </p>
        {!isFiltered && (
          <Link to="/vacancies/new" className="mt-2">
            <Button variant="primary">{t('vacancies.empty.cta')}</Button>
          </Link>
        )}
      </div>
    </Card>
  )
}

