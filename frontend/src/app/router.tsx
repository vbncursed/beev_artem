import { Navigate, Route, Routes } from 'react-router-dom'
import { useAuth } from '@/app/providers/AuthProvider'
import { AppLayout } from '@/presentation/layouts/AppLayout'
import { AuthPage } from '@/presentation/pages/AuthPage'
import { VacanciesPage } from '@/presentation/pages/VacanciesPage'
import { VacancyCreatePage } from '@/presentation/pages/VacancyCreatePage'
import { VacancyDetailsPage } from '@/presentation/pages/VacancyDetailsPage'
import { Spinner } from '@/presentation/ui'
import { type ReactNode } from 'react'

/**
 * Top-level routes. The router is intentionally flat — pages compose
 * their own layouts; only `/auth` opts out of the AppLayout chrome.
 */
export function AppRouter() {
  return (
    <Routes>
      <Route
        path="/auth"
        element={
          <PublicOnly>
            <AuthPage />
          </PublicOnly>
        }
      />

      <Route
        element={
          <Protected>
            <AppLayout />
          </Protected>
        }
      >
        <Route index element={<Navigate to="/vacancies" replace />} />
        <Route path="/vacancies" element={<VacanciesPage />} />
        <Route path="/vacancies/new" element={<VacancyCreatePage />} />
        <Route path="/vacancies/:id" element={<VacancyDetailsPage />} />
      </Route>

      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  )
}

function Protected({ children }: { children: ReactNode }) {
  const { status } = useAuth()
  if (status === 'loading') return <FullPageLoader />
  if (status === 'anonymous') return <Navigate to="/auth" replace />
  return <>{children}</>
}

function PublicOnly({ children }: { children: ReactNode }) {
  const { status } = useAuth()
  if (status === 'loading') return <FullPageLoader />
  if (status === 'authenticated') return <Navigate to="/vacancies" replace />
  return <>{children}</>
}

function FullPageLoader() {
  return (
    <div className="flex min-h-full items-center justify-center bg-canvas">
      <Spinner size={28} />
    </div>
  )
}

