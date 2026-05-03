import { useEffect, useState } from 'react'
import { cn } from '@/shared/lib/cn'
import { MinusIcon, PlusIcon } from './icons'
import { formatWeight, parseWeight } from './weight'

/**
 * Weighted-input control: `[−] 0.05 [+]` with the centre being a free-form
 * text field so users can type "0.06" or "0.1" without fighting browser
 * step validation. Local draft state holds the literal string the user is
 * typing so intermediate forms ("0.", "0.0") survive re-renders. We commit
 * to the parent on blur, Enter, or +/− clicks.
 */
type Props = {
  value: number
  onChange: (next: number) => void
  step?: number
  min?: number
  max?: number
  error?: string
  ariaLabel?: string
  disabled?: boolean
  className?: string
}

const DEFAULT_STEP = 0.05

export function WeightStepper({
  value,
  onChange,
  step = DEFAULT_STEP,
  min = 0,
  max = 1,
  error,
  ariaLabel,
  disabled,
  className,
}: Props) {
  const [draft, setDraft] = useState(() => formatWeight(value))

  useEffect(() => {
    const parsed = parseWeight(draft)
    if (parsed === null || Math.abs(parsed - value) > 1e-9) {
      setDraft(formatWeight(value))
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [value])

  const clamp = (n: number) => Math.min(max, Math.max(min, n))
  const round = (n: number) => Math.round(n * 100) / 100

  const commit = (raw: string) => {
    const parsed = parseWeight(raw)
    if (parsed === null) {
      setDraft(formatWeight(value))
      return
    }
    const next = round(clamp(parsed))
    onChange(next)
    setDraft(formatWeight(next))
  }

  const dec = () => onChange(round(clamp(value - step)))
  const inc = () => onChange(round(clamp(value + step)))

  const invalid = Boolean(error)

  return (
    <div className={cn('flex flex-col gap-1.5', className)}>
      <div
        role="group"
        aria-label={ariaLabel}
        className={cn(
          'inline-flex h-12 items-stretch overflow-hidden rounded-md bg-canvas',
          'border border-hairline transition-colors',
          'focus-within:border-primary focus-within:shadow-[inset_0_0_0_1px_var(--color-primary)]',
          invalid &&
            'border-semantic-down focus-within:border-semantic-down focus-within:shadow-[inset_0_0_0_1px_var(--color-semantic-down)]',
          disabled && 'opacity-60',
        )}
      >
        <StepperButton
          onClick={dec}
          disabled={disabled || value <= min}
          ariaLabel="−"
        >
          <MinusIcon />
        </StepperButton>
        <input
          type="text"
          inputMode="decimal"
          aria-label={ariaLabel}
          aria-invalid={invalid || undefined}
          disabled={disabled}
          value={draft}
          onChange={(e) => setDraft(e.target.value)}
          onBlur={(e) => commit(e.target.value)}
          onKeyDown={(e) => {
            if (e.key === 'Enter') {
              e.preventDefault()
              commit(e.currentTarget.value)
              e.currentTarget.blur()
            }
          }}
          className={cn(
            'text-number-display w-full bg-transparent text-center tabular-nums',
            'text-ink focus:outline-none',
          )}
        />
        <StepperButton
          onClick={inc}
          disabled={disabled || value >= max}
          ariaLabel="+"
        >
          <PlusIcon />
        </StepperButton>
      </div>
      {error && <p className="text-caption text-semantic-down">{error}</p>}
    </div>
  )
}

function StepperButton({
  children,
  onClick,
  disabled,
  ariaLabel,
}: {
  children: React.ReactNode
  onClick: () => void
  disabled?: boolean
  ariaLabel: string
}) {
  return (
    <button
      type="button"
      onClick={onClick}
      disabled={disabled}
      aria-label={ariaLabel}
      className={cn(
        'flex w-9 shrink-0 cursor-pointer items-center justify-center text-ink transition-colors',
        'hover:bg-surface-strong disabled:cursor-not-allowed disabled:text-muted-soft disabled:hover:bg-transparent',
        'focus-visible:outline-2 focus-visible:outline-[-2px] focus-visible:outline-primary',
      )}
    >
      {children}
    </button>
  )
}
