import { Link, useNavigate } from 'react-router-dom'
import { useI18n } from '@/app/providers/I18nProvider'
import { BadgePill, Button } from '@/presentation/ui'
import { VacancyForm } from '@/presentation/features/vacancies/VacancyForm'

export function VacancyCreatePage() {
  const { t } = useI18n()
  const navigate = useNavigate()

  return (
    <>
      <section className="bg-canvas">
        <div className="mx-auto max-w-[1200px] px-6 pt-[96px] pb-12">
          <div className="flex items-center gap-3">
            <Link to="/vacancies" className="cursor-pointer">
              <Button variant="secondary-light">{t('create.back')}</Button>
            </Link>
            <BadgePill>{t('create.eyebrow')}</BadgePill>
          </div>
          <h1 className="text-display-md mt-6">{t('create.title')}</h1>
          <p className="text-body-md text-body mt-3 max-w-[560px]">
            {t('create.subtitle')}
          </p>
        </div>
      </section>

      <section className="bg-surface-soft">
        <div className="mx-auto max-w-[820px] px-6 py-12 pb-[96px]">
          <VacancyForm onCancel={() => navigate('/vacancies')} />
        </div>
      </section>
    </>
  )
}
