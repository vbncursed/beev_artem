import { Link, NavLink, Outlet, useLocation } from 'react-router-dom'
import { useAuth } from '@/app/providers/AuthProvider'
import { useI18n } from '@/app/providers/I18nProvider'
import {
  Button,
  Footer,
  LanguageSwitcher,
  SkipLink,
  ThemeToggle,
  TopNav,
  TopNavLink,
  Wordmark,
} from '@/presentation/ui'

const NAV_ITEMS = [{ to: '/vacancies', labelKey: 'nav.vacancies' }] as const

/**
 * Authenticated app shell. SkipLink → TopNav → main → Footer.
 * `id="main-content"` on <main> is the SkipLink target.
 */
export function AppLayout() {
  const { user, logout } = useAuth()
  const { t } = useI18n()
  const location = useLocation()

  return (
    <div className="flex min-h-full flex-col bg-canvas text-ink">
      <SkipLink />
      <TopNav
        brand={
          <Link to="/vacancies" className="cursor-pointer">
            <Wordmark />
          </Link>
        }
        links={NAV_ITEMS.map((item) => (
          <NavLink
            key={item.to}
            to={item.to}
            className={({ isActive }) =>
              `text-nav-link cursor-pointer transition-colors ${
                isActive || location.pathname.startsWith(item.to)
                  ? 'text-ink'
                  : 'text-body hover:text-ink'
              }`
            }
          >
            {t(item.labelKey)}
          </NavLink>
        ))}
        actions={
          <>
            <LanguageSwitcher />
            <ThemeToggle />
            {user && (
              <span className="text-caption text-muted hidden md:inline">
                {user.email}
              </span>
            )}
            <Button variant="secondary-light" onClick={() => void logout()}>
              {t('common.signOut')}
            </Button>
          </>
        }
      />
      <main id="main-content" className="flex-1">
        <Outlet />
      </main>
      <Footer />
    </div>
  )
}

export { TopNavLink }
