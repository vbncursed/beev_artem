import { z } from 'zod'

/**
 * Mirrors backend `validateCreateInput` in
 * vacancy/internal/usecase/validate.go. Backend stays the source of truth;
 * these rules just save a round-trip and surface inline errors.
 *
 *   - title: required, ≤ 255 chars
 *   - description: optional, ≤ 4000 chars
 *   - skills: at least 1; each requires non-empty name and weight 0..1
 *   - role: optional (backend auto-detects from title+description regardless)
 *
 * Backend also normalizes weights: when every weight is 0 it spreads
 * 1/len equally. We accept all-zeros and surface a helper note instead
 * of erroring.
 */

export const skillSchema = z.object({
  name: z
    .string()
    .trim()
    .min(1, 'Skill name is required')
    .max(64, 'Up to 64 characters'),
  weight: z
    .number({ error: 'Number from 0 to 1' })
    .min(0, 'Min 0')
    .max(1, 'Max 1'),
  mustHave: z.boolean().optional(),
  niceToHave: z.boolean().optional(),
})

/**
 * Note: `role` is NOT part of the form. Backend `vacancy/internal/usecase/create.go`
 * always overrides it with `DetectRole(title, description)` before validation,
 * so any user input was ignored. We collapse the field to keep the UI honest.
 */
export const vacancyFormSchema = z.object({
  title: z
    .string()
    .trim()
    .min(1, 'Title is required')
    .max(255, 'Up to 255 characters'),
  description: z
    .string()
    .max(4000, 'Up to 4000 characters')
    .optional()
    .or(z.literal('')),
  skills: z
    .array(skillSchema)
    .min(1, 'Add at least one skill'),
})

export type VacancyFormInput = z.infer<typeof vacancyFormSchema>
export type SkillFormInput = z.infer<typeof skillSchema>
