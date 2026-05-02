import {
  forwardRef,
  useId,
  type InputHTMLAttributes,
  type ReactNode,
} from 'react'
import { cn } from '@/shared/lib/cn'

/**
 * DESIGN.md → text-input:
 *   bg canvas, rounded-md (12px), padding 14×16, height 48px,
 *   1px hairline border, focus → 2px primary border.
 *
 * Errors render as semantic-down text below the field — never red background.
 */
type TextInputOwnProps = {
  label?: string
  hint?: string
  error?: string
  iconLeft?: ReactNode
  iconRight?: ReactNode
}

export type TextInputProps = TextInputOwnProps &
  Omit<InputHTMLAttributes<HTMLInputElement>, keyof TextInputOwnProps>

export const TextInput = forwardRef<HTMLInputElement, TextInputProps>(
  function TextInput(
    { label, hint, error, iconLeft, iconRight, className, id, ...rest },
    ref,
  ) {
    const reactId = useId()
    const inputId = id ?? reactId
    const hintId = `${inputId}-hint`
    const errorId = `${inputId}-error`
    const invalid = Boolean(error)

    return (
      <div className={cn('flex flex-col gap-1.5', className)}>
        {label && (
          <label
            htmlFor={inputId}
            className="text-caption-strong text-body-strong"
          >
            {label}
          </label>
        )}
        <div
          className={cn(
            'relative flex h-12 items-center rounded-md bg-canvas',
            'border border-hairline',
            'focus-within:border-primary focus-within:shadow-[inset_0_0_0_1px_var(--color-primary)]',
            invalid &&
              'border-semantic-down focus-within:border-semantic-down focus-within:shadow-[inset_0_0_0_1px_var(--color-semantic-down)]',
            rest.disabled && 'opacity-60',
          )}
        >
          {iconLeft && (
            <span className="pl-4 text-muted" aria-hidden>
              {iconLeft}
            </span>
          )}
          <input
            ref={ref}
            id={inputId}
            aria-invalid={invalid || undefined}
            aria-describedby={
              error ? errorId : hint ? hintId : undefined
            }
            className={cn(
              'text-body-md h-full w-full bg-transparent px-4 text-ink placeholder:text-muted',
              'focus:outline-none',
              iconLeft ? 'pl-2' : undefined,
              iconRight ? 'pr-2' : undefined,
            )}
            {...rest}
          />
          {iconRight && (
            <span className="pr-4 text-muted" aria-hidden>
              {iconRight}
            </span>
          )}
        </div>
        {error ? (
          <p id={errorId} className="text-caption text-semantic-down">
            {error}
          </p>
        ) : hint ? (
          <p id={hintId} className="text-caption text-muted">
            {hint}
          </p>
        ) : null}
      </div>
    )
  },
)
