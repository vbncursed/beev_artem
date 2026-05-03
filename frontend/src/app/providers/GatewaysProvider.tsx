import { createContext, use, useMemo, type ReactNode } from 'react'
import type { AdminGateway } from '@/application/admin/ports'
import type { AnalysisGateway } from '@/application/analysis/ports'
import type { ResumeGateway } from '@/application/resume/ports'
import type { VacancyGateway } from '@/application/vacancy/ports'
import { useHttp } from '@/app/providers/AuthProvider'
import { AdminHttpGateway } from '@/infrastructure/admin/AdminHttpGateway'
import { AnalysisHttpGateway } from '@/infrastructure/analysis/AnalysisHttpGateway'
import { ResumeHttpGateway } from '@/infrastructure/resume/ResumeHttpGateway'
import { VacancyHttpGateway } from '@/infrastructure/vacancy/VacancyHttpGateway'

type Gateways = {
  vacancy: VacancyGateway
  resume: ResumeGateway
  analysis: AnalysisGateway
  admin: AdminGateway
}

const GatewaysContext = createContext<Gateways | null>(null)

export function GatewaysProvider({ children }: { children: ReactNode }) {
  const http = useHttp()
  const value = useMemo<Gateways>(
    () => ({
      vacancy: new VacancyHttpGateway(http),
      resume: new ResumeHttpGateway(http),
      analysis: new AnalysisHttpGateway(http),
      admin: new AdminHttpGateway(http),
    }),
    [http],
  )
  return <GatewaysContext value={value}>{children}</GatewaysContext>
}

export function useGateways(): Gateways {
  const ctx = use(GatewaysContext)
  if (!ctx)
    throw new Error('useGateways must be used within <GatewaysProvider>')
  return ctx
}
