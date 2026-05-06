import { useRef, useState, type DragEvent } from 'react'
import { useI18n } from '@/app/providers/I18nProvider'
import { Button, Card } from '@/presentation/ui'
import { ACCEPTED_RESUME_TYPES, MAX_RESUME_BYTES } from '@/domain/resume/types'
import { cn } from '@/shared/lib/cn'

type Props = {
  busy: boolean
  onPick: (files: File[]) => void
}

export function ResumeUploader({ busy, onPick }: Props) {
  const { t } = useI18n()
  const inputRef = useRef<HTMLInputElement>(null)
  const [over, setOver] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const accept = (list: FileList | null | undefined) => {
    setError(null)
    if (!list || list.length === 0) return
    const files = Array.from(list)
    const tooBig = files.find((f) => f.size > MAX_RESUME_BYTES)
    if (tooBig) {
      setError(t('upload.tooLarge', { name: tooBig.name }))
      return
    }
    onPick(files)
  }

  const onDrop = (e: DragEvent<HTMLDivElement>) => {
    e.preventDefault()
    setOver(false)
    accept(e.dataTransfer.files)
  }

  return (
    <Card
      variant="product-light"
      className={cn(
        'relative cursor-pointer border-2 border-dashed transition-colors',
        over ? 'border-primary' : 'border-hairline',
        busy && 'pointer-events-none opacity-60',
      )}
      onDragOver={(e) => {
        e.preventDefault()
        setOver(true)
      }}
      onDragLeave={() => setOver(false)}
      onDrop={onDrop}
      onClick={() => inputRef.current?.click()}
    >
      <input
        ref={inputRef}
        type="file"
        multiple
        accept={ACCEPTED_RESUME_TYPES}
        className="hidden"
        onChange={(e) => {
          accept(e.target.files)
          e.target.value = ''
        }}
      />
      <div className="flex flex-col items-center gap-3 py-6 text-center">
        <UploadGlyph />
        <p className="text-title-md">{t('upload.title')}</p>
        <p className="text-body-sm text-body max-w-[420px]">
          {t('upload.subtitle')}
        </p>
        <Button
          variant="primary"
          type="button"
          loading={busy}
          onClick={(e) => {
            e.stopPropagation()
            inputRef.current?.click()
          }}
        >
          {busy ? t('upload.uploading') : t('upload.cta')}
        </Button>
        {error && (
          <p role="alert" className="text-caption text-semantic-down">
            {error}
          </p>
        )}
      </div>
    </Card>
  )
}

function UploadGlyph() {
  return (
    <span className="inline-flex size-12 items-center justify-center rounded-full bg-surface-strong text-primary">
      <svg
        viewBox="0 0 24 24"
        width="20"
        height="20"
        fill="none"
        stroke="currentColor"
        strokeWidth="2"
        strokeLinecap="round"
        strokeLinejoin="round"
        aria-hidden
      >
        <path d="M12 16V4" />
        <path d="m6 10 6-6 6 6" />
        <path d="M4 20h16" />
      </svg>
    </span>
  )
}
