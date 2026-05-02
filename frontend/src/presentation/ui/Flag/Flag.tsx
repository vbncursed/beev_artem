import { cn } from '@/shared/lib/cn'

/**
 * Flag glyphs for the language switcher. Geometric & flat per DESIGN.md
 * iconography rules. Rendered inside circular plates so they read at the
 * 24px nav-icon scale without high detail.
 */

const COMMON_PROPS = {
  width: 18,
  height: 18,
  xmlns: 'http://www.w3.org/2000/svg',
  'aria-hidden': true,
  focusable: false as const,
}

export function FlagRu({ className }: { className?: string }) {
  return (
    <svg
      {...COMMON_PROPS}
      viewBox="0 0 18 18"
      className={cn('overflow-hidden rounded-full', className)}
    >
      <rect width="18" height="6" fill="#ffffff" />
      <rect y="6" width="18" height="6" fill="#0039a6" />
      <rect y="12" width="18" height="6" fill="#d52b1e" />
    </svg>
  )
}

/**
 * Stars-and-Stripes, simplified for an 18px circular plate. Six white
 * stripes overlay the red field; the blue canton occupies the upper-left
 * quadrant. At this scale the 50 stars would be unreadable, so we skip
 * them — the canton + stripe rhythm is enough to read as US.
 */
export function FlagUs({ className }: { className?: string }) {
  return (
    <svg
      {...COMMON_PROPS}
      viewBox="0 0 19 10"
      className={cn('overflow-hidden rounded-full', className)}
      preserveAspectRatio="xMidYMid slice"
    >
      <rect width="19" height="10" fill="#b22234" />
      <rect y="0.77" width="19" height="0.77" fill="#ffffff" />
      <rect y="2.31" width="19" height="0.77" fill="#ffffff" />
      <rect y="3.85" width="19" height="0.77" fill="#ffffff" />
      <rect y="5.38" width="19" height="0.77" fill="#ffffff" />
      <rect y="6.92" width="19" height="0.77" fill="#ffffff" />
      <rect y="8.46" width="19" height="0.77" fill="#ffffff" />
      <rect width="7.6" height="5.38" fill="#3c3b6e" />
    </svg>
  )
}
