import { useI18n } from '@/app/providers/I18nProvider'
import { BadgePill } from '@/presentation/ui'

/**
 * Maps the LLM's free-form `hr_recommendation` string into one of three
 * tones. The model is meant to return `"hire" | "maybe" | "no"`, but we
 * defend against synonyms ("yes" / "reject") so a chatty model doesn't
 * fall through to "Maybe".
 */
export function RecommendationBadge({ value }: { value: string }) {
  const { t } = useI18n()
  const v = value.toLowerCase()
  if (v.includes('hire') || v === 'yes')
    return <BadgePill tone="up">{t('rec.hire')}</BadgePill>
  if (v.includes('no') || v.includes('reject'))
    return <BadgePill tone="down">{t('rec.no')}</BadgePill>
  return <BadgePill>{t('rec.maybe')}</BadgePill>
}
