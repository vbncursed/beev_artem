import { useI18n } from '@/app/providers/I18nProvider'
import { useTheme } from '@/app/providers/ThemeProvider'
import { cn } from '@/shared/lib/cn'

/**
 * Theme toggle. SVG icons only (no emoji per DESIGN). Pill geometry from
 * `search-input-pill` family: 44px height, rounded-pill, surface-strong fill.
 * Active half slides with a translate transform — no extra brand color.
 */
const SunIcon = () => (
  <svg
    width="16"
    height="16"
    viewBox="0 0 24 24"
    fill="none"
    stroke="currentColor"
    strokeWidth="2"
    strokeLinecap="round"
    strokeLinejoin="round"
    aria-hidden
  >
    <circle cx="12" cy="12" r="4" />
    <path d="M12 2v2M12 20v2M4.93 4.93l1.41 1.41M17.66 17.66l1.41 1.41M2 12h2M20 12h2M6.34 17.66l-1.41 1.41M19.07 4.93l-1.41 1.41" />
  </svg>
)

const MoonIcon = () => (
  <svg
    width="16"
    height="16"
    viewBox="0 0 24 24"
    fill="none"
    stroke="currentColor"
    strokeWidth="2"
    strokeLinecap="round"
    strokeLinejoin="round"
    aria-hidden
  >
    <path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79Z" />
  </svg>
)

export function ThemeToggle({ className }: { className?: string }) {
  const { theme, toggle } = useTheme()
  const { t } = useI18n()
  const isDark = theme === 'dark'

  return (
    <button
      type="button"
      onClick={toggle}
      aria-label={isDark ? t('theme.toLight') : t('theme.toDark')}
      aria-pressed={isDark}
      className={cn(
        'relative inline-flex h-11 w-[88px] cursor-pointer items-center rounded-pill bg-surface-strong p-1 transition-colors',
        'focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary',
        className,
      )}
    >
      <span
        aria-hidden
        className={cn(
          'absolute top-1 left-1 inline-flex size-9 items-center justify-center rounded-full bg-canvas text-ink shadow-soft transition-transform duration-200 ease-out',
          isDark && 'translate-x-[44px]',
        )}
      >
        {isDark ? <MoonIcon /> : <SunIcon />}
      </span>
      <span className="ml-1 inline-flex size-9 items-center justify-center text-muted">
        <SunIcon />
      </span>
      <span className="inline-flex size-9 items-center justify-center text-muted">
        <MoonIcon />
      </span>
    </button>
  )
}
