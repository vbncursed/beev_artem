import { z } from 'zod'

/**
 * Client-side validation. Mirrors the auth service's request validators
 * (see beev/auth/internal/usecase). Backend remains the source of truth —
 * these rules just save a round-trip and surface inline errors.
 */
export const emailSchema = z
  .string()
  .min(1, 'Email is required')
  .email('Invalid email format')

export const passwordSchema = z
  .string()
  .min(8, 'At least 8 characters')
  .max(72, 'Maximum 72 characters')

export const credentialsSchema = z.object({
  email: emailSchema,
  password: passwordSchema,
})

export const registerSchema = credentialsSchema
  .extend({
    confirmPassword: z.string(),
  })
  .refine((d) => d.password === d.confirmPassword, {
    path: ['confirmPassword'],
    message: 'Passwords do not match',
  })

export type CredentialsInput = z.infer<typeof credentialsSchema>
export type RegisterInput = z.infer<typeof registerSchema>
