import { Navigate, Route, Routes } from 'react-router-dom'
import { useAuth } from '@/app/providers/AuthProvider'
import { AppLayout } from '@/presentation/layouts/AppLayout'
import { AuthPage } from '@/presentation/pages/AuthPage'
import { VacanciesPage } from '@/presentation/pages/VacanciesPage'
import { VacancyCreatePage } from '@/presentation/pages/VacancyCreatePage'
import { VacancyDetailsPage } from '@/presentation/pages/VacancyDetailsPage'
import { AdminPage } from '@/presentation/pages/AdminPage'
import { PrivacyPage } from '@/presentation/pages/legal/PrivacyPage'
import { TermsPage } from '@/presentation/pages/legal/TermsPage'
import { SupportPage } from '@/presentation/pages/legal/SupportPage'
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

      {/* Legal pages — public, accessible without auth so they can be
          linked from emails / external sites without forcing a login. */}
      <Route path="/privacy" element={<PrivacyPage />} />
      <Route path="/terms" element={<TermsPage />} />
      <Route path="/support" element={<SupportPage />} />

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
        <Route
          path="/admin"
          element={
            <AdminOnly>
              <AdminPage />
            </AdminOnly>
          }
        />
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

// AdminOnly wraps AdminPage. If the caller is not authenticated → /auth;
// if authenticated but role !== 'admin' → silently redirect to /vacancies
// (no error toast — non-admins shouldn't even know this route exists).
function AdminOnly({ children }: { children: ReactNode }) {
  const { status, user } = useAuth()
  if (status === 'loading') return <FullPageLoader />
  if (status === 'anonymous') return <Navigate to="/auth" replace />
  if (user?.role !== 'admin') return <Navigate to="/vacancies" replace />
  return <>{children}</>
}

function FullPageLoader() {
  return (
    <div className="flex min-h-full items-center justify-center bg-canvas">
      <Spinner size={28} />
    </div>
  )
}

