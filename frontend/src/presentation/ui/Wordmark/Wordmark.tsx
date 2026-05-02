import { BrandMark } from '@/presentation/ui/BrandMark'
import { cn } from '@/shared/lib/cn'

/**
 * Brand lockup: BrandMark + "Cadence" wordtype. The mark provides the
 * single Coinbase Blue moment per DESIGN.md; the wordtype stays ink/on-dark
 * so the lockup reads in both themes.
 */
export function Wordmark({
  tone = 'light',
  className,
}: {
  tone?: 'light' | 'dark'
  className?: string
}) {
  return (
    <span
      className={cn(
        'inline-flex items-center gap-2 font-display text-[20px] leading-none font-semibold tracking-[-0.02em]',
        tone === 'dark' ? 'text-on-dark' : 'text-ink',
        className,
      )}
    >
      <BrandMark size={22} />
      Cadence
    </span>
  )
}
