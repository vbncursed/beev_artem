/**
 * Backend errors come from grpc-gateway. Each service returns either:
 *   1. A `google.rpc.Status` JSON: { code, message, details: [...errdetails.ErrorInfo] }
 *   2. A plain { error: string } (rare)
 *   3. Empty body with HTTP status only
 *
 * We normalize all three into `ApiError` so use-cases never touch transport.
 */

export type ApiErrorReason =
  | 'UNAUTHORIZED'
  | 'FORBIDDEN'
  | 'NOT_FOUND'
  | 'INVALID_ARGUMENT'
  | 'CONFLICT'
  | 'RATE_LIMITED'
  | 'NETWORK'
  | 'UNKNOWN'

export class ApiError extends Error {
  readonly status: number
  readonly reason: ApiErrorReason
  readonly domain?: string
  readonly details?: unknown

  constructor(init: {
    status: number
    reason: ApiErrorReason
    message: string
    domain?: string
    details?: unknown
  }) {
    super(init.message)
    this.name = 'ApiError'
    this.status = init.status
    this.reason = init.reason
    this.domain = init.domain
    this.details = init.details
  }
}

type GrpcStatusBody = {
  code?: number
  message?: string
  details?: Array<{
    '@type'?: string
    reason?: string
    domain?: string
    metadata?: Record<string, string>
  }>
}

const REASON_BY_STATUS: Record<number, ApiErrorReason> = {
  400: 'INVALID_ARGUMENT',
  401: 'UNAUTHORIZED',
  403: 'FORBIDDEN',
  404: 'NOT_FOUND',
  409: 'CONFLICT',
  429: 'RATE_LIMITED',
}

export function toApiError(status: number, body: unknown): ApiError {
  const fallbackReason: ApiErrorReason =
    REASON_BY_STATUS[status] ?? (status >= 500 ? 'UNKNOWN' : 'UNKNOWN')

  if (body && typeof body === 'object') {
    const grpc = body as GrpcStatusBody
    const errInfo = grpc.details?.find((d) =>
      typeof d['@type'] === 'string'
        ? d['@type'].includes('ErrorInfo')
        : Boolean(d.reason),
    )
    return new ApiError({
      status,
      reason: errInfo?.reason
        ? mapReason(errInfo.reason, fallbackReason)
        : fallbackReason,
      message: grpc.message ?? defaultMessage(fallbackReason),
      domain: errInfo?.domain,
      details: grpc.details,
    })
  }

  return new ApiError({
    status,
    reason: fallbackReason,
    message: defaultMessage(fallbackReason),
  })
}

function mapReason(
  raw: string,
  fallback: ApiErrorReason,
): ApiErrorReason {
  const r = raw.toUpperCase()
  if (r.includes('UNAUTHORIZED') || r.includes('UNAUTHENTICATED')) {
    return 'UNAUTHORIZED'
  }
  if (r.includes('FORBIDDEN') || r.includes('PERMISSION')) return 'FORBIDDEN'
  if (r.includes('NOT_FOUND')) return 'NOT_FOUND'
  if (r.includes('CONFLICT') || r.includes('ALREADY_EXISTS')) return 'CONFLICT'
  if (r.includes('INVALID') || r.includes('ARGUMENT'))
    return 'INVALID_ARGUMENT'
  if (r.includes('RATE') || r.includes('LIMIT')) return 'RATE_LIMITED'
  return fallback
}

function defaultMessage(reason: ApiErrorReason): string {
  switch (reason) {
    case 'UNAUTHORIZED':
      return 'Authentication required'
    case 'FORBIDDEN':
      return 'Access denied'
    case 'NOT_FOUND':
      return 'Not found'
    case 'INVALID_ARGUMENT':
      return 'Invalid request'
    case 'CONFLICT':
      return 'Conflict with current state'
    case 'RATE_LIMITED':
      return 'Too many requests, slow down'
    case 'NETWORK':
      return 'Network error'
    default:
      return 'Something went wrong'
  }
}
