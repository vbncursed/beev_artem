import { useI18n } from '@/app/providers/I18nProvider'
import { LegalLayout } from './LegalLayout'
import { SupportBodyEn } from './SupportBodyEn'
import { SupportBodyRu } from './SupportBodyRu'

export function SupportPage() {
  const { locale, t } = useI18n()
  return (
    <LegalLayout
      eyebrow={t('legal.eyebrow.support')}
      title={t('legal.support.title')}
      subtitle={t('legal.support.subtitle')}
    >
      {locale === 'ru' ? <SupportBodyRu /> : <SupportBodyEn />}
    </LegalLayout>
  )
}
