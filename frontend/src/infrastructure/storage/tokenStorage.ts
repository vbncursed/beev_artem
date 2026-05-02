import type { Session } from '@/domain/auth/types'

const KEY = 'cadence:session'

export const tokenStorage = {
  read(): Session | null {
    try {
      const raw = localStorage.getItem(KEY)
      if (!raw) return null
      const parsed = JSON.parse(raw) as Partial<Session>
      if (
        typeof parsed.userId === 'string' &&
        typeof parsed.accessToken === 'string' &&
        typeof parsed.refreshToken === 'string'
      ) {
        return parsed as Session
      }
      return null
    } catch {
      return null
    }
  },
  write(session: Session): void {
    try {
      localStorage.setItem(KEY, JSON.stringify(session))
    } catch {
      // ignore quota / disabled storage
    }
  },
  clear(): void {
    try {
      localStorage.removeItem(KEY)
    } catch {
      // ignore
    }
  },
}
