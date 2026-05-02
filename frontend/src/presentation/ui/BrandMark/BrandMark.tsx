import { cn } from '@/shared/lib/cn'

/**
 * Cadence brand mark. Three ascending bars inside a rounded square —
 * geometric and minimal per DESIGN.md, primary blue is the only voltage.
 *
 * Same shape lives in /public/favicon.svg so the tab icon and the
 * in-app wordmark stay in lockstep.
 */
export function BrandMark({
  size = 24,
  className,
}: {
  size?: number
  className?: string
}) {
  return (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      viewBox="0 0 32 32"
      width={size}
      height={size}
      className={cn('shrink-0', className)}
      aria-hidden
      focusable="false"
    >
      <rect width="32" height="32" rx="8" fill="var(--color-primary)" />
      <rect x="7" y="18" width="4" height="8" rx="2" fill="#ffffff" />
      <rect x="14" y="12" width="4" height="14" rx="2" fill="#ffffff" />
      <rect x="21" y="6" width="4" height="20" rx="2" fill="#ffffff" />
    </svg>
  )
}
