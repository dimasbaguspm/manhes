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
  /**
   * Call this from any tap/click handler on the manga strip.
   * Two taps within 350 ms toggle header visibility.
   */
  onStripTap: () => void
}

const Ctx = createContext<InteractiveCtx | null>(null)

export function InteractiveProvider({ children }: { children: ReactNode }) {
  const [scrollPct, setScrollPct] = useState(0)
  const [headerVisible, setHeaderVisible] = useState(true)
  const isTouchingRef = useRef(false)
  const lastTapRef = useRef(0)

  useEffect(() => {
    const handle = () => {
      const el = document.documentElement
      const total = el.scrollHeight - el.clientHeight
      setScrollPct(total > 0 ? Math.round((el.scrollTop / total) * 100) : 0)
    }
    window.addEventListener('scroll', handle, { passive: true })
    return () => window.removeEventListener('scroll', handle)
  }, [])

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

  const onStripTap = useCallback(() => {
    const now = Date.now()
    if (now - lastTapRef.current < 350) {
      setHeaderVisible(v => !v)
      lastTapRef.current = 0
    } else {
      lastTapRef.current = now
    }
  }, [])

  return (
    <Ctx.Provider value={{ scrollPct, headerVisible, setHeaderVisible, isTouchingRef, onStripTap }}>
      {children}
    </Ctx.Provider>
  )
}

export function useInteractive(): InteractiveCtx {
  const ctx = useContext(Ctx)
  if (!ctx) throw new Error('useInteractive must be used within InteractiveProvider')
  return ctx
}
