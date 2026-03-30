/**
 * useAutoScroll
 *
 * Scrolls the page downward at `speed` px per tick (~16 ms) while `active` is true.
 * Pauses automatically whenever the user has a finger on the screen (isTouchingRef).
 * Calls `onStop` and deactivates when the bottom of the page is reached.
 *
 * Delegates the interval management to useInterval so the callback is always
 * current — no manual onStop ref needed here.
 */
import type { MutableRefObject } from 'react'
import { useInterval } from '@/hooks/use-interval'

export function useAutoScroll(
  active: boolean,
  speed: number,
  isTouchingRef: MutableRefObject<boolean>,
  onStop: () => void,
): void {
  useInterval(() => {
    if (isTouchingRef.current) return
    const el = document.documentElement
    if (el.scrollTop + el.clientHeight >= el.scrollHeight - 2) {
      onStop()
      return
    }
    window.scrollBy(0, speed)
  }, active ? 16 : null)
}
