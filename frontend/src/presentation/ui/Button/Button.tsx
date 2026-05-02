import { forwardRef, type ButtonHTMLAttributes, type ReactNode } from 'react'
import { cn } from '@/shared/lib/cn'

/**
 * DESIGN.md mapping:
 *   primary           → button-primary       (44px pill, #0052ff)
 *   primary-cta       → button-pill-cta      (56px pill, hero CTA)
 *   secondary-light   → button-secondary-light
 *   secondary-dark    → button-secondary-dark
 *   outline-on-dark   → button-outline-on-dark
 *   text              → button-tertiary-text (inline link, primary text)
 *
 * All variants are pill-rounded (rounded-pill / 100px). Sharp corners forbidden.
 */
export type ButtonVariant =
  | 'primary'
  | 'primary-cta'
  | 'secondary-light'
  | 'secondary-dark'
  | 'outline-on-dark'
  | 'text'

export type ButtonSize = 'md' | 'lg'

type ButtonOwnProps = {
  variant?: ButtonVariant
  size?: ButtonSize
  loading?: boolean
  iconLeft?: ReactNode
  iconRight?: ReactNode
  fullWidth?: boolean
}

export type ButtonProps = ButtonOwnProps &
  Omit<ButtonHTMLAttributes<HTMLButtonElement>, keyof ButtonOwnProps>

const VARIANT_CLASSES: Record<ButtonVariant, string> = {
  primary:
    'bg-primary text-on-primary hover:bg-primary-active active:bg-primary-active disabled:bg-primary-disabled',
  'primary-cta':
    'bg-primary text-on-primary hover:bg-primary-active active:bg-primary-active disabled:bg-primary-disabled',
  'secondary-light':
    'bg-surface-strong text-ink hover:bg-hairline disabled:opacity-50',
  'secondary-dark':
    'bg-surface-dark-elevated text-on-dark hover:opacity-90 disabled:opacity-50',
  'outline-on-dark':
    'bg-transparent text-on-dark border border-on-dark/40 hover:border-on-dark disabled:opacity-50',
  text: 'bg-transparent text-primary hover:opacity-80 disabled:opacity-50 px-0',
}

const SIZE_CLASSES: Record<ButtonSize, string> = {
  md: 'h-11 px-5',
  lg: 'h-14 px-8',
}

export const Button = forwardRef<HTMLButtonElement, ButtonProps>(
  function Button(
    {
      variant = 'primary',
      size,
      loading = false,
      iconLeft,
      iconRight,
      fullWidth = false,
      className,
      disabled,
      type = 'button',
      children,
      ...rest
    },
    ref,
  ) {
    const resolvedSize: ButtonSize =
      size ?? (variant === 'primary-cta' ? 'lg' : 'md')

    const isText = variant === 'text'

    return (
      <button
        ref={ref}
        type={type}
        disabled={disabled || loading}
        aria-busy={loading || undefined}
        className={cn(
          'text-button inline-flex shrink-0 items-center justify-center gap-2 rounded-pill transition-colors duration-150',
          'cursor-pointer disabled:cursor-not-allowed',
          'focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary',
          !isText && SIZE_CLASSES[resolvedSize],
          VARIANT_CLASSES[variant],
          fullWidth && 'w-full',
          className,
        )}
        {...rest}
      >
        {loading ? (
          <span
            className="inline-block size-4 animate-spin rounded-full border-2 border-current border-t-transparent"
            aria-hidden="true"
          />
        ) : (
          iconLeft && <span className="inline-flex">{iconLeft}</span>
        )}
        {children}
        {!loading && iconRight && (
          <span className="inline-flex">{iconRight}</span>
        )}
      </button>
    )
  },
)
