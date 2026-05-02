import type { HTMLAttributes, ReactNode } from 'react'
import { cn } from '@/shared/lib/cn'

/**
 * DESIGN.md → badge-pill:
 *   bg surface-strong, text ink, caption-strong (12/600), rounded-pill, 4×12.
 *   Optional 'inverse' tone for featured states (DESIGN: pricing-tier-featured pattern).
 */
export type BadgeTone = 'default' | 'inverse' | 'up' | 'down'

type BadgePillOwnProps = {
  tone?: BadgeTone
  children: ReactNode
}

export type BadgePillProps = BadgePillOwnProps &
  Omit<HTMLAttributes<HTMLSpanElement>, keyof BadgePillOwnProps>

const TONE_CLASSES: Record<BadgeTone, string> = {
  default: 'bg-surface-strong text-ink',
  inverse: 'bg-surface-dark text-on-dark',
  up: 'bg-surface-strong text-semantic-up',
  down: 'bg-surface-strong text-semantic-down',
}

export function BadgePill({
  tone = 'default',
  className,
  children,
  ...rest
}: BadgePillProps) {
  return (
    <span
      className={cn(
        'text-caption-strong inline-flex items-center rounded-pill px-3 py-1 uppercase tracking-wide',
        TONE_CLASSES[tone],
        className,
      )}
      {...rest}
    >
      {children}
    </span>
  )
}
