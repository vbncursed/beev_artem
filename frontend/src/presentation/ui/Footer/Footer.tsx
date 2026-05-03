import { Link } from 'react-router-dom'
import { useI18n } from '@/app/providers/I18nProvider'
import { BrandMark } from '@/presentation/ui/BrandMark'
import { cn } from '@/shared/lib/cn'

/**
 * DESIGN.md → footer-light: canvas background, body text, body-sm
 * (14/400). We sit it on top of a hairline border so the footer reads
 * as a closing strip without introducing a second elevation tier.
 *
 * Layout adapts: stacked on mobile, three-column flex at md+.
 */
export function Footer({ className }: { className?: string }) {
  const { t } = useI18n()
  const year = new Date().getFullYear()

  return (
    <footer
      className={cn(
        'border-t border-hairline bg-canvas',
        className,
      )}
    >
      {/* Three-column grid keeps the nav column anchored to the page
          center regardless of how wide brand or copyright sections grow.
          Mobile collapses to a vertical stack. */}
      <div className="mx-auto flex max-w-[1200px] flex-col gap-6 px-6 py-10 md:grid md:grid-cols-3 md:items-center md:gap-4">
        <div className="flex items-center gap-3 md:justify-self-start">
          <BrandMark size={20} />
          <p className="text-caption text-muted">{t('footer.tagline')}</p>
        </div>

        <nav
          aria-label={t('footer.navAria')}
          className="text-body-sm text-body flex flex-wrap gap-x-6 gap-y-2 md:justify-self-center"
        >
          <FooterLink to="/privacy">{t('footer.privacy')}</FooterLink>
          <FooterLink to="/terms">{t('footer.terms')}</FooterLink>
          <FooterLink to="/support">{t('footer.help')}</FooterLink>
        </nav>

        <p className="text-caption text-muted md:justify-self-end">
          {t('footer.copyright', { year })}
        </p>
      </div>
    </footer>
  )
}

function FooterLink({
  to,
  children,
}: {
  to: string
  children: React.ReactNode
}) {
  return (
    <Link
      to={to}
      className={cn(
        'cursor-pointer rounded-sm transition-colors',
        'hover:text-ink',
        'focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary',
      )}
    >
      {children}
    </Link>
  )
}
