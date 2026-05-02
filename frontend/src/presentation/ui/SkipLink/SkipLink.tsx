import { useI18n } from '@/app/providers/I18nProvider'

/**
 * Visually hidden link that becomes visible on focus, allowing keyboard
 * users to bypass the nav and jump straight to main content. Targets the
 * id passed in (`#main-content` by convention).
 */
export function SkipLink({ targetId = 'main-content' }: { targetId?: string }) {
  const { t } = useI18n()
  return (
    <a
      href={`#${targetId}`}
      className="sr-only focus:not-sr-only focus:fixed focus:top-2 focus:left-2 focus:z-50 focus:rounded-pill focus:bg-primary focus:px-4 focus:py-2 focus:text-on-primary focus:shadow-soft focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-on-primary"
    >
      {t('a11y.skipToMain')}
    </a>
  )
}
