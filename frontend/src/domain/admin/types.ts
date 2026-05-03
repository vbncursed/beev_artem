/**
 * Mirror of admin/internal/domain/domain.go. Numeric counters arrive as
 * strings from grpc-gateway (proto uint64 → string in JSON to dodge
 * JS number precision loss); we coerce to JS number on parse — safe
 * for the magnitudes we expect on this dashboard.
 */
export type SystemStats = {
  usersTotal: number
  adminsTotal: number
  vacanciesTotal: number
  candidatesTotal: number
  analysesTotal: number
  analysesDone: number
  analysesFailed: number
}

export type AdminUserView = {
  id: number
  email: string
  role: 'admin' | 'user' | string
  createdAt: string
  vacanciesOwned: number
  candidatesUploaded: number
}
