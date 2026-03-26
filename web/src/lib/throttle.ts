/**
 * Returns a throttled wrapper around `fn` that executes at most once every `ms`
 * milliseconds. Timing is measured with performance.now() for sub-millisecond
 * resolution. The first call is always let through immediately.
 */
export function throttle<Args extends unknown[]>(
  fn: (...args: Args) => void,
  ms: number,
): (...args: Args) => void {
  let lastCall = 0
  return (...args: Args) => {
    const now = performance.now()
    if (now - lastCall >= ms) {
      lastCall = now
      fn(...args)
    }
  }
}
