import { useI18n } from '@/app/providers/I18nProvider'
import { AssetIconCircular, PriceCell } from '@/presentation/ui'
import type {
  AnalysisStatus,
  CandidateWithAnalysis,
} from '@/domain/analysis/types'
import { cn } from '@/shared/lib/cn'
import { initials } from '@/shared/lib/format'
import { CandidateStatusBadge } from './CandidateStatusBadge'

type Props = {
  candidate: CandidateWithAnalysis
  selected?: boolean
  onSelect?: () => void
}

export function CandidateRow({ candidate, selected, onSelect }: Props) {
  const { t } = useI18n()
  const tone = scoreTone(candidate.matchScore, candidate.analysisStatus)
  return (
    <button
      type="button"
      onClick={onSelect}
      className={cn(
        'flex w-full cursor-pointer items-center gap-4 rounded-md px-3 py-3 text-left transition-colors',
        'hover:bg-surface-strong',
        'focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary',
        selected && 'bg-surface-strong',
      )}
    >
      <AssetIconCircular>{initials(candidate.fullName)}</AssetIconCircular>
      <div className="min-w-0 flex-1">
        <p className="text-title-sm text-ink truncate">
          {candidate.fullName || t('candidate.unnamed')}
        </p>
        <p className="text-caption text-muted truncate">
          {candidate.email || '—'}
        </p>
      </div>
      <div className="flex flex-col items-end">
        <PriceCell tone={tone} value={formatScore(candidate.matchScore)} />
        <span className="text-caption text-muted hidden sm:inline">
          {t('candidate.match')}
        </span>
      </div>
      <CandidateStatusBadge status={candidate.analysisStatus} />
    </button>
  )
}

function formatScore(score: number): string {
  if (!Number.isFinite(score) || score === 0) return '—'
  return score.toFixed(1)
}

function scoreTone(
  score: number,
  status: AnalysisStatus,
): 'up' | 'down' | 'neutral' {
  if (status !== 'done' || !Number.isFinite(score) || score === 0)
    return 'neutral'
  if (score >= 70) return 'up'
  if (score < 40) return 'down'
  return 'neutral'
}
