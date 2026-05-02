import type { ReactNode } from 'react'

/**
 * DESIGN.md → cta-band-dark.
 * Pre-footer band: surface-dark, on-dark text, 96px vertical padding,
 * centered display headline + two CTAs.
 */
export function CtaBandDark({
  title,
  subtitle,
  actions,
}: {
  title: ReactNode
  subtitle?: ReactNode
  actions?: ReactNode
}) {
  return (
    <section className="w-full bg-surface-dark text-on-dark">
      <div className="mx-auto flex max-w-[1200px] flex-col items-center px-6 py-[96px] text-center">
        <h2 className="text-display-lg max-w-[760px]">{title}</h2>
        {subtitle && (
          <p className="text-body-md text-on-dark-soft mt-6 max-w-[560px]">
            {subtitle}
          </p>
        )}
        {actions && (
          <div className="mt-10 flex flex-wrap items-center justify-center gap-3">
            {actions}
          </div>
        )}
      </div>
    </section>
  )
}
