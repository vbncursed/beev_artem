import type { ReactNode } from 'react'
import { Link } from 'react-router-dom'
import { useI18n } from '@/app/providers/I18nProvider'
import {
  BadgePill,
  Footer,
  LanguageSwitcher,
  SkipLink,
  ThemeToggle,
  TopNav,
  Wordmark,
} from '@/presentation/ui'

/**
 * Shared chrome for the privacy / terms / support trio. Public — the
 * three pages are reachable without authentication so they can be linked
 * from emails or external sites without forcing a login wall.
 *
 * Layout follows DESIGN.md page rhythm: light hero band with eyebrow +
 * display-md title, then a single white content section capped at 760px
 * for editorial reading width.
 */
type Props = {
  eyebrow: string
  title: string
  subtitle?: string
  children: ReactNode
}

export function LegalLayout({ eyebrow, title, subtitle, children }: Props) {
  const { t } = useI18n()
  return (
    <div className="flex min-h-full flex-col bg-canvas text-ink">
      <SkipLink />
      <TopNav
        brand={
          <Link to="/" className="cursor-pointer">
            <Wordmark />
          </Link>
        }
        actions={
          <>
            <LanguageSwitcher />
            <ThemeToggle />
          </>
        }
      />

      <main id="main-content" className="flex-1">
        <section className="bg-canvas">
          <div className="mx-auto max-w-[1200px] px-6 pt-[96px] pb-12">
            <BadgePill>{eyebrow}</BadgePill>
            <h1 className="text-display-md mt-4">{title}</h1>
            {subtitle && (
              <p className="text-body-md text-body mt-3 max-w-[640px]">
                {subtitle}
              </p>
            )}
            <p className="text-caption text-muted mt-6">
              {t('legal.lastUpdated', { date: '2026-05-03' })}
            </p>
          </div>
        </section>

        <section className="bg-surface-soft">
          <div className="prose-cadence mx-auto max-w-[760px] px-6 py-12 pb-[96px]">
            {children}
          </div>
        </section>
      </main>

      <Footer />
    </div>
  )
}

/**
 * Tiny markup primitives used by every legal page so heading levels +
 * spacing stay consistent without dragging in @tailwindcss/typography.
 */
export function H2({ children }: { children: ReactNode }) {
  return (
    <h2 className="text-title-lg text-ink mt-12 mb-4 first:mt-0">{children}</h2>
  )
}

export function H3({ children }: { children: ReactNode }) {
  return (
    <h3 className="text-title-md text-ink mt-8 mb-3">{children}</h3>
  )
}

export function P({ children }: { children: ReactNode }) {
  return (
    <p className="text-body-md text-body mt-3 first:mt-0 break-words">
      {children}
    </p>
  )
}

export function UL({ children }: { children: ReactNode }) {
  return (
    <ul className="text-body-md text-body mt-3 list-disc space-y-1.5 pl-6">
      {children}
    </ul>
  )
}
