import { useEffect, useState } from 'react'
import { useGateways } from '@/app/providers/GatewaysProvider'
import { useI18n } from '@/app/providers/I18nProvider'
import { ApiError } from '@/infrastructure/http/errors'
import type { Vacancy } from '@/domain/vacancy/types'

type State =
  | { phase: 'loading' }
  | { phase: 'ready'; vacancy: Vacancy }
  | { phase: 'error'; message: string }

export function useVacancyDetails(vacancyId: string) {
  const { vacancy: gateway } = useGateways()
  const { t } = useI18n()
  const [state, setState] = useState<State>({ phase: 'loading' })

  useEffect(() => {
    let cancelled = false
    setState({ phase: 'loading' })
    void (async () => {
      try {
        const v = await gateway.get(vacancyId)
        if (cancelled) return
        setState({ phase: 'ready', vacancy: v })
      } catch (cause) {
        if (cancelled) return
        const message =
          cause instanceof ApiError
            ? cause.message
            : t('details.error.loadVacancy')
        setState({ phase: 'error', message })
      }
    })()
    return () => {
      cancelled = true
    }
  }, [gateway, vacancyId, t])

  return state
}
