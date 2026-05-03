import { BadgePill } from '@/presentation/ui'

/**
 * Skills group rendered inside Skills Breakdown. Tone determines the
 * BadgePill colour: matched=up (green), missing=down (red), extra=neutral.
 */
type Props = {
  label: string
  tone: 'up' | 'down' | 'neutral'
  skills: string[]
}

export function SkillCloud({ label, tone, skills }: Props) {
  return (
    <div className="flex flex-col gap-2">
      <p className="text-caption text-muted">{label}</p>
      <div className="flex flex-wrap gap-1.5">
        {skills.map((s) => (
          <BadgePill key={s} tone={tone === 'neutral' ? 'default' : tone}>
            {s}
          </BadgePill>
        ))}
      </div>
    </div>
  )
}
