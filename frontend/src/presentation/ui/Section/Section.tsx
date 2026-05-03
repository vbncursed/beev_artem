import type { ReactNode } from 'react'
import { cn } from '@/shared/lib/cn'

/**
 * Editorial content section: a small uppercase label on top, then content.
 * Stacks vertically; consecutive sections in the same parent get a
 * hairline divider between them via the `border-t` rule that's null on
 * the first child.
 *
 * Used by AnalysisDetails to break the long card into HR-recommendation /
 * skills-breakdown / candidate-profile / feedback sub-sections without
 * dragging in @tailwindcss/typography.
 */
type Props = {
  title: ReactNode
  children: ReactNode
  className?: string
}

export function Section({ title, children, className }: Props) {
  return (
    <section
      className={cn(
        'flex flex-col gap-2 border-t border-hairline pt-4 first:border-t-0 first:pt-0',
        className,
      )}
    >
      <p className="text-caption-strong text-muted uppercase">{title}</p>
      {children}
    </section>
  )
}
