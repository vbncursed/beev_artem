import { useEffect, useState } from 'react'
import { useGateways } from '@/app/providers/GatewaysProvider'
import { useI18n } from '@/app/providers/I18nProvider'
import { ApiError } from '@/infrastructure/http/errors'
import { BadgePill, Button, Card, PriceCell, Spinner } from '@/presentation/ui'
import type { Analysis } from '@/domain/analysis/types'
import { pluralKey } from '@/shared/i18n/dictionaries'

export function AnalysisDetails({
  analysisId,
  onDeleted,
}: {
  analysisId: string
  onDeleted?: () => void
}) {
  const { t, locale } = useI18n()
  const { analysis: gateway, resume: resumeGateway } = useGateways()
  const [state, setState] = useState<
    | { phase: 'loading' }
    | { phase: 'ready'; analysis: Analysis }
    | { phase: 'error'; message: string }
  >({ phase: 'loading' })
  const [downloadError, setDownloadError] = useState<string | null>(null)
  const [downloading, setDownloading] = useState(false)
  const [deleteError, setDeleteError] = useState<string | null>(null)
  const [confirmingDelete, setConfirmingDelete] = useState(false)
  const [deleting, setDeleting] = useState(false)

  const onDelete = async (candidateId: string) => {
    if (!candidateId || deleting) return
    setDeleteError(null)
    setDeleting(true)
    try {
      await resumeGateway.deleteCandidate(candidateId)
      onDeleted?.()
    } catch (cause) {
      const message =
        cause instanceof ApiError ? cause.message : t('analysis.deleteFailed')
      setDeleteError(message)
      setDeleting(false)
    }
  }

  const onDownload = async (resumeId: string) => {
    if (!resumeId || downloading) return
    setDownloadError(null)
    setDownloading(true)
    try {
      const file = await resumeGateway.downloadResume(resumeId)
      // ArrayBuffer is the universal Blob part — Uint8Array's `buffer` may be
      // SharedArrayBuffer in modern TS lib. Slice copies into a fresh ABuf.
      const buffer = file.data.slice().buffer as ArrayBuffer
      const blob = new Blob([buffer], {
        type: file.fileType || 'application/octet-stream',
      })
      const url = URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      a.download = file.fileName || 'resume'
      document.body.appendChild(a)
      a.click()
      document.body.removeChild(a)
      // Defer revoke so the browser actually uses the URL.
      setTimeout(() => URL.revokeObjectURL(url), 1000)
    } catch (cause) {
      const message =
        cause instanceof ApiError ? cause.message : t('analysis.downloadFailed')
      setDownloadError(message)
    } finally {
      setDownloading(false)
    }
  }

  useEffect(() => {
    let cancelled = false
    setState({ phase: 'loading' })
    void (async () => {
      try {
        const a = await gateway.get(analysisId)
        if (cancelled) return
        setState({ phase: 'ready', analysis: a })
      } catch (cause) {
        if (cancelled) return
        const message =
          cause instanceof ApiError
            ? cause.message
            : t('analysis.error.load')
        setState({ phase: 'error', message })
      }
    })()
    return () => {
      cancelled = true
    }
  }, [gateway, analysisId, t])

  if (state.phase === 'loading') {
    return (
      <Card variant="feature" className="flex items-center justify-center py-16">
        <Spinner size={20} />
      </Card>
    )
  }

  if (state.phase === 'error') {
    return (
      <Card variant="feature">
        <BadgePill tone="down">{t('common.error')}</BadgePill>
        <p className="text-body-md text-body mt-3">{state.message}</p>
      </Card>
    )
  }

  const a = state.analysis

  return (
    <Card variant="feature" className="flex flex-col gap-6">
      <header className="flex items-start justify-between gap-3">
        <div>
          <p className="text-caption-strong text-muted uppercase">
            {t('analysis.matchScore')}
          </p>
          <p className="mt-1">
            <PriceCell
              tone={
                a.matchScore >= 70
                  ? 'up'
                  : a.matchScore < 40
                    ? 'down'
                    : 'neutral'
              }
              value={a.matchScore.toFixed(1)}
              className="text-[44px] leading-none"
            />
          </p>
        </div>
        <div className="flex flex-col items-end gap-2">
          <StatusBadge status={a.status} />
          {a.resumeId && (
            <Button
              variant="secondary-light"
              size="md"
              onClick={() => void onDownload(a.resumeId)}
              loading={downloading}
              iconLeft={<DownloadIcon />}
            >
              {t('analysis.downloadResume')}
            </Button>
          )}
          {a.candidateId && !confirmingDelete && (
            <Button
              variant="text"
              size="md"
              onClick={() => {
                setConfirmingDelete(true)
                setDeleteError(null)
              }}
              iconLeft={<TrashSmall />}
              className="text-body hover:text-semantic-down"
            >
              {t('analysis.deleteCandidate')}
            </Button>
          )}
        </div>
      </header>

      {confirmingDelete && a.candidateId && (
        <div
          role="alertdialog"
          aria-label={t('analysis.deleteConfirm')}
          className="flex flex-wrap items-center justify-end gap-3 rounded-md bg-surface-soft px-4 py-3"
        >
          <span className="text-body-sm text-ink mr-auto">
            {t('analysis.deleteConfirm')}
          </span>
          <Button
            variant="secondary-light"
            size="md"
            onClick={() => setConfirmingDelete(false)}
            disabled={deleting}
          >
            {t('analysis.deleteCancel')}
          </Button>
          <Button
            variant="primary"
            size="md"
            onClick={() => void onDelete(a.candidateId)}
            loading={deleting}
            className="bg-semantic-down hover:opacity-90"
          >
            {t('analysis.deleteConfirmCta')}
          </Button>
        </div>
      )}

      {deleteError && (
        <p role="alert" className="text-caption text-semantic-down -mt-2">
          {deleteError}
        </p>
      )}

      {downloadError && (
        <p role="alert" className="text-caption text-semantic-down -mt-2">
          {downloadError}
        </p>
      )}

      {a.status === 'failed' && a.errorMessage && (
        <p
          role="alert"
          className="text-body-sm text-semantic-down rounded-md bg-surface-soft px-4 py-3"
        >
          {a.errorMessage}
        </p>
      )}

      {a.ai?.hrRecommendation && (
        <Section title={t('analysis.recommendation')}>
          <RecommendationBadge value={a.ai.hrRecommendation} />
          {a.ai.hrRationale && (
            <p className="text-body-md text-body mt-3 break-words whitespace-pre-line">
              {a.ai.hrRationale}
            </p>
          )}
        </Section>
      )}

      {a.breakdown && (
        <Section title={t('analysis.breakdown')}>
          {a.breakdown.matchedSkills.length > 0 && (
            <SkillCloud
              label={t('analysis.matched')}
              tone="up"
              skills={a.breakdown.matchedSkills}
            />
          )}
          {a.breakdown.missingSkills.length > 0 && (
            <SkillCloud
              label={t('analysis.missing')}
              tone="down"
              skills={a.breakdown.missingSkills}
            />
          )}
          {a.breakdown.extraSkills.length > 0 && (
            <SkillCloud
              label={t('analysis.extra')}
              tone="neutral"
              skills={a.breakdown.extraSkills}
            />
          )}
          {a.breakdown.explanation && (
            <p className="text-body-sm text-body mt-2">
              {a.breakdown.explanation}
            </p>
          )}
        </Section>
      )}

      {a.profile && (
        <Section title={t('analysis.profile')}>
          <ProfileLine
            label={t('analysis.experience')}
            value={formatYears(a.profile.yearsExperience, locale, t)}
          />
          {a.profile.positions.length > 0 && (
            <ProfileLine
              label={t('analysis.positions')}
              value={a.profile.positions.join(' · ')}
            />
          )}
          {a.profile.technologies.length > 0 && (
            <ProfileLine
              label={t('analysis.technologies')}
              value={a.profile.technologies.join(', ')}
            />
          )}
          {a.profile.education.length > 0 && (
            <ProfileLine
              label={t('analysis.education')}
              value={a.profile.education.join(' · ')}
            />
          )}
          {a.profile.summary && (
            <p className="text-body-md text-body mt-3 break-words whitespace-pre-line">
              {a.profile.summary}
            </p>
          )}
        </Section>
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
}

function Section({
  title,
  children,
}: {
  title: string
  children: React.ReactNode
}) {
  return (
    <section className="flex flex-col gap-2 border-t border-hairline pt-4 first:border-t-0 first:pt-0">
      <p className="text-caption-strong text-muted uppercase">{title}</p>
      {children}
    </section>
  )
}

function RecommendationBadge({ value }: { value: string }) {
  const { t } = useI18n()
  const v = value.toLowerCase()
  if (v.includes('hire') || v === 'yes')
    return <BadgePill tone="up">{t('rec.hire')}</BadgePill>
  if (v.includes('no') || v.includes('reject'))
    return <BadgePill tone="down">{t('rec.no')}</BadgePill>
  return <BadgePill>{t('rec.maybe')}</BadgePill>
}

function SkillCloud({
  label,
  tone,
  skills,
}: {
  label: string
  tone: 'up' | 'down' | 'neutral'
  skills: string[]
}) {
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

function ProfileLine({ label, value }: { label: string; value: string }) {
  return (
    <div className="flex items-baseline gap-3">
      <span className="text-caption text-muted w-28 shrink-0 uppercase">
        {label}
      </span>
      <span className="text-body-md text-ink min-w-0 break-words">{value}</span>
    </div>
  )
}

/**
 * Years experience copy. Russian uses 1/2-4/5+ form (год/года/лет);
 * English collapses to one/many. We pick the form by the integer floor of
 * the value so "3.5 года" works correctly (3 → "few" → "года" in RU).
 * The displayed number keeps one decimal when it's not an exact int.
 */
function formatYears(
  years: number,
  locale: 'ru' | 'en',
  t: (k: string, vars?: Record<string, string | number>) => string,
): string {
  const display = Number.isInteger(years) ? String(years) : years.toFixed(1)
  const key = pluralKey('analysis.years', Math.floor(years), locale)
  return t(key, { n: display })
}

function TrashSmall() {
  return (
    <svg
      width="14"
      height="14"
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
    </svg>
  )
}

function DownloadIcon() {
  return (
    <svg
      width="14"
      height="14"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
      aria-hidden
    >
      <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4" />
      <path d="M7 10l5 5 5-5" />
      <path d="M12 15V3" />
    </svg>
  )
}

function StatusBadge({ status }: { status: Analysis['status'] }) {
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
