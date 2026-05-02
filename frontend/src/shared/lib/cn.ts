export type ClassValue =
  | string
  | number
  | null
  | undefined
  | false
  | ClassValue[]
  | { [key: string]: unknown }

export function cn(...inputs: ClassValue[]): string {
  const out: string[] = []
  for (const v of inputs) {
    if (!v) continue
    if (typeof v === 'string' || typeof v === 'number') {
      out.push(String(v))
    } else if (Array.isArray(v)) {
      const nested = cn(...v)
      if (nested) out.push(nested)
    } else if (typeof v === 'object') {
      for (const key in v) if (v[key]) out.push(key)
    }
  }
  return out.join(' ')
}
