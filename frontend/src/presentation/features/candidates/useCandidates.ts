import { useCallback, useEffect, useState } from 'react'
import { useGateways } from '@/app/providers/GatewaysProvider'
import { useI18n } from '@/app/providers/I18nProvider'
import { ApiError } from '@/infrastructure/http/errors'
import type { CandidateWithAnalysis } from '@/domain/analysis/types'

const PAGE_SIZE = 50

export function useCandidates(vacancyId: string) {
  const { analysis: gateway } = useGateways()
  const { t } = useI18n()
  const [candidates, setCandidates] = useState<CandidateWithAnalysis[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const refresh = useCallback(async () => {
    setLoading(true)
    setError(null)
    try {
      const page = await gateway.listCandidates({
        vacancyId,
        limit: PAGE_SIZE,
        scoreOrder: 'SORT_ORDER_DESC',
      })
      setCandidates(page.candidates)
    } catch (cause) {
      const message =
        cause instanceof ApiError
          ? cause.message
          : t('details.error.loadCandidates')
      setError(message)
    } finally {
      setLoading(false)
    }
  }, [gateway, vacancyId, t])

  useEffect(() => {
    void refresh()
  }, [refresh])

  return { candidates, loading, error, refresh }
}
