import type { HTMLAttributes, ReactNode } from 'react'
import { cn } from '@/shared/lib/cn'

/**
 * DESIGN.md cards mapping:
 *   feature        → feature-card        (canvas, hairline border, rounded-xl, 32px)
 *   product-light  → product-ui-card-light
 *   product-dark   → product-ui-card-dark (surface-dark-elevated, on-dark)
 *   pricing        → pricing-tier-card   (canvas + hairline)
 *   pricing-featured → pricing-tier-featured (surface-dark, on-dark inversion)
 *
 * 80% surfaces are flat. The single shadow tier (`shadow-soft`) is opt-in
 * via the `elevated` prop and reserved for hover affordances.
 */
export type CardVariant =
  | 'feature'
  | 'product-light'
  | 'product-dark'
  | 'pricing'
  | 'pricing-featured'

type CardOwnProps = {
  variant?: CardVariant
  /** Apply the single soft shadow tier (DESIGN.md elevation §). */
  elevated?: boolean
  /** Tighten padding to 24px for compact rows. Default 32px per DESIGN. */
  compact?: boolean
  children: ReactNode
}

export type CardProps = CardOwnProps &
  Omit<HTMLAttributes<HTMLDivElement>, keyof CardOwnProps>

const VARIANT_CLASSES: Record<CardVariant, string> = {
  feature: 'bg-canvas text-ink border border-hairline',
  'product-light': 'bg-canvas text-ink border border-hairline',
  'product-dark': 'bg-surface-dark-elevated text-on-dark',
  pricing: 'bg-canvas text-ink border border-hairline',
  'pricing-featured': 'bg-surface-dark text-on-dark',
}

export function Card({
  variant = 'feature',
  elevated = false,
  compact = false,
  className,
  children,
  ...rest
}: CardProps) {
  return (
    <div
      className={cn(
        'rounded-xl',
        compact ? 'p-6' : 'p-8',
        VARIANT_CLASSES[variant],
        elevated && 'shadow-soft',
        className,
      )}
      {...rest}
    >
      {children}
    </div>
  )
}
