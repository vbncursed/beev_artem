import { useI18n } from '@/app/providers/I18nProvider'
import { BadgePill } from '@/presentation/ui'
import type { AnalysisStatus } from '@/domain/analysis/types'

/**
 * Single source of truth for analysis-status pills. Was duplicated inline
 * in CandidateRow and AnalysisDetails — same switch, same labels.
 */
export function CandidateStatusBadge({ status }: { status: AnalysisStatus }) {
  const { t } = useI18n()
  switch (status) {
    case 'done':
      return <BadgePill tone="up">{t('analysis.status.done')}</BadgePill>
    case 'failed':
      return <BadgePill tone="down">{t('analysis.status.failed')}</BadgePill>
    case 'queued':
      return <BadgePill>{t('analysis.status.queued')}</BadgePill>
    case 'running':
      return <BadgePill>{t('analysis.status.running')}</BadgePill>
    default:
      return <BadgePill>{t('analysis.status.unknown')}</BadgePill>
  }
}
