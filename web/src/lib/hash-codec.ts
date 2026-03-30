/**
 * Generic hash and encode utilities.
 *
 * Nothing in here is specific to the reader or to any particular data shape —
 * callers supply the storage prefix and the payload type.
 */

// ── Hash ──────────────────────────────────────────────────────────────────────

/** FNV-1a 32-bit hash → 8-char lowercase hex string. */
export function fnv1a(s: string): string {
  let h = 0x811c9dc5
  for (let i = 0; i < s.length; i++) {
    h ^= s.charCodeAt(i)
    h = Math.imul(h, 0x01000193) >>> 0
  }
  return h.toString(16).padStart(8, '0')
}

// ── Compound hash strings  ────────────────────────────────────────────────────
//
// Format: `{primary}{sep}{payloadKey}`
//
//   primary    — caller-defined string (e.g. scroll position token)
//   payloadKey — fnv1a fingerprint of JSON.stringify(payload)
//   sep        — single character that cannot appear in either segment (default "~")
//
// The full payload JSON is stored in localStorage under `storagePrefix + payloadKey`
// so it can be recovered from the fingerprint alone.

/**
 * Fingerprint `payload` with fnv1a, persist its JSON to localStorage, and
 * return the compound string `${primary}${sep}${key}`.
 */
export function encodeHash(
  primary: string,
  payload: unknown,
  storagePrefix: string,
  sep = '~',
): string {
  const json = JSON.stringify(payload)
  const key = fnv1a(json)
  try { localStorage.setItem(storagePrefix + key, json) } catch { /* private mode */ }
  return `${primary}${sep}${key}`
}

/**
 * Extract the primary segment from a compound hash string.
 * Returns the full string unchanged when the separator is absent.
 */
export function primaryOf(combined: string, sep = '~'): string {
  const idx = combined.indexOf(sep)
  return idx !== -1 ? combined.slice(0, idx) : combined
}

/**
 * Recover and parse the payload stored under the fingerprint in localStorage.
 * When both `defaults` and the stored value are plain objects the two are merged
 * (spread) so new fields added in later app versions always have a fallback value.
 * Returns null when the segment is absent, the key is not in storage, or parsing fails.
 */
export function decodeHash<T>(
  combined: string,
  storagePrefix: string,
  defaults: T,
  sep = '~',
): T | null {
  const idx = combined.indexOf(sep)
  if (idx === -1) return null
  try {
    const raw = localStorage.getItem(storagePrefix + combined.slice(idx + 1))
    if (!raw) return null
    const parsed = JSON.parse(raw) as T
    if (
      typeof defaults === 'object' && defaults !== null &&
      typeof parsed  === 'object' && parsed  !== null
    ) {
      return { ...defaults, ...parsed }
    }
    return parsed
  } catch {
    return null
  }
}
