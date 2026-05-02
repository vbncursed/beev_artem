import type { ReactNode } from 'react'
import { cn } from '@/shared/lib/cn'

/**
 * DESIGN.md → hero-band-dark / hero-band-light.
 * Vertical padding 96px (`space-section`). Display headline left,
 * optional subhead + CTAs, optional layered product-UI mockup right.
 *
 * Hero h1 steps down responsively per DESIGN: 80 → 64 → 52 → 44 → 36px.
 */
export type HeroBandTone = 'light' | 'dark'

export type HeroBandProps = {
  tone?: HeroBandTone
  eyebrow?: ReactNode
  title: ReactNode
  subtitle?: ReactNode
  actions?: ReactNode
  visual?: ReactNode
  className?: string
}

export function HeroBand({
  tone = 'light',
  eyebrow,
  title,
  subtitle,
  actions,
  visual,
  className,
}: HeroBandProps) {
  const surface =
    tone === 'dark' ? 'bg-surface-dark text-on-dark' : 'bg-canvas text-ink'

  const subtitleColor = tone === 'dark' ? 'text-on-dark-soft' : 'text-body'

  return (
    <section className={cn('w-full', surface, className)}>
      <div className="mx-auto grid max-w-[1200px] grid-cols-1 gap-12 px-6 py-[96px] md:grid-cols-12">
        <div
          className={cn(
            'flex flex-col justify-center',
            visual ? 'md:col-span-7' : 'md:col-span-12',
          )}
        >
          {eyebrow && <div className="mb-6">{eyebrow}</div>}
          <h1 className="text-[40px] leading-[1.05] tracking-[-1px] font-normal sm:text-[52px] md:text-[64px] md:tracking-[-1.6px] lg:text-[80px] lg:leading-[1] lg:tracking-[-2px]">
            {title}
          </h1>
          {subtitle && (
            <p
              className={cn(
                'text-body-md mt-6 max-w-[560px]',
                subtitleColor,
              )}
            >
              {subtitle}
            </p>
          )}
          {actions && (
            <div className="mt-10 flex flex-wrap items-center gap-3">
              {actions}
            </div>
          )}
        </div>
        {visual && (
          <div className="relative md:col-span-5">{visual}</div>
        )}
      </div>
    </section>
  )
}
