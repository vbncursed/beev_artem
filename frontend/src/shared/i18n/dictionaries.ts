// Composition root for the i18n layer. The actual translation tables
// live under `locales/<locale>.ts` so each file stays focused (≤300
// LOC) and adding a new language is a single new file + a one-line
// addition to `dictionaries`.
import { en } from './locales/en'
import { ru } from './locales/ru'
import type { Dict, Locale } from './types'

export type { Dict, Locale } from './types'

export const dictionaries: Record<Locale, Dict> = { ru, en }

/**
 * Russian-style pluralisation: 1, 2–4, 5–20, then mod 10.
 * Returns the suffix to append to a base key, e.g. `vacancies.count` →
 *   `vacancies.countOne | countFew | countMany`.
 * For English the mapping collapses to one/many.
 */
export function pluralKey(
  baseKey: string,
  n: number,
  locale: Locale,
): string {
  const abs = Math.abs(n)
  if (locale === 'ru') {
    const mod10 = abs % 10
    const mod100 = abs % 100
    if (mod10 === 1 && mod100 !== 11) return `${baseKey}One`
    if (mod10 >= 2 && mod10 <= 4 && (mod100 < 12 || mod100 > 14))
      return `${baseKey}Few`
    return `${baseKey}Many`
  }
  return abs === 1 ? `${baseKey}One` : `${baseKey}Many`
}
