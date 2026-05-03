import { useI18n } from '@/app/providers/I18nProvider'
import { BadgePill } from '@/presentation/ui'
import type { SkillWeight } from '@/domain/vacancy/types'

/**
 * Inline skills overview rendered in the vacancy hero. Shows total
 * count + must-have count in the eyebrow, then chips for each skill.
 * `inverse` tone marks must-haves so the visual rhythm matches the
 * weight badge in the picker UI.
 */
export function SkillsSummary({ skills }: { skills: SkillWeight[] }) {
  const { t } = useI18n()
  if (skills.length === 0) return null
  const must = skills.filter((s) => s.mustHave).length
  return (
    <div className="mt-2 flex flex-col gap-3">
      <p className="text-caption-strong text-muted uppercase">
        {must > 0
          ? t('details.skillsHeaderWithMust', { n: skills.length, m: must })
          : t('details.skillsHeader', { n: skills.length })}
      </p>
      <div className="flex flex-wrap gap-1.5">
        {skills.map((s) => (
          <BadgePill key={s.name} tone={s.mustHave ? 'inverse' : 'default'}>
            {s.name}
            <span className="ml-1 opacity-70">
              {s.weight ? s.weight.toFixed(2) : ''}
            </span>
          </BadgePill>
        ))}
      </div>
    </div>
  )
}
