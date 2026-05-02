import { useEffect, useState } from 'react'
import { cn } from '@/shared/lib/cn'

/**
 * Weighted-input control: `[−] 0.05 [+]` with the centre being a free-form
 * text field so users can type "0.06" or "0.1" without fighting browser
 * step validation.
 *
 * Why this exists vs <input type="number" step=...>:
 *   - the browser validates against `step` and rejects values that don't
 *     fit the grid (typing "0.06" with step=0.05 wipes the input);
 *   - the spinner UI on `type=number` is browser-themed; we need ours
 *     to match DESIGN.md geometry (rounded-md, hairline border).
 *
 * Local draft state holds the literal string the user is typing so
 * intermediate forms ("0.", "0.0") survive re-renders. We commit to the
 * parent on blur, Enter, or +/− clicks.
 */
type Props = {
  value: number
  onChange: (next: number) => void
  /** Increment per click on +/−. Default 0.05. */
  step?: number
  /** Inclusive bounds — defaults match the backend rule (0..1). */
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

  // Keep the input in sync when the parent commits a new value (via +/−
  // or because the row was cleared/replaced). Only resync if the parsed
  // draft no longer matches — avoids stomping on mid-typing strings.
  useEffect(() => {
    const parsed = parse(draft)
    if (parsed === null || Math.abs(parsed - value) > 1e-9) {
      setDraft(formatWeight(value))
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [value])

  const clamp = (n: number) => Math.min(max, Math.max(min, n))
  const round = (n: number) => Math.round(n * 100) / 100

  const commit = (raw: string) => {
    const parsed = parse(raw)
    if (parsed === null) {
      setDraft(formatWeight(value))
      return
    }
    const next = round(clamp(parsed))
    onChange(next)
    setDraft(formatWeight(next))
  }

  const dec = () => {
    const next = round(clamp(value - step))
    onChange(next)
  }
  const inc = () => {
    const next = round(clamp(value + step))
    onChange(next)
  }

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
          <Minus />
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
          <Plus />
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

function Minus() {
  return (
    <svg
      width="14"
      height="14"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2.5"
      strokeLinecap="round"
      aria-hidden
    >
      <path d="M5 12h14" />
    </svg>
  )
}

function Plus() {
  return (
    <svg
      width="14"
      height="14"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2.5"
      strokeLinecap="round"
      aria-hidden
    >
      <path d="M12 5v14M5 12h14" />
    </svg>
  )
}

function parse(raw: string): number | null {
  const trimmed = raw.trim().replace(',', '.')
  if (!trimmed) return null
  const n = Number(trimmed)
  return Number.isFinite(n) ? n : null
}

function formatWeight(n: number): string {
  if (!Number.isFinite(n)) return '0'
  // Two decimals, then strip trailing zeros so "0.10" → "0.1" and "0" stays "0".
  const s = n.toFixed(2)
  return s.replace(/\.?0+$/, '') || '0'
}
