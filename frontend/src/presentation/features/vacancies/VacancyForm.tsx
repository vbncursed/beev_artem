import { useI18n } from '@/app/providers/I18nProvider'
import { Button, Card, TextArea, TextInput } from '@/presentation/ui'
import { SkillsEditor } from './SkillsEditor'
import { useVacancyForm } from './useVacancyForm'

/**
 * Pure render of the new-vacancy form. State, validation, and submit
 * live in `useVacancyForm`; this file is just layout + binding.
 */
export function VacancyForm({ onCancel }: { onCancel?: () => void }) {
  const { t } = useI18n()
  const { values, errors, pending, updateField, submit } = useVacancyForm()

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
