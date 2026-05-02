import { useEffect, useMemo, useState } from 'react'
import { useGateways } from '@/app/providers/GatewaysProvider'
import { useI18n } from '@/app/providers/I18nProvider'
import { ApiError } from '@/infrastructure/http/errors'
import type { Vacancy } from '@/domain/vacancy/types'

type State =
  | { phase: 'loading'; vacancies: Vacancy[] }
  | { phase: 'ready'; vacancies: Vacancy[]; total: number }
  | { phase: 'error'; vacancies: Vacancy[]; message: string }

const PAGE_SIZE = 24

export function useVacancies(query: string) {
  const { vacancy: gateway } = useGateways()
  const { t } = useI18n()
  const [state, setState] = useState<State>({
    phase: 'loading',
    vacancies: [],
  })

  useEffect(() => {
    const controller = new AbortController()
    setState((prev) => ({ phase: 'loading', vacancies: prev.vacancies }))

    void (async () => {
      try {
        const page = await gateway.list({ query, limit: PAGE_SIZE })
        if (controller.signal.aborted) return
        setState({
          phase: 'ready',
          vacancies: page.vacancies,
          total: page.total,
        })
      } catch (cause) {
        if (controller.signal.aborted) return
        const message =
          cause instanceof ApiError
            ? cause.message
            : t('vacancies.error.loadFailed')
        setState((prev) => ({
          phase: 'error',
          vacancies: prev.vacancies,
          message,
        }))
      }
    })()

    return () => controller.abort()
  }, [gateway, query, t])

  return useMemo(
    () => ({
      vacancies: state.vacancies,
      isLoading: state.phase === 'loading',
      error: state.phase === 'error' ? state.message : null,
      total: state.phase === 'ready' ? state.total : state.vacancies.length,
    }),
    [state],
  )
}
