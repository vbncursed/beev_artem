// Type contracts for the i18n layer. Kept in their own file so locale
// dictionaries can import the shape without dragging in pluralKey.
export type Locale = 'ru' | 'en'

/** A flat key→value map. Each locale module exports a default Dict. */
export type Dict = Record<string, string>
