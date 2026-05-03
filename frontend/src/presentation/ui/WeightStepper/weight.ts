export function parseWeight(raw: string): number | null {
  const trimmed = raw.trim().replace(',', '.')
  if (!trimmed) return null
  const n = Number(trimmed)
  return Number.isFinite(n) ? n : null
}

export function formatWeight(n: number): string {
  if (!Number.isFinite(n)) return '0'
  // Two decimals, then strip trailing zeros so "0.10" → "0.1" and "0" stays "0".
  const s = n.toFixed(2)
  return s.replace(/\.?0+$/, '') || '0'
}
