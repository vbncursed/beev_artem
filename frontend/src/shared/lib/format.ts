/**
 * Tiny formatters reused across pages. Keep them pure and locale-agnostic
 * (locale arg passed in) — no hidden access to I18nProvider here so they
 * can be called from selectors / use-cases too.
 */

/** Initials from a full name or email. Empty input → "?". */
export function initials(input: string): string {
  if (!input) return '?'
  const cleaned = input.split('@')[0]?.trim() ?? input.trim()
  const parts = cleaned.split(/\s+/).slice(0, 2)
  const out = parts.map((p) => p[0]?.toUpperCase() ?? '').join('')
  return out || '?'
}

/** Human-friendly date in the user's locale. ISO in, "03 May 2026" out. */
export function formatDate(iso: string, locale: 'ru' | 'en' = 'ru'): string {
  if (!iso) return ''
  const d = new Date(iso)
  if (Number.isNaN(d.getTime())) return ''
  const lang = locale === 'ru' ? 'ru-RU' : 'en-GB'
  return d.toLocaleDateString(lang, {
    day: '2-digit',
    month: 'short',
    year: 'numeric',
  })
}
