import { useId } from 'react'
import { useI18n } from '@/app/providers/I18nProvider'
import {
  Button,
  Checkbox,
  TextInput,
  WeightStepper,
} from '@/presentation/ui'
import { cn } from '@/shared/lib/cn'
import type { SkillWeight } from '@/domain/vacancy/types'

export type SkillRow = SkillWeight

export type SkillRowError = {
  name?: string
  weight?: string
}

type Props = {
  value: SkillRow[]
  errors?: SkillRowError[]
  onChange: (next: SkillRow[]) => void
  disabled?: boolean
}

const EMPTY_ROW: SkillRow = { name: '', weight: 0 }

export function SkillsEditor({ value, errors = [], onChange, disabled }: Props) {
  const { t } = useI18n()
  const headerId = useId()

  const update = (index: number, patch: Partial<SkillRow>) => {
    const next = value.map((row, i) =>
      i === index ? { ...row, ...patch } : row,
    )
    onChange(next)
  }

  const remove = (index: number) => {
    onChange(value.filter((_, i) => i !== index))
  }

  const append = () => {
    onChange([...value, { ...EMPTY_ROW }])
  }

  const allZero = value.length > 0 && value.every((r) => r.weight === 0)

  return (
    <fieldset
      aria-labelledby={headerId}
      className="flex flex-col gap-3"
      disabled={disabled}
    >
      <div className="flex items-center justify-between">
        <legend
          id={headerId}
          className="text-caption-strong text-body-strong"
        >
          {t('skills.legend')}
        </legend>
        <span className="text-caption text-muted">{t('skills.hint')}</span>
      </div>

      {value.length === 0 && (
        <p className="text-body-sm text-muted">{t('skills.empty')}</p>
      )}

      <ul className="flex flex-col gap-3">
        {value.map((row, i) => {
          const err = errors[i]
          return (
            <li
              key={i}
              className="grid grid-cols-12 items-start gap-3 rounded-lg bg-surface-soft p-3"
            >
              <TextInput
                className="col-span-12 md:col-span-5"
                placeholder={t('skills.namePlaceholder')}
                value={row.name}
                onChange={(e) => update(i, { name: e.target.value })}
                error={err?.name}
                aria-label={t('skills.skillNameAria', { index: i + 1 })}
              />

              <WeightStepper
                className="col-span-6 md:col-span-3"
                value={row.weight}
                onChange={(weight) => update(i, { weight })}
                error={err?.weight}
                ariaLabel={t('skills.weightAria', { index: i + 1 })}
                disabled={disabled}
              />

              <div className="col-span-6 flex flex-col gap-2 self-center md:col-span-3">
                <Checkbox
                  label={t('skills.must')}
                  checked={Boolean(row.mustHave)}
                  onChange={(checked) =>
                    update(i, {
                      mustHave: checked,
                      // Must and Nice are mutually exclusive — flipping one
                      // clears the other so the data shape stays sane.
                      niceToHave: checked ? false : row.niceToHave,
                    })
                  }
                />
                <Checkbox
                  label={t('skills.nice')}
                  checked={Boolean(row.niceToHave)}
                  onChange={(checked) =>
                    update(i, {
                      niceToHave: checked,
                      mustHave: checked ? false : row.mustHave,
                    })
                  }
                />
              </div>

              <button
                type="button"
                onClick={() => remove(i)}
                aria-label={t('skills.removeAria', { index: i + 1 })}
                title={t('skills.remove')}
                className={cn(
                  'col-span-12 inline-flex size-9 cursor-pointer items-center justify-center self-center rounded-full md:col-span-1 md:justify-self-end',
                  'text-muted transition-colors hover:bg-surface-strong hover:text-semantic-down',
                  'focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary',
                )}
              >
                <TrashIcon />
              </button>
            </li>
          )
        })}
      </ul>

      <div className="flex items-center justify-between">
        <Button variant="text" onClick={append} type="button">
          {t('skills.add')}
        </Button>
        {allZero && (
          <span className="text-caption text-muted">{t('skills.allZero')}</span>
        )}
      </div>
    </fieldset>
  )
}

function TrashIcon() {
  return (
    <svg
      width="16"
      height="16"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
      aria-hidden
    >
      <path d="M3 6h18" />
      <path d="M8 6V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2" />
      <path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6" />
      <path d="M10 11v6" />
      <path d="M14 11v6" />
    </svg>
  )
}
