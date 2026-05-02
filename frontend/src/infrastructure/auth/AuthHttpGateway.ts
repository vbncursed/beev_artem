import type { AuthGateway } from '@/application/auth/ports'
import type { Credentials, Session, User } from '@/domain/auth/types'
import type { HttpClient } from '@/infrastructure/http/client'

type AuthResponseDto = {
  userId: string
  accessToken: string
  refreshToken: string
}

type MeResponseDto = {
  userId: string
  email: string
  role: string
}

const PATH = {
  login: '/api/v1/auth/login',
  register: '/api/v1/auth/register',
  refresh: '/api/v1/auth/refresh',
  me: '/api/v1/auth/me',
  logout: '/api/v1/auth/logout',
} as const

export class AuthHttpGateway implements AuthGateway {
  private readonly http: HttpClient

  constructor(http: HttpClient) {
    this.http = http
  }

  async login(c: Credentials): Promise<Session> {
    const dto = await this.http.request<AuthResponseDto>(PATH.login, {
      method: 'POST',
      body: c,
      noAuth: true,
    })
    return toSession(dto)
  }

  async register(c: Credentials): Promise<Session> {
    const dto = await this.http.request<AuthResponseDto>(PATH.register, {
      method: 'POST',
      body: c,
      noAuth: true,
    })
    return toSession(dto)
  }

  async refresh(refreshToken: string): Promise<Session> {
    const dto = await this.http.request<AuthResponseDto>(PATH.refresh, {
      method: 'POST',
      body: { refreshToken },
      noAuth: true,
      noRetry: true,
    })
    return toSession(dto)
  }

  async me(accessToken: string): Promise<User> {
    const dto = await this.http.request<MeResponseDto>(PATH.me, {
      method: 'GET',
      headers: { Authorization: `Bearer ${accessToken}` },
      noAuth: true,
    })
    return { id: dto.userId, email: dto.email, role: dto.role }
  }

  async logout(refreshToken: string): Promise<void> {
    await this.http.request(PATH.logout, {
      method: 'POST',
      body: { refreshToken },
    })
  }
}

function toSession(dto: AuthResponseDto): Session {
  return {
    userId: dto.userId,
    accessToken: dto.accessToken,
    refreshToken: dto.refreshToken,
  }
}
