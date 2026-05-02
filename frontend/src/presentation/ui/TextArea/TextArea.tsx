import {
  forwardRef,
  useId,
  type TextareaHTMLAttributes,
} from 'react'
import { cn } from '@/shared/lib/cn'

/**
 * Multi-line companion to TextInput. Same focus geometry: border hairline
 * → primary on focus, error renders semantic-down text below the field.
 */
type TextAreaOwnProps = {
  label?: string
  hint?: string
  error?: string
}

export type TextAreaProps = TextAreaOwnProps &
  Omit<TextareaHTMLAttributes<HTMLTextAreaElement>, keyof TextAreaOwnProps>

export const TextArea = forwardRef<HTMLTextAreaElement, TextAreaProps>(
  function TextArea(
    { label, hint, error, className, id, rows = 5, ...rest },
    ref,
  ) {
    const reactId = useId()
    const fieldId = id ?? reactId
    const hintId = `${fieldId}-hint`
    const errorId = `${fieldId}-error`
    const invalid = Boolean(error)

    return (
      <div className={cn('flex flex-col gap-1.5', className)}>
        {label && (
          <label
            htmlFor={fieldId}
            className="text-caption-strong text-body-strong"
          >
            {label}
          </label>
        )}
        <div
          className={cn(
            'flex rounded-md bg-canvas',
            'border border-hairline transition-colors',
            'focus-within:border-primary focus-within:shadow-[inset_0_0_0_1px_var(--color-primary)]',
            invalid &&
              'border-semantic-down focus-within:border-semantic-down focus-within:shadow-[inset_0_0_0_1px_var(--color-semantic-down)]',
            rest.disabled && 'opacity-60',
          )}
        >
          <textarea
            ref={ref}
            id={fieldId}
            rows={rows}
            aria-invalid={invalid || undefined}
            aria-describedby={
              error ? errorId : hint ? hintId : undefined
            }
            className="text-body-md min-h-[120px] w-full resize-y bg-transparent px-4 py-3 text-ink placeholder:text-muted focus:outline-none"
            {...rest}
          />
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
