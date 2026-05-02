export type UserId = string

export type UserRole = 'user' | 'admin' | string

export type User = {
  id: UserId
  email: string
  role: UserRole
}

export type Session = {
  userId: UserId
  accessToken: string
  refreshToken: string
}

export type Credentials = {
  email: string
  password: string
}
