import { useI18n } from '@/app/providers/I18nProvider'
import { Card, ErrorCard, PriceCell, Section, Spinner } from '@/presentation/ui'
import type { Analysis } from '@/domain/analysis/types'
import { pluralKey } from '@/shared/i18n/dictionaries'
import { formatYears } from '@/shared/lib/format'
import { AnalysisActionBar } from './AnalysisActionBar'
import { CandidateStatusBadge } from './CandidateStatusBadge'
import { ProfileLine } from './ProfileLine'
import { RecommendationBadge } from './RecommendationBadge'
import { SkillCloud } from './SkillCloud'
import { useAnalysisFetch } from './useAnalysisFetch'

/**
 * Right-hand panel on `/vacancies/:id`. Pure layout: data fetching is
 * delegated to `useAnalysisFetch`, the download/delete actions to
 * `AnalysisActionBar`, and each rendered sub-section to its own component.
 * Adding a new section = a new <Section> block, nothing else.
 */
export function AnalysisDetails({
  analysisId,
  onDeleted,
}: {
  analysisId: string
  onDeleted?: () => void
}) {
  const { t, locale } = useI18n()
  const state = useAnalysisFetch(analysisId)

  if (state.phase === 'loading') {
    return (
      <Card variant="feature" className="flex items-center justify-center py-16">
        <Spinner size={20} />
      </Card>
    )
  }
  if (state.phase === 'error') {
    return <ErrorCard message={state.message} />
  }

  const a = state.analysis
  const score = a.matchScore

  return (
    <Card variant="feature" className="flex flex-col gap-6">
      <header className="flex items-start justify-between gap-3">
        <div>
          <p className="text-caption-strong text-muted uppercase">
            {t('analysis.matchScore')}
          </p>
          <p className="mt-1">
            <PriceCell
              tone={score >= 70 ? 'up' : score < 40 ? 'down' : 'neutral'}
              value={score.toFixed(1)}
              className="text-[44px] leading-none"
            />
          </p>
        </div>
        <div className="flex flex-col items-end gap-2">
          <CandidateStatusBadge status={a.status} />
          <AnalysisActionBar
            resumeId={a.resumeId}
            candidateId={a.candidateId}
            onDeleted={onDeleted}
          />
        </div>
      </header>

      {a.status === 'failed' && a.errorMessage && (
        <p
          role="alert"
          className="text-body-sm text-semantic-down rounded-md bg-surface-soft px-4 py-3"
        >
          {a.errorMessage}
        </p>
      )}

      {a.ai?.hrRecommendation && (
        <RecommendationSection ai={a.ai} />
      )}

      {a.breakdown && (
        <BreakdownSection breakdown={a.breakdown} />
      )}

      {a.profile && (
        <ProfileSection profile={a.profile} />
      )}

      {a.ai?.candidateFeedback && (
        <Section title={t('analysis.feedback')}>
          <p className="text-body-md text-body break-words whitespace-pre-line">
            {a.ai.candidateFeedback}
          </p>
        </Section>
      )}
    </Card>
  )

  function RecommendationSection({ ai }: { ai: NonNullable<Analysis['ai']> }) {
    return (
      <Section title={t('analysis.recommendation')}>
        <RecommendationBadge value={ai.hrRecommendation} />
        {ai.hrRationale && (
          <p className="text-body-md text-body mt-3 break-words whitespace-pre-line">
            {ai.hrRationale}
          </p>
        )}
      </Section>
    )
  }

  function BreakdownSection({
    breakdown,
  }: {
    breakdown: NonNullable<Analysis['breakdown']>
  }) {
    return (
      <Section title={t('analysis.breakdown')}>
        {breakdown.matchedSkills.length > 0 && (
          <SkillCloud
            label={t('analysis.matched')}
            tone="up"
            skills={breakdown.matchedSkills}
          />
        )}
        {breakdown.missingSkills.length > 0 && (
          <SkillCloud
            label={t('analysis.missing')}
            tone="down"
            skills={breakdown.missingSkills}
          />
        )}
        {breakdown.extraSkills.length > 0 && (
          <SkillCloud
            label={t('analysis.extra')}
            tone="neutral"
            skills={breakdown.extraSkills}
          />
        )}
        {breakdown.explanation && (
          <p className="text-body-sm text-body mt-2">{breakdown.explanation}</p>
        )}
      </Section>
    )
  }

  function ProfileSection({
    profile,
  }: {
    profile: NonNullable<Analysis['profile']>
  }) {
    return (
      <Section title={t('analysis.profile')}>
        <ProfileLine
          label={t('analysis.experience')}
          value={formatYears(profile.yearsExperience, locale, pluralKey, t)}
        />
        {profile.positions.length > 0 && (
          <ProfileLine
            label={t('analysis.positions')}
            value={profile.positions.join(' · ')}
          />
        )}
        {profile.technologies.length > 0 && (
          <ProfileLine
            label={t('analysis.technologies')}
            value={profile.technologies.join(', ')}
          />
        )}
        {profile.education.length > 0 && (
          <ProfileLine
            label={t('analysis.education')}
            value={profile.education.join(' · ')}
          />
        )}
        {profile.summary && (
          <p className="text-body-md text-body mt-3 break-words whitespace-pre-line">
            {profile.summary}
          </p>
        )}
      </Section>
    )
  }
}
