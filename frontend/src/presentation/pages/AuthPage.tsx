import { useState } from 'react'
import { Link } from 'react-router-dom'
import { useI18n } from '@/app/providers/I18nProvider'
import {
  AssetIconCircular,
  BadgePill,
  Card,
  Footer,
  LanguageSwitcher,
  PillSwitcher,
  PriceCell,
  SkipLink,
  ThemeToggle,
  TopNav,
  Wordmark,
} from '@/presentation/ui'
import { AuthForm } from '@/presentation/features/auth/AuthForm'
import type { AuthMode } from '@/presentation/features/auth/useAuthForm'

export function AuthPage() {
  const { t } = useI18n()
  const [mode, setMode] = useState<AuthMode>('login')

  const AUTH_OPTIONS = [
    { value: 'login', label: t('auth.modeLogin') },
    { value: 'register', label: t('auth.modeRegister') },
  ] as const satisfies ReadonlyArray<{ value: AuthMode; label: string }>

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

      <section id="main-content" className="flex-1">
        <div className="mx-auto grid max-w-[1200px] gap-12 px-6 py-[96px] md:grid-cols-12">
          <div className="flex flex-col justify-center md:col-span-6">
            <BadgePill>{t('auth.eyebrow')}</BadgePill>
            <h1 className="mt-6 text-[40px] leading-[1.05] tracking-[-1px] font-normal sm:text-[52px] md:text-[56px] md:tracking-[-1.4px] lg:text-[64px] lg:leading-[1] lg:tracking-[-1.6px]">
              {mode === 'login'
                ? t('auth.titleLogin')
                : t('auth.titleRegister')}
            </h1>
            <p className="text-body-md text-body mt-6 max-w-[480px]">
              {t('auth.subtitle')}
            </p>

            <div className="mt-10 hidden md:block">
              <DecorativeStack />
            </div>
          </div>

          <div className="md:col-span-6 lg:col-span-5 lg:col-start-8">
            <Card variant="feature" elevated className="w-full max-w-[460px]">
              <div className="flex flex-col gap-8">
                <PillSwitcher
                  className="self-center"
                  ariaLabel="Authentication mode"
                  options={AUTH_OPTIONS}
                  value={mode}
                  onChange={setMode}
                />
                <div>
                  <h2 className="text-title-lg">
                    {mode === 'login'
                      ? t('auth.formTitleLogin')
                      : t('auth.formTitleRegister')}
                  </h2>
                  <p className="text-body-sm text-body mt-1">
                    {mode === 'login'
                      ? t('auth.formHintLogin')
                      : t('auth.formHintRegister')}
                  </p>
                </div>
                <AuthForm mode={mode} />
              </div>
            </Card>
          </div>
        </div>
      </section>
      <Footer />
    </div>
  )
}

function DecorativeStack() {
  return (
    <div className="relative h-[200px] w-full max-w-[420px]">
      <Card
        variant="product-light"
        compact
        elevated
        className="absolute top-0 left-0 w-[260px]"
      >
        <p className="text-caption-strong text-muted uppercase">Vacancy</p>
        <p className="text-title-md mt-2">Senior Go Engineer</p>
        <div className="mt-4 flex items-center justify-between">
          <span className="text-body-sm text-body">Match score</span>
          <PriceCell tone="up" value="+91.4" />
        </div>
      </Card>
      <Card
        variant="feature"
        compact
        className="absolute right-0 bottom-0 w-[240px] rotate-[-3deg]"
      >
        <div className="flex items-center gap-3">
          <AssetIconCircular>AK</AssetIconCircular>
          <div>
            <p className="text-title-sm">Anna K.</p>
            <p className="text-caption text-muted">Backend · 6y</p>
          </div>
        </div>
        <p className="text-number-display text-semantic-up mt-3">88.2 / 100</p>
      </Card>
    </div>
  )
}
