import { BrowserRouter } from 'react-router-dom'
import { AuthProvider } from '@/app/providers/AuthProvider'
import { GatewaysProvider } from '@/app/providers/GatewaysProvider'
import { I18nProvider } from '@/app/providers/I18nProvider'
import { ThemeProvider } from '@/app/providers/ThemeProvider'
import { AppRouter } from '@/app/router'

export default function App() {
  return (
    <I18nProvider>
      <ThemeProvider>
        <BrowserRouter>
          <AuthProvider>
            <GatewaysProvider>
              <AppRouter />
            </GatewaysProvider>
          </AuthProvider>
        </BrowserRouter>
      </ThemeProvider>
    </I18nProvider>
  )
}
