import {
  createContext,
  use,
  useCallback,
  useEffect,
  useMemo,
  useRef,
  useState,
  type ReactNode,
} from 'react'
import type { Credentials, Session, User } from '@/domain/auth/types'
import { AuthHttpGateway } from '@/infrastructure/auth/AuthHttpGateway'
import { HttpClient, type SessionHolder } from '@/infrastructure/http/client'
import { tokenStorage } from '@/infrastructure/storage/tokenStorage'

type AuthStatus = 'loading' | 'authenticated' | 'anonymous'

type AuthContextValue = {
  status: AuthStatus
  user: User | null
  session: Session | null
  login: (c: Credentials) => Promise<void>
  register: (c: Credentials) => Promise<void>
  logout: () => Promise<void>
}

const AuthContext = createContext<AuthContextValue | null>(null)

export function AuthProvider({ children }: { children: ReactNode }) {
  const [session, setSession] = useState<Session | null>(() =>
    tokenStorage.read(),
  )
  const [user, setUser] = useState<User | null>(null)
  const [status, setStatus] = useState<AuthStatus>(() =>
    tokenStorage.read() ? 'loading' : 'anonymous',
  )

  // Keep latest session in a ref so the SessionHolder closure (created
  // once) always sees current tokens without re-creating the HttpClient.
  const sessionRef = useRef<Session | null>(session)
  useEffect(() => {
    sessionRef.current = session
  }, [session])

  // Single-flight refresh: concurrent 401s coalesce into one /auth/refresh.
  const refreshInFlight = useRef<Promise<Session | null> | null>(null)

  const applySession = useCallback((next: Session | null) => {
    sessionRef.current = next
    setSession(next)
    if (next) tokenStorage.write(next)
    else tokenStorage.clear()
  }, [])

  const { http, gateway } = useMemo(() => {
    const holder: SessionHolder = {
      getAccessToken: () => sessionRef.current?.accessToken ?? null,
      refresh: () => {
        if (!sessionRef.current) return Promise.resolve(null)
        if (refreshInFlight.current) return refreshInFlight.current

        const rt = sessionRef.current.refreshToken
        const promise = (async (): Promise<Session | null> => {
          try {
            const next = await gatewayRef.current!.refresh(rt)
            applySession(next)
            return next
          } catch {
            applySession(null)
            return null
          } finally {
            refreshInFlight.current = null
          }
        })()
        refreshInFlight.current = promise
        return promise
      },
      clear: () => {
        applySession(null)
        setUser(null)
        setStatus('anonymous')
      },
    }
    const httpClient = new HttpClient(holder)
    const authGateway = new AuthHttpGateway(httpClient)
    return { http: httpClient, gateway: authGateway }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [applySession])

  // Late-bound gateway reference avoids the chicken-and-egg between
  // SessionHolder.refresh and AuthHttpGateway construction.
  const gatewayRef = useRef<AuthHttpGateway | null>(null)
  gatewayRef.current = gateway

  // Bootstrap: if we have a stored session, fetch /me to verify and
  // load the user. A 401 here triggers refresh-then-retry inside HttpClient.
  useEffect(() => {
    const stored = sessionRef.current
    if (!stored) {
      setStatus('anonymous')
      return
    }
    let cancelled = false
    ;(async () => {
      try {
        const me = await gateway.me(stored.accessToken)
        if (cancelled) return
        setUser(me)
        setStatus('authenticated')
      } catch {
        if (cancelled) return
        applySession(null)
        setUser(null)
        setStatus('anonymous')
      }
    })()
    return () => {
      cancelled = true
    }
    // Run once on mount.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  const login = useCallback(
    async (c: Credentials) => {
      const next = await gateway.login(c)
      applySession(next)
      const me = await gateway.me(next.accessToken)
      setUser(me)
      setStatus('authenticated')
    },
    [gateway, applySession],
  )

  const register = useCallback(
    async (c: Credentials) => {
      const next = await gateway.register(c)
      applySession(next)
      const me = await gateway.me(next.accessToken)
      setUser(me)
      setStatus('authenticated')
    },
    [gateway, applySession],
  )

  const logout = useCallback(async () => {
    const current = sessionRef.current
    if (current) {
      try {
        await gateway.logout(current.refreshToken)
      } catch {
        // logging out best-effort — proceed regardless
      }
    }
    applySession(null)
    setUser(null)
    setStatus('anonymous')
  }, [gateway, applySession])

  const value = useMemo<AuthContextValue>(
    () => ({ status, user, session, login, register, logout }),
    [status, user, session, login, register, logout],
  )

  // Expose http for sibling gateways (vacancy/resume/analysis) — wired in
  // their own providers via useHttp(). Stash on context too for convenience.
  return (
    <AuthContext value={value}>
      <HttpClientContext value={http}>{children}</HttpClientContext>
    </AuthContext>
  )
}

export function useAuth(): AuthContextValue {
  const ctx = use(AuthContext)
  if (!ctx) throw new Error('useAuth must be used within <AuthProvider>')
  return ctx
}

const HttpClientContext = createContext<HttpClient | null>(null)

export function useHttp(): HttpClient {
  const ctx = use(HttpClientContext)
  if (!ctx) throw new Error('useHttp must be used within <AuthProvider>')
  return ctx
}
