/**
 * useInterval
 *
 * Runs `callback` on a fixed interval of `ms` milliseconds.
 * Pass `null` as `ms` to pause the interval without unmounting.
 *
 * The callback ref is updated on every render so the function is never stale —
 * callers do not need to memoise or ref-wrap their callbacks before passing them in.
 */
import { useEffect, useRef } from 'react'

export function useInterval(callback: () => void, ms: number | null): void {
  const callbackRef = useRef(callback)
  callbackRef.current = callback

  useEffect(() => {
    if (ms === null) return
    const id = setInterval(() => callbackRef.current(), ms)
    return () => clearInterval(id)
  }, [ms])
}
