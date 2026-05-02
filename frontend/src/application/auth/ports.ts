import type { Credentials, Session, User } from '@/domain/auth/types'

/**
 * Application-layer port. Implementations live in `infrastructure/auth/`.
 * Use-cases depend on this interface, not on HTTP details.
 */
export interface AuthGateway {
  login(credentials: Credentials): Promise<Session>
  register(credentials: Credentials): Promise<Session>
  refresh(refreshToken: string): Promise<Session>
  me(accessToken: string): Promise<User>
  logout(refreshToken: string): Promise<void>
}
