import { useState, type FormEvent } from 'react'
import { useAuth } from '@/app/providers/AuthProvider'
import { useI18n } from '@/app/providers/I18nProvider'
import { ApiError } from '@/infrastructure/http/errors'
import {
  credentialsSchema,
  registerSchema,
  type RegisterInput,
} from '@/domain/auth/rules'

export type AuthMode = 'login' | 'register'

export type FieldErrors = Partial<
  Record<'email' | 'password' | 'confirmPassword' | '_form', string>
>

type FormShape = {
  email: string
  password: string
  confirmPassword: string
}

const ZOD_TO_KEY: Record<string, string> = {
  'Email is required': 'auth.error.emailRequired',
  'Invalid email format': 'auth.error.emailInvalid',
  'At least 8 characters': 'auth.error.passwordShort',
  'Maximum 72 characters': 'auth.error.passwordLong',
  'Passwords do not match': 'auth.error.passwordsNoMatch',
}

export function useAuthForm(mode: AuthMode) {
  const { login, register } = useAuth()
  const { t } = useI18n()
  const [values, setValues] = useState<FormShape>({
    email: '',
    password: '',
    confirmPassword: '',
  })
  const [errors, setErrors] = useState<FieldErrors>({})
  const [pending, setPending] = useState(false)

  const update = (field: keyof FormShape) => (value: string) => {
    setValues((v) => ({ ...v, [field]: value }))
    if (errors[field as keyof FieldErrors] || errors._form) {
      setErrors((e) => {
        const next = { ...e }
        delete next[field as keyof FieldErrors]
        delete next._form
        return next
      })
    }
  }

  const submit = async (e: FormEvent<HTMLFormElement>) => {
    e.preventDefault()
    if (pending) return

    const parsed =
      mode === 'login'
        ? credentialsSchema.safeParse({
            email: values.email,
            password: values.password,
          })
        : registerSchema.safeParse(values satisfies RegisterInput)

    if (!parsed.success) {
      const next: FieldErrors = {}
      for (const issue of parsed.error.issues) {
        const k = issue.path[0]
        if (typeof k === 'string' && !next[k as keyof FieldErrors]) {
          next[k as keyof FieldErrors] = translate(issue.message, t)
        }
      }
      setErrors(next)
      return
    }

    setPending(true)
    setErrors({})
    try {
      if (mode === 'login') {
        await login({ email: values.email, password: values.password })
      } else {
        await register({ email: values.email, password: values.password })
      }
    } catch (cause) {
      setErrors({ _form: messageFor(cause, mode, t) })
    } finally {
      setPending(false)
    }
  }

  return { values, errors, pending, update, submit }
}

function translate(zodMessage: string, t: (k: string) => string): string {
  const key = ZOD_TO_KEY[zodMessage]
  return key ? t(key) : zodMessage
}

function messageFor(
  cause: unknown,
  mode: AuthMode,
  t: (k: string) => string,
): string {
  if (cause instanceof ApiError) {
    switch (cause.reason) {
      case 'UNAUTHORIZED':
        return t('auth.error.invalidCreds')
      case 'CONFLICT':
        return t('auth.error.exists')
      case 'INVALID_ARGUMENT':
        return cause.message || t('auth.error.invalidArg')
      case 'RATE_LIMITED':
        return t('auth.error.rateLimited')
      case 'NETWORK':
        return t('auth.error.network')
      default:
        return (
          cause.message ||
          t(
            mode === 'login'
              ? 'auth.error.fallbackLogin'
              : 'auth.error.fallbackRegister',
          )
        )
    }
  }
  return t(
    mode === 'login'
      ? 'auth.error.fallbackLogin'
      : 'auth.error.fallbackRegister',
  )
}
