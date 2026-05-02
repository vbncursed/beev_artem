import { useState, type FormEvent } from 'react'
import { useNavigate } from 'react-router-dom'
import { useGateways } from '@/app/providers/GatewaysProvider'
import { useI18n } from '@/app/providers/I18nProvider'
import { Button, Card, TextArea, TextInput } from '@/presentation/ui'
import {
  vacancyFormSchema,
  type VacancyFormInput,
} from '@/domain/vacancy/rules'
import type { CreateVacancyInput, SkillWeight } from '@/domain/vacancy/types'
import { ApiError } from '@/infrastructure/http/errors'
import { SkillsEditor, type SkillRowError } from './SkillsEditor'

type FieldErrors = {
  title?: string
  description?: string
  skills?: string
  skillRows?: SkillRowError[]
  _form?: string
}

const INITIAL_VALUES: VacancyFormInput = {
  title: '',
  description: '',
  skills: [{ name: '', weight: 0 }],
}

const ZOD_TO_KEY: Record<string, string> = {
  'Title is required': 'form.error.titleRequired',
  'Up to 255 characters': 'form.error.titleMax',
  'Up to 4000 characters': 'form.error.descriptionMax',
  'Add at least one skill': 'form.error.skillsMin',
  'Skill name is required': 'form.error.skillNameRequired',
  'Up to 64 characters': 'form.error.skillNameMax',
  'Number from 0 to 1': 'form.error.weightRange',
  'Min 0': 'form.error.weightMin',
  'Max 1': 'form.error.weightMax',
}

export function VacancyForm({ onCancel }: { onCancel?: () => void }) {
  const { t } = useI18n()
  const { vacancy: gateway } = useGateways()
  const navigate = useNavigate()
  const [values, setValues] = useState<VacancyFormInput>(INITIAL_VALUES)
  const [errors, setErrors] = useState<FieldErrors>({})
  const [pending, setPending] = useState(false)

  const updateField = <K extends keyof VacancyFormInput>(
    field: K,
    value: VacancyFormInput[K],
  ) => {
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

    const parsed = vacancyFormSchema.safeParse(values)
    if (!parsed.success) {
      const next: FieldErrors = {}
      const skillRows: SkillRowError[] = []
      for (const issue of parsed.error.issues) {
        const [head, idx, leaf] = issue.path
        const message = translate(issue.message, t)
        if (head === 'skills' && typeof idx === 'number') {
          const row = (skillRows[idx] ??= {})
          if (leaf === 'name' && !row.name) row.name = message
          if (leaf === 'weight' && !row.weight) row.weight = message
        } else if (head === 'skills') {
          next.skills ??= message
        } else if (typeof head === 'string') {
          const k = head as keyof FieldErrors
          if (k !== 'skillRows' && !next[k]) {
            next[k] = message as never
          }
        }
      }
      if (skillRows.length) next.skillRows = skillRows
      setErrors(next)
      return
    }

    setPending(true)
    setErrors({})
    try {
      const payload: CreateVacancyInput = {
        title: parsed.data.title,
        description: parsed.data.description ?? '',
        skills: parsed.data.skills.map(toSkillWeight),
      }
      const created = await gateway.create(payload)
      navigate(`/vacancies/${created.id}`)
    } catch (cause) {
      setErrors({ _form: messageFor(cause, t) })
      setPending(false)
    }
  }

  return (
    <Card variant="feature" elevated>
      <form onSubmit={submit} noValidate className="flex flex-col gap-6">
        <TextInput
          label={t('form.titleLabel')}
          placeholder={t('form.titlePlaceholder')}
          value={values.title}
          onChange={(e) => updateField('title', e.target.value)}
          error={errors.title}
          maxLength={255}
          autoFocus
          required
        />

        <TextArea
          label={t('form.descriptionLabel')}
          placeholder={t('form.descriptionPlaceholder')}
          value={values.description ?? ''}
          onChange={(e) => updateField('description', e.target.value)}
          error={errors.description}
          hint={t('form.descriptionCounter', {
            count: (values.description ?? '').length,
          })}
          maxLength={4000}
          rows={6}
        />

        <SkillsEditor
          value={values.skills}
          errors={errors.skillRows}
          onChange={(next) => updateField('skills', next)}
          disabled={pending}
        />
        {errors.skills && (
          <p className="text-caption text-semantic-down -mt-3">
            {errors.skills}
          </p>
        )}

        {errors._form && (
          <p
            role="alert"
            className="text-body-sm text-semantic-down rounded-md bg-surface-soft px-4 py-3"
          >
            {errors._form}
          </p>
        )}

        <div className="flex flex-wrap items-center gap-3">
          <Button type="submit" variant="primary-cta" loading={pending}>
            {t('form.submit')}
          </Button>
          {onCancel && (
            <Button
              type="button"
              variant="secondary-light"
              onClick={onCancel}
              disabled={pending}
            >
              {t('common.cancel')}
            </Button>
          )}
        </div>
      </form>
    </Card>
  )
}

function translate(zod: string, t: (k: string) => string): string {
  const key = ZOD_TO_KEY[zod]
  return key ? t(key) : zod
}

function toSkillWeight(s: {
  name: string
  weight: number
  mustHave?: boolean
  niceToHave?: boolean
}): SkillWeight {
  return {
    name: s.name.trim(),
    weight: s.weight,
    mustHave: s.mustHave,
    niceToHave: s.niceToHave,
  }
}

function messageFor(cause: unknown, t: (k: string) => string): string {
  if (cause instanceof ApiError) {
    switch (cause.reason) {
      case 'INVALID_ARGUMENT':
        return cause.message || t('form.error.invalidArg')
      case 'UNAUTHORIZED':
        return t('form.error.unauthorized')
      case 'RATE_LIMITED':
        return t('form.error.rateLimited')
      case 'NETWORK':
        return t('form.error.network')
      default:
        return cause.message || t('form.error.fallback')
    }
  }
  return t('form.error.fallback')
}
