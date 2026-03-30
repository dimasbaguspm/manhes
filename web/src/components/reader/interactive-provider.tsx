/**
 * InteractiveProvider
 *
 * Observes the root-level window events (scroll, touch) and manages the header
 * visibility toggle driven by double-tap/click. Consumers call useInteractive()
 * to subscribe to these signals without wiring their own event listeners.
 */
import {
  createContext,
  useContext,
  useState,
  useEffect,
  useRef,
  useCallback,
  type ReactNode,
  type MutableRefObject,
  type Dispatch,
  type SetStateAction,
} from 'react'

export interface InteractiveCtx {
  /** 0–100 reading progress percentage derived from window scroll position. */
  scrollPct: number
  /** Whether the top header is currently visible. */
  headerVisible: boolean
  /** Toggle or set header visibility directly. */
  setHeaderVisible: Dispatch<SetStateAction<boolean>>
  /**
   * Ref that is true while the user has at least one finger on the screen.
   * Read-only for consumers; mutated internally by touch listeners.
   * Designed to be read inside intervals/rAF without causing re-renders.
   */
  isTouchingRef: MutableRefObject<boolean>
  /** Call on pointerdown on the manga strip. Handles double-tap and hold detection. */
  onStripPointerDown: () => void
  /** Call on pointerup on the manga strip. Completes tap or hold gesture. */
  onStripPointerUp: () => void
  /** Call on pointercancel on the manga strip (e.g. scroll gesture took over). */
  onStripPointerCancel: () => void
  /**
   * Assign a callback to receive double-tap-hold events (second tap held ≥ 400 ms).
   * ReaderContent sets this to open the settings panel.
   */
  doubleTapHoldCallbackRef: MutableRefObject<(() => void) | null>
}

export const Ctx = createContext<InteractiveCtx | null>(null)

export function InteractiveProvider({ children }: { children: ReactNode }) {
  const [scrollPct, setScrollPct] = useState(0)
  const [headerVisible, setHeaderVisible] = useState(true)
  const isTouchingRef = useRef(false)

  // ── Pointer-based double-tap + hold detection ─────────────────────────────
  const lastDownRef = useRef(0)
  const holdTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null)
  const isSecondTapPendingRef = useRef(false)
  const doubleTapHoldCallbackRef = useRef<(() => void) | null>(null)

  // Clear any pending hold timer on unmount.
  useEffect(() => {
    return () => {
      if (holdTimerRef.current !== null) clearTimeout(holdTimerRef.current)
    }
  }, [])

  const onStripPointerDown = useCallback(() => {
    const now = Date.now()
    if (now - lastDownRef.current < 350) {
      // Second tap within window — start hold timer.
      isSecondTapPendingRef.current = true
      holdTimerRef.current = setTimeout(() => {
        holdTimerRef.current = null
        isSecondTapPendingRef.current = false
        lastDownRef.current = 0
        doubleTapHoldCallbackRef.current?.()
      }, 400)
    } else {
      isSecondTapPendingRef.current = false
    }
    lastDownRef.current = now
  }, [])

  const onStripPointerUp = useCallback(() => {
    if (holdTimerRef.current !== null) {
      clearTimeout(holdTimerRef.current)
      holdTimerRef.current = null
      if (isSecondTapPendingRef.current) {
        // Released before hold threshold → quick double-tap, toggle header.
        setHeaderVisible(v => !v)
      }
      isSecondTapPendingRef.current = false
    }
  }, [])

  const onStripPointerCancel = useCallback(() => {
    if (holdTimerRef.current !== null) {
      clearTimeout(holdTimerRef.current)
      holdTimerRef.current = null
    }
    isSecondTapPendingRef.current = false
  }, [])

  // ── Scroll progress ───────────────────────────────────────────────────────
  useEffect(() => {
    const handle = () => {
      const el = document.documentElement
      const total = el.scrollHeight - el.clientHeight
      setScrollPct(total > 0 ? Math.round((el.scrollTop / total) * 100) : 0)
    }
    window.addEventListener('scroll', handle, { passive: true })
    return () => window.removeEventListener('scroll', handle)
  }, [])

  // ── Touch tracking (for autoScroll pause) ────────────────────────────────
  useEffect(() => {
    const onStart = () => { isTouchingRef.current = true }
    const onEnd = () => { isTouchingRef.current = false }
    window.addEventListener('touchstart', onStart, { passive: true })
    window.addEventListener('touchend', onEnd, { passive: true })
    window.addEventListener('touchcancel', onEnd, { passive: true })
    return () => {
      window.removeEventListener('touchstart', onStart)
      window.removeEventListener('touchend', onEnd)
      window.removeEventListener('touchcancel', onEnd)
    }
  }, [])

  return (
    <Ctx.Provider value={{
      scrollPct,
      headerVisible,
      setHeaderVisible,
      isTouchingRef,
      onStripPointerDown,
      onStripPointerUp,
      onStripPointerCancel,
      doubleTapHoldCallbackRef,
    }}>
      {children}
    </Ctx.Provider>
  )
}

export function useInteractive(): InteractiveCtx {
  const ctx = useContext(Ctx)
  if (!ctx) throw new Error('useInteractive must be used within InteractiveProvider')
  return ctx
}
