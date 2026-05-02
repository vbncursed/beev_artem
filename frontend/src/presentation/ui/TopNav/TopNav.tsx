import type { ReactNode } from 'react'
import { cn } from '@/shared/lib/cn'

/**
 * DESIGN.md → top-nav-light / top-nav-on-dark.
 * Height 64px, layout: wordmark left, primary horizontal menu (slot),
 * actions right (search + theme + auth CTAs).
 *
 * Pure layout component — provides surface + spacing, never renders
 * specific menu items.
 */
export type TopNavTone = 'light' | 'dark'

export type TopNavProps = {
  tone?: TopNavTone
  brand: ReactNode
  links?: ReactNode
  actions?: ReactNode
  className?: string
}

export function TopNav({
  tone = 'light',
  brand,
  links,
  actions,
  className,
}: TopNavProps) {
  const surface =
    tone === 'dark'
      ? 'bg-surface-dark text-on-dark'
      : 'bg-canvas text-ink border-b border-hairline'

  return (
    <header
      className={cn(
        'sticky top-0 z-30 flex h-16 w-full items-center',
        surface,
        className,
      )}
    >
      <div className="mx-auto flex w-full max-w-[1200px] items-center gap-8 px-6">
        <div className="flex items-center gap-8">{brand}</div>
        {links && (
          <nav
            aria-label="Primary"
            className="text-nav-link hidden flex-1 items-center gap-6 lg:flex"
          >
            {links}
          </nav>
        )}
        <div className="ml-auto flex items-center gap-3">{actions}</div>
      </div>
    </header>
  )
}

export function TopNavLink({
  active,
  children,
  ...rest
}: {
  active?: boolean
  children: ReactNode
} & React.AnchorHTMLAttributes<HTMLAnchorElement>) {
  return (
    <a
      {...rest}
      className={cn(
        'text-nav-link cursor-pointer transition-colors',
        active ? 'text-ink' : 'text-body hover:text-ink',
        rest.className,
      )}
    >
      {children}
    </a>
  )
}
