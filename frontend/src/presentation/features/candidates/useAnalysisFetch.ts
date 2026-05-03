import { useEffect, useState } from 'react'
import { useGateways } from '@/app/providers/GatewaysProvider'
import { useI18n } from '@/app/providers/I18nProvider'
import { ApiError } from '@/infrastructure/http/errors'
import type { Analysis } from '@/domain/analysis/types'

type State =
  | { phase: 'loading' }
  | { phase: 'ready'; analysis: Analysis }
  | { phase: 'error'; message: string }

/**
 * Single-shot fetch for one analysis. Lives outside the component so the
 * render path stays focused on layout — and so the loading/error state
 * machine can be reused by any future analysis-detail surface.
 */
export function useAnalysisFetch(analysisId: string): State {
  const { analysis: gateway } = useGateways()
  const { t } = useI18n()
  const [state, setState] = useState<State>({ phase: 'loading' })

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
          cause instanceof ApiError ? cause.message : t('analysis.error.load')
        setState({ phase: 'error', message })
      }
    })()
    return () => {
      cancelled = true
    }
  }, [gateway, analysisId, t])

  return state
}
