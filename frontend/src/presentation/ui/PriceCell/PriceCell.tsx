import { cn } from '@/shared/lib/cn'

/**
 * DESIGN.md → price-up-cell / price-down-cell.
 * Color only — green or red text in number-display (mono). Never a fill.
 */
export type PriceTone = 'up' | 'down' | 'neutral'

export function PriceCell({
  value,
  tone = 'neutral',
  className,
}: {
  value: string | number
  tone?: PriceTone
  className?: string
}) {
  const color =
    tone === 'up'
      ? 'text-semantic-up'
      : tone === 'down'
        ? 'text-semantic-down'
        : 'text-ink'
  return (
    <span className={cn('text-number-display tabular-nums', color, className)}>
      {value}
    </span>
  )
}
