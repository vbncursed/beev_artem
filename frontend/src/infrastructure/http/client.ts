import type { Session } from '@/domain/auth/types'
import { ApiError, toApiError } from './errors'

/**
 * Thin fetch wrapper.
 *
 * - Base URL comes from Vite proxy in dev (`/api`) or VITE_API_URL in prod.
 * - Bearer token is injected by `SessionHolder.getAccessToken()`.
 * - On 401 the client calls `SessionHolder.refresh()` once and retries the
 *   request. Concurrent 401s are coalesced into a single refresh promise.
 * - Errors are normalized to `ApiError`.
 */

export type Json = unknown

export type RequestOptions = {
  method?: 'GET' | 'POST' | 'PATCH' | 'PUT' | 'DELETE'
  body?: Json | FormData
  headers?: Record<string, string>
  /** Skip Authorization header — used by /auth/login, /register, /refresh. */
  noAuth?: boolean
  /** Already a retry — don't recurse on 401. */
  noRetry?: boolean
  signal?: AbortSignal
}

export interface SessionHolder {
  getAccessToken(): string | null
  /**
   * Refresh the access token. Returns the new session, or null if the
   * refresh failed (caller should bubble the original 401).
   *
   * Implementations MUST coalesce concurrent calls into a single
   * refresh request.
   */
  refresh(): Promise<Session | null>
  /** Called when refresh fails — clears stored session and notifies UI. */
  clear(): void
}

const BASE_URL = (import.meta.env.VITE_API_URL ?? '').replace(/\/$/, '')

export class HttpClient {
  private readonly session: SessionHolder

  constructor(session: SessionHolder) {
    this.session = session
  }

  async request<T = Json>(path: string, opts: RequestOptions = {}): Promise<T> {
    const url = `${BASE_URL}${path}`
    const headers: Record<string, string> = { ...(opts.headers ?? {}) }

    const isFormData = opts.body instanceof FormData
    if (opts.body !== undefined && !isFormData) {
      headers['Content-Type'] ??= 'application/json'
    }

    if (!opts.noAuth) {
      const token = this.session.getAccessToken()
      if (token) headers['Authorization'] = `Bearer ${token}`
    }

    let response: Response
    try {
      response = await fetch(url, {
        method: opts.method ?? (opts.body !== undefined ? 'POST' : 'GET'),
        headers,
        body:
          opts.body === undefined
            ? undefined
            : isFormData
              ? (opts.body as FormData)
              : JSON.stringify(opts.body),
        signal: opts.signal,
      })
    } catch (cause) {
      throw new ApiError({
        status: 0,
        reason: 'NETWORK',
        message: 'Network unreachable',
        details: cause,
      })
    }

    if (response.status === 401 && !opts.noAuth && !opts.noRetry) {
      const refreshed = await this.session.refresh()
      if (refreshed) {
        return this.request<T>(path, { ...opts, noRetry: true })
      }
      this.session.clear()
    }

    if (!response.ok) {
      const body = await safeJson(response)
      throw toApiError(response.status, body)
    }

    if (response.status === 204) return undefined as T
    const ct = response.headers.get('content-type') ?? ''
    if (!ct.includes('application/json')) return undefined as T
    return (await response.json()) as T
  }

  get<T = Json>(path: string, opts?: Omit<RequestOptions, 'method' | 'body'>) {
    return this.request<T>(path, { ...opts, method: 'GET' })
  }
  post<T = Json>(
    path: string,
    body?: Json | FormData,
    opts?: Omit<RequestOptions, 'method' | 'body'>,
  ) {
    return this.request<T>(path, { ...opts, method: 'POST', body })
  }
  patch<T = Json>(
    path: string,
    body?: Json,
    opts?: Omit<RequestOptions, 'method' | 'body'>,
  ) {
    return this.request<T>(path, { ...opts, method: 'PATCH', body })
  }
  delete<T = Json>(path: string, opts?: Omit<RequestOptions, 'method' | 'body'>) {
    return this.request<T>(path, { ...opts, method: 'DELETE' })
  }
}

async function safeJson(response: Response): Promise<unknown> {
  try {
    const text = await response.text()
    if (!text) return null
    return JSON.parse(text)
  } catch {
    return null
  }
}
