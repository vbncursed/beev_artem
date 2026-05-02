import { useEffect, useId, useRef, useState, type ComponentType } from 'react'
import { useI18n } from '@/app/providers/I18nProvider'
import { FlagRu, FlagUs } from '@/presentation/ui/Flag'
import { cn } from '@/shared/lib/cn'
import type { Locale } from '@/shared/i18n/dictionaries'

/**
 * Language dropdown. Trigger uses pill geometry from
 * `search-input-pill` (h-44, rounded-pill, surface-strong + hairline);
 * menu uses card-radius (rounded-md) + the single hairline border tier
 * + soft shadow. Primary stays scarce — only the focus ring uses it.
 */
type FlagComponent = ComponentType<{ className?: string }>

type Option = {
  value: Locale
  label: string
  Flag: FlagComponent
}

export function LanguageSwitcher({ className }: { className?: string }) {
  const { locale, setLocale, t } = useI18n()
  const [open, setOpen] = useState(false)
  const containerRef = useRef<HTMLDivElement>(null)
  const triggerId = useId()
  const menuId = `${triggerId}-menu`

  const options: Option[] = [
    { value: 'ru', label: t('locale.ru'), Flag: FlagRu },
    { value: 'en', label: t('locale.en'), Flag: FlagUs },
  ]

  const current = options.find((o) => o.value === locale) ?? options[0]

  // Close on outside click and Escape
  useEffect(() => {
    if (!open) return
    const onPointer = (e: PointerEvent) => {
      if (!containerRef.current?.contains(e.target as Node)) setOpen(false)
    }
    const onKey = (e: KeyboardEvent) => {
      if (e.key === 'Escape') setOpen(false)
    }
    document.addEventListener('pointerdown', onPointer)
    document.addEventListener('keydown', onKey)
    return () => {
      document.removeEventListener('pointerdown', onPointer)
      document.removeEventListener('keydown', onKey)
    }
  }, [open])

  const choose = (next: Locale) => {
    setLocale(next)
    setOpen(false)
  }

  return (
    <div ref={containerRef} className={cn('relative', className)}>
      <button
        id={triggerId}
        type="button"
        onClick={() => setOpen((v) => !v)}
        aria-haspopup="listbox"
        aria-expanded={open}
        aria-controls={menuId}
        aria-label={t('locale.label')}
        className={cn(
          'inline-flex h-11 cursor-pointer items-center gap-2 rounded-pill px-3 transition-colors',
          'bg-surface-strong text-ink',
          'border border-hairline',
          'focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary',
          open && 'border-primary',
        )}
      >
        <current.Flag />
        <span className="text-caption-strong uppercase tracking-wide">
          {current.value}
        </span>
        <ChevronDown open={open} />
      </button>

      {open && (
        <ul
          id={menuId}
          role="listbox"
          aria-labelledby={triggerId}
          className={cn(
            'absolute right-0 z-40 mt-2 min-w-[180px] overflow-hidden',
            'rounded-md border border-hairline bg-canvas shadow-soft',
            'flex flex-col py-1',
          )}
        >
          {options.map((opt) => {
            const active = opt.value === locale
            return (
              <li key={opt.value}>
                <button
                  type="button"
                  role="option"
                  aria-selected={active}
                  onClick={() => choose(opt.value)}
                  className={cn(
                    'flex w-full cursor-pointer items-center gap-3 px-3 py-2 text-left transition-colors',
                    'hover:bg-surface-soft',
                    'focus-visible:outline-2 focus-visible:outline-[-2px] focus-visible:outline-primary',
                    active && 'bg-surface-soft',
                  )}
                >
                  <opt.Flag />
                  <span className="text-body-sm text-ink flex-1">
                    {opt.label}
                  </span>
                  {active && <Check />}
                </button>
              </li>
            )
          })}
        </ul>
      )}
    </div>
  )
}

function ChevronDown({ open }: { open: boolean }) {
  return (
    <svg
      width="12"
      height="12"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2.5"
      strokeLinecap="round"
      strokeLinejoin="round"
      aria-hidden
      className={cn(
        'text-muted transition-transform duration-150',
        open && 'rotate-180',
      )}
    >
      <path d="m6 9 6 6 6-6" />
    </svg>
  )
}

function Check() {
  return (
    <svg
      width="14"
      height="14"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2.5"
      strokeLinecap="round"
      strokeLinejoin="round"
      aria-hidden
      className="text-primary"
    >
      <path d="M20 6 9 17l-5-5" />
    </svg>
  )
}
