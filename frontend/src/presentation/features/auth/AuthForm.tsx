import { useI18n } from '@/app/providers/I18nProvider'
import { Button, TextInput } from '@/presentation/ui'
import { useAuthForm, type AuthMode } from './useAuthForm'

export function AuthForm({ mode }: { mode: AuthMode }) {
  const { t } = useI18n()
  const { values, errors, pending, update, submit } = useAuthForm(mode)

  return (
    <form onSubmit={submit} noValidate className="flex flex-col gap-5">
      <TextInput
        label={t('common.email')}
        type="email"
        autoComplete="email"
        value={values.email}
        onChange={(e) => update('email')(e.target.value)}
        error={errors.email}
        autoFocus
        required
      />
      <TextInput
        label={t('common.password')}
        type="password"
        autoComplete={mode === 'login' ? 'current-password' : 'new-password'}
        value={values.password}
        onChange={(e) => update('password')(e.target.value)}
        error={errors.password}
        hint={mode === 'register' ? t('auth.passwordHintRegister') : undefined}
        required
      />
      {mode === 'register' && (
        <TextInput
          label={t('auth.confirmPassword')}
          type="password"
          autoComplete="new-password"
          value={values.confirmPassword}
          onChange={(e) => update('confirmPassword')(e.target.value)}
          error={errors.confirmPassword}
          required
        />
      )}

      {errors._form && (
        <p role="alert" className="text-body-sm text-semantic-down -mt-1">
          {errors._form}
        </p>
      )}

      <Button type="submit" variant="primary-cta" fullWidth loading={pending}>
        {mode === 'login' ? t('auth.submitLogin') : t('auth.submitRegister')}
      </Button>
    </form>
  )
}
