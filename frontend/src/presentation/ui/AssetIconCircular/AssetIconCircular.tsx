import type { ReactNode } from 'react'
import { cn } from '@/shared/lib/cn'

/**
 * DESIGN.md → asset-icon-circular: surface-strong, rounded-full, 32px.
 * In beev we reuse this for candidate/avatar plates next to asset rows.
 */
export function AssetIconCircular({
  children,
  size = 32,
  className,
}: {
  children: ReactNode
  size?: number
  className?: string
}) {
  return (
    <span
      className={cn(
        'inline-flex shrink-0 items-center justify-center rounded-full bg-surface-strong text-ink',
        className,
      )}
      style={{ width: size, height: size }}
    >
      <span className="text-caption-strong">{children}</span>
    </span>
  )
}
