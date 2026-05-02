import { useId } from 'react'
import { cn } from '@/shared/lib/cn'

/**
 * Segmented pill switcher. Geometry follows DESIGN.md `search-input-pill`:
 * bg surface-strong, rounded-pill, height 44px. Active segment fills with
 * primary on `tone="primary"` (auth/login/register), or inverts to surface-dark
 * on `tone="inverse"` (used as filter chips on top of light bands).
 *
 * No second brand color introduced — primary stays the only voltage.
 */
export type PillSwitcherTone = 'primary' | 'inverse'

export type PillSwitcherOption<T extends string> = {
  value: T
  label: string
}

export type PillSwitcherProps<T extends string> = {
  options: ReadonlyArray<PillSwitcherOption<T>>
  value: T
  onChange: (next: T) => void
  tone?: PillSwitcherTone
  ariaLabel?: string
  className?: string
}

export function PillSwitcher<T extends string>({
  options,
  value,
  onChange,
  tone = 'primary',
  ariaLabel,
  className,
}: PillSwitcherProps<T>) {
  const groupId = useId()

  const activeFill =
    tone === 'primary'
      ? 'bg-primary text-on-primary'
      : 'bg-surface-dark text-on-dark'

  return (
    <div
      role="tablist"
      aria-label={ariaLabel}
      className={cn(
        'inline-flex h-11 items-center gap-1 rounded-pill bg-surface-strong p-1',
        className,
      )}
    >
      {options.map((opt) => {
        const active = opt.value === value
        return (
          <button
            key={opt.value}
            id={`${groupId}-${opt.value}`}
            type="button"
            role="tab"
            aria-selected={active}
            tabIndex={active ? 0 : -1}
            onClick={() => onChange(opt.value)}
            className={cn(
              'text-button h-full cursor-pointer rounded-pill px-5 transition-colors duration-150',
              'focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary',
              active
                ? activeFill
                : 'bg-transparent text-body hover:text-ink',
            )}
          >
            {opt.label}
          </button>
        )
      })}
    </div>
  )
}
