import { useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import { useGateways } from '@/app/providers/GatewaysProvider'
import { useI18n } from '@/app/providers/I18nProvider'
import { BadgePill, Button, Card, ErrorCard, Spinner } from '@/presentation/ui'
import { KNOWN_ROLE_VALUES, type SkillWeight } from '@/domain/vacancy/types'
import { ApiError } from '@/infrastructure/http/errors'
import { ResumeUploader } from '@/presentation/features/resume/ResumeUploader'
import { CandidateRow } from '@/presentation/features/candidates/CandidateRow'
import { AnalysisDetails } from '@/presentation/features/candidates/AnalysisDetails'
import { useCandidates } from '@/presentation/features/candidates/useCandidates'
import { useVacancyDetails } from '@/presentation/features/vacancies/useVacancyDetails'
import { VacancyStatusBadge } from '@/presentation/features/vacancies/VacancyStatusBadge'

export function VacancyDetailsPage() {
  const { t } = useI18n()
  const { id = '' } = useParams<{ id: string }>()
  const vacancyState = useVacancyDetails(id)
  const { candidates, loading, error, refresh } = useCandidates(id)
  const [uploadStatus, setUploadStatus] = useState<{
    busy: boolean
    error?: string
    successName?: string
  }>({ busy: false })
  const [selectedAnalysisId, setSelectedAnalysisId] = useState<string | null>(
    null,
  )
  const { resume: resumeGateway, analysis: analysisGateway } = useGateways()

  const onPickFile = async (file: File) => {
    setUploadStatus({ busy: true })
    try {
      const { resume } = await resumeGateway.createCandidateFromResume({
        vacancyId: id,
        file,
      })
      try {
        await analysisGateway.start({
          vacancyId: id,
          resumeId: resume.id,
          useLlm: true,
        })
      } catch {
        // candidate created even if enqueue fails
      }
      setUploadStatus({ busy: false, successName: file.name })
      await refresh()
    } catch (cause) {
      const message =
        cause instanceof ApiError
          ? cause.message
          : t('details.uploadFailure')
      setUploadStatus({ busy: false, error: message })
    }
  }

  const roleLabel =
    vacancyState.phase === 'ready'
      ? KNOWN_ROLE_VALUES.has(vacancyState.vacancy.role)
        ? t(`roles.${vacancyState.vacancy.role}`)
        : vacancyState.vacancy.role
      : ''

  return (
    <>
      <section className="bg-canvas">
        <div className="mx-auto max-w-[1200px] px-6 pt-[96px] pb-12">
          <Link to="/vacancies">
            <Button variant="secondary-light">{t('create.back')}</Button>
          </Link>

          {vacancyState.phase === 'loading' && (
            <div className="mt-8">
              <Spinner size={20} />
            </div>
          )}

          {vacancyState.phase === 'error' && (
            <div className="mt-8">
              <ErrorCard message={vacancyState.message} />
            </div>
          )}

          {vacancyState.phase === 'ready' && (
            <div className="mt-8 flex flex-col gap-4">
              <div className="flex flex-wrap items-center gap-2">
                <BadgePill>{roleLabel}</BadgePill>
                <VacancyStatusBadge status={vacancyState.vacancy.status} />
              </div>
              <h1 className="text-display-md">{vacancyState.vacancy.title}</h1>
              {vacancyState.vacancy.description && (
                <p className="text-body-md text-body max-w-[760px] break-words whitespace-pre-line">
                  {vacancyState.vacancy.description}
                </p>
              )}
              <SkillsSummary skills={vacancyState.vacancy.skills} />
            </div>
          )}
        </div>
      </section>

      <section className="bg-surface-soft">
        <div className="mx-auto max-w-[1200px] px-6 py-12">
          <ResumeUploader busy={uploadStatus.busy} onPick={onPickFile} />
          {uploadStatus.successName && (
            <p role="status" className="text-caption text-semantic-up mt-3">
              {t('details.uploadSuccess', { name: uploadStatus.successName })}
            </p>
          )}
          {uploadStatus.error && (
            <p role="alert" className="text-caption text-semantic-down mt-3">
              {uploadStatus.error}
            </p>
          )}
        </div>
      </section>

      <section className="bg-canvas">
        <div className="mx-auto max-w-[1200px] px-6 py-12 pb-[96px]">
          <div className="flex items-center justify-between pb-6">
            <div>
              <p className="text-caption-strong text-muted uppercase">
                {t('details.candidates')}
              </p>
              <p className="text-title-lg mt-1">
                {candidates.length}{' '}
                <span className="text-body text-body-md font-normal">
                  {t('details.candidatesSubtitle')}
                </span>
              </p>
            </div>
            {loading && <Spinner size={16} />}
          </div>

          {error ? (
            <ErrorCard message={error} />
          ) : candidates.length === 0 && !loading ? (
            <Card variant="feature" className="text-center">
              <div className="mx-auto flex max-w-[420px] flex-col items-center gap-3 py-12">
                <BadgePill>{t('details.empty.badge')}</BadgePill>
                <h3 className="text-title-lg">{t('details.empty.title')}</h3>
                <p className="text-body-md text-body">
                  {t('details.empty.hint')}
                </p>
              </div>
            </Card>
          ) : (
            <div className="grid grid-cols-1 gap-6 lg:grid-cols-12">
              <Card
                variant="feature"
                className="p-2 lg:col-span-7"
                compact={false}
              >
                <ul className="flex flex-col divide-y divide-hairline">
                  {candidates.map((c) => (
                    <li key={c.candidateId}>
                      <CandidateRow
                        candidate={c}
                        selected={c.analysisId === selectedAnalysisId}
                        onSelect={() => setSelectedAnalysisId(c.analysisId)}
                      />
                    </li>
                  ))}
                </ul>
              </Card>

              <div className="lg:col-span-5">
                {selectedAnalysisId ? (
                  <AnalysisDetails
                    analysisId={selectedAnalysisId}
                    onDeleted={() => {
                      setSelectedAnalysisId(null)
                      void refresh()
                    }}
                  />
                ) : (
                  <Card variant="feature" className="text-center">
                    <div className="flex flex-col items-center gap-3 py-12">
                      <BadgePill>{t('details.selectPrompt.badge')}</BadgePill>
                      <p className="text-body-md text-body">
                        {t('details.selectPrompt.text')}
                      </p>
                    </div>
                  </Card>
                )}
              </div>
            </div>
          )}
        </div>
      </section>
    </>
  )
}

function SkillsSummary({ skills }: { skills: SkillWeight[] }) {
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

