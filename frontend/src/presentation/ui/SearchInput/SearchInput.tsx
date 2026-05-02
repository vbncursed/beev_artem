import { forwardRef, type InputHTMLAttributes } from 'react'
import { cn } from '@/shared/lib/cn'

/**
 * DESIGN.md → search-input-pill:
 *   bg surface-strong, rounded-pill, padding 12×20, height 44px.
 */
export type SearchInputProps = InputHTMLAttributes<HTMLInputElement>

const SearchIcon = () => (
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
    <circle cx="11" cy="11" r="7" />
    <path d="m21 21-4.3-4.3" />
  </svg>
)

export const SearchInput = forwardRef<HTMLInputElement, SearchInputProps>(
  function SearchInput({ className, type = 'search', ...rest }, ref) {
    return (
      <label
        className={cn(
          'flex h-11 items-center gap-2 rounded-pill bg-surface-strong px-5 text-ink',
          'border border-hairline transition-colors',
          'focus-within:border-primary',
          className,
        )}
      >
        <span className="text-muted">
          <SearchIcon />
        </span>
        <input
          ref={ref}
          type={type}
          className="text-body-md h-full w-full bg-transparent text-ink placeholder:text-muted focus:outline-none"
          {...rest}
        />
      </label>
    )
  },
)
