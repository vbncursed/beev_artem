import { useId, type InputHTMLAttributes, type ReactNode } from 'react'
import { cn } from '@/shared/lib/cn'

/**
 * DESIGN-friendly checkbox. The native input stays in the DOM (a11y +
 * keyboard navigation) but is visually replaced by a styled box that
 * follows the system tokens: hairline-bordered canvas when off, primary
 * fill when on. No second brand color introduced.
 *
 * Use `<Checkbox label="Must" checked={...} onChange={...} />`.
 */
type OwnProps = {
  label: ReactNode
  checked: boolean
  onChange: (checked: boolean) => void
  /** Adds a visible focus ring even when checked is toggled via mouse. */
  className?: string
}

export type CheckboxProps = OwnProps &
  Omit<InputHTMLAttributes<HTMLInputElement>, keyof OwnProps | 'type'>

export function Checkbox({
  label,
  checked,
  onChange,
  className,
  disabled,
  ...rest
}: CheckboxProps) {
  const reactId = useId()
  const id = rest.id ?? reactId

  return (
    <label
      htmlFor={id}
      className={cn(
        'inline-flex select-none items-center gap-2',
        disabled ? 'cursor-not-allowed opacity-60' : 'cursor-pointer',
        className,
      )}
    >
      <input
        {...rest}
        id={id}
        type="checkbox"
        className="peer sr-only"
        checked={checked}
        disabled={disabled}
        onChange={(e) => onChange(e.target.checked)}
      />
      <span
        aria-hidden
        className={cn(
          'inline-flex size-[18px] shrink-0 items-center justify-center rounded-xs transition-colors',
          'border',
          checked
            ? 'bg-primary border-primary text-on-primary'
            : 'bg-canvas border-hairline text-transparent',
          'peer-focus-visible:outline-2 peer-focus-visible:outline-offset-2 peer-focus-visible:outline-primary',
        )}
      >
        <CheckGlyph />
      </span>
      <span className="text-caption-strong text-ink uppercase tracking-wide">
        {label}
      </span>
    </label>
  )
}

function CheckGlyph() {
  return (
    <svg
      width="12"
      height="12"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="3"
      strokeLinecap="round"
      strokeLinejoin="round"
    >
      <path d="M20 6 9 17l-5-5" />
    </svg>
  )
}
