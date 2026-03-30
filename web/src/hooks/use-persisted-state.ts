/**
 * usePersistedState
 *
 * A thin bridge between the browser's localStorage API and React state.
 * Initialises from storage on mount; persists on every write.
 *
 * When both the stored value and `fallback` are plain objects they are merged
 * on read, so new fields added to `fallback` in later app versions always have
 * a value even when an older entry is already in storage.
 */
import { useState } from 'react'

interface Options<T> {
  /** localStorage key. */
  key: string
  /** Value used when nothing is stored yet (also used as the merge base). */
  fallback: T
}

export function usePersistedState<T>({ key, fallback }: Options<T>) {
  const [value, setValueRaw] = useState<T>(() => {
    try {
      const raw = localStorage.getItem(key)
      if (raw !== null) {
        const parsed = JSON.parse(raw) as T
        if (
          typeof fallback === 'object' && fallback !== null &&
          typeof parsed  === 'object' && parsed  !== null
        ) {
          return { ...fallback, ...parsed } as T
        }
        return parsed
      }
    } catch { /* localStorage unavailable or JSON malformed */ }
    return fallback
  })

  function setValue(updater: T | ((prev: T) => T)) {
    setValueRaw(prev => {
      const next = typeof updater === 'function'
        ? (updater as (prev: T) => T)(prev)
        : updater
      try { localStorage.setItem(key, JSON.stringify(next)) } catch {}
      return next
    })
  }

  return [value, setValue] as const
}
