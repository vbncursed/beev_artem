import type { Credentials, Session } from '@/domain/auth/types'
import type { AuthGateway } from './ports'

/**
 * Thin use-cases. Validation lives in `domain/auth/rules`; the gateway
 * is the only collaborator. Real value comes from the boundary they
 * draw between presentation and infrastructure.
 */
export const loginUseCase =
  (gateway: AuthGateway) =>
  (credentials: Credentials): Promise<Session> =>
    gateway.login(credentials)

export const registerUseCase =
  (gateway: AuthGateway) =>
  (credentials: Credentials): Promise<Session> =>
    gateway.register(credentials)

export const logoutUseCase =
  (gateway: AuthGateway) =>
  (refreshToken: string): Promise<void> =>
    gateway.logout(refreshToken)
