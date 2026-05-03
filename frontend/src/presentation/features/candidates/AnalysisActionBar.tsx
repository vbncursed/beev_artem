import { useState } from 'react'
import { useGateways } from '@/app/providers/GatewaysProvider'
import { useI18n } from '@/app/providers/I18nProvider'
import { ApiError } from '@/infrastructure/http/errors'
import { Button, DownloadIcon, TrashIcon } from '@/presentation/ui'

/**
 * The download + remove actions next to the analysis status badge.
 * Self-contained: owns its own busy/error state for both flows so the
 * parent (AnalysisDetails) doesn't have to thread props for them.
 *
 * On successful delete the parent is told via `onDeleted` so it can
 * clear the selected candidate and refresh the list.
 */
type Props = {
  resumeId: string
  candidateId: string
  onDeleted?: () => void
}

export function AnalysisActionBar({ resumeId, candidateId, onDeleted }: Props) {
  const { t } = useI18n()
  const { resume } = useGateways()
  const [downloading, setDownloading] = useState(false)
  const [downloadError, setDownloadError] = useState<string | null>(null)
  const [deleting, setDeleting] = useState(false)
  const [deleteError, setDeleteError] = useState<string | null>(null)
  const [confirmingDelete, setConfirmingDelete] = useState(false)

  const onDownload = async () => {
    if (!resumeId || downloading) return
    setDownloadError(null)
    setDownloading(true)
    try {
      const file = await resume.downloadResume(resumeId)
      // ArrayBuffer is the universal Blob part — Uint8Array's `buffer`
      // may be SharedArrayBuffer in modern TS lib. Slice copies into
      // a fresh ABuf.
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
        cause instanceof ApiError
          ? cause.message
          : t('analysis.downloadFailed')
      setDownloadError(message)
    } finally {
      setDownloading(false)
    }
  }

  const onDelete = async () => {
    if (!candidateId || deleting) return
    setDeleteError(null)
    setDeleting(true)
    try {
      await resume.deleteCandidate(candidateId)
      onDeleted?.()
    } catch (cause) {
      const message =
        cause instanceof ApiError ? cause.message : t('analysis.deleteFailed')
      setDeleteError(message)
      setDeleting(false)
    }
  }

  return (
    <>
      <div className="flex flex-col items-end gap-2">
        {resumeId && (
          <Button
            variant="secondary-light"
            size="md"
            onClick={() => void onDownload()}
            loading={downloading}
            iconLeft={<DownloadIcon />}
          >
            {t('analysis.downloadResume')}
          </Button>
        )}
        {candidateId && !confirmingDelete && (
          <Button
            variant="text"
            size="md"
            onClick={() => {
              setConfirmingDelete(true)
              setDeleteError(null)
            }}
            iconLeft={<TrashIcon size={14} />}
            className="text-body hover:text-semantic-down"
          >
            {t('analysis.deleteCandidate')}
          </Button>
        )}
      </div>

      {confirmingDelete && candidateId && (
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
            onClick={() => void onDelete()}
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
    </>
  )
}
