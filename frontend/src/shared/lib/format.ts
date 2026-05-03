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

/**
 * Years experience copy with Russian plural forms (1 год / 2-4 года /
 * 5+ лет). Plural form is picked by the integer floor — "3.5" lands in
 * the "few" bucket → "3.5 года". English collapses to one/many.
 */
export function formatYears(
  years: number,
  locale: 'ru' | 'en',
  pluralKey: (base: string, n: number, l: 'ru' | 'en') => string,
  t: (k: string, vars?: Record<string, string | number>) => string,
): string {
  const display = Number.isInteger(years) ? String(years) : years.toFixed(1)
  const key = pluralKey('analysis.years', Math.floor(years), locale)
  return t(key, { n: display })
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
