import { useEffect, useRef, useState, type RefObject } from 'react'
import type { CanvasPageLayout } from '../../hooks/usePageAnchor'
import type { WorkerOutMessage } from '../../workers/imageLoader.worker'

export interface CanvasLoadingInfo {
  loading: boolean
  loaded: number
  total: number
}

interface ReaderCanvasProps {
  urls: string[]
  /** Forwarded ref — the parent holds this so anchor tracking can read container geometry. */
  containerRef: RefObject<HTMLDivElement>
  onLayout: (layout: CanvasPageLayout[]) => void
  /**
   * Called whenever loading state changes. Parent is responsible for showing
   * a loading overlay; this component only renders the pages and inline errors.
   */
  onLoadingState?: (info: CanvasLoadingInfo) => void
}

/**
 * Loads all page images off the main thread via a Web Worker, then renders
 * each page as a native <img> element.
 *
 * The worker fetches each URL and transfers the raw ArrayBuffer (zero-copy)
 * back to the main thread. The main thread creates a blob URL for each buffer
 * and sets it as the img src. Native <img> rendering preserves full source
 * resolution with proper HiDPI support — no canvas scaling artifacts.
 *
 * Loading progress and state are reported upward via onLoadingState.
 */
export function ReaderCanvas({ urls, containerRef, onLayout, onLoadingState }: ReaderCanvasProps) {
  const [pageUrls, setPageUrls] = useState<(string | null)[]>(() => new Array(urls.length).fill(null))
  const [status, setStatus] = useState<'loading' | 'ready' | 'error'>('loading')
  const [errorMsg, setErrorMsg] = useState<string | null>(null)

  // Always-fresh refs so callbacks never use stale values.
  const onLoadingStateRef = useRef(onLoadingState)
  useEffect(() => { onLoadingStateRef.current = onLoadingState })

  const onLayoutRef = useRef(onLayout)
  useEffect(() => { onLayoutRef.current = onLayout })

  // Per-page img element refs — populated by React during commit.
  const imgRefs = useRef<(HTMLImageElement | null)[]>([])

  // settleRef is the img onLoad/onError callback. Cleared in cleanup so events
  // from a previous chapter (blob URLs being revoked → img errors) never corrupt
  // the new chapter's settled/fail counters.
  const settleRef = useRef<((fail: boolean) => void) | null>(null)

  useEffect(() => {
    if (!urls.length) return

    setStatus('loading')
    setErrorMsg(null)
    setPageUrls(new Array(urls.length).fill(null))
    imgRefs.current = new Array(urls.length).fill(null)
    onLoadingStateRef.current?.({ loading: true, loaded: 0, total: urls.length })

    let cancelled = false
    const createdUrls: string[] = []
    let workerDone = false
    let settled = 0
    let failCount = 0

    function checkDone() {
      if (!workerDone || settled < urls.length) return

      if (failCount >= urls.length) {
        setErrorMsg('Failed to load any pages')
        setStatus('error')
        onLoadingStateRef.current?.({ loading: false, loaded: urls.length, total: urls.length })
        return
      }

      // Compute layout from rendered img positions (CSS pixels — no scale needed).
      const layout: CanvasPageLayout[] = []
      let top = 0
      for (let i = 0; i < urls.length; i++) {
        const h = imgRefs.current[i]?.offsetHeight ?? 0
        layout.push({ top, height: h })
        top += h
      }
      onLayoutRef.current(layout)
      setStatus('ready')
      onLoadingStateRef.current?.({ loading: false, loaded: urls.length, total: urls.length })
    }

    settleRef.current = (fail: boolean) => {
      settled++
      if (fail) failCount++
      checkDone()
    }

    const worker = new Worker(
      new URL('../../workers/imageLoader.worker.ts', import.meta.url),
      { type: 'module' },
    )

    worker.onmessage = (e: MessageEvent<WorkerOutMessage>) => {
      if (cancelled) return
      const msg = e.data

      if (msg.type === 'page') {
        const blob = new Blob([msg.buffer], { type: msg.mime })
        const url = URL.createObjectURL(blob)
        createdUrls.push(url)
        setPageUrls(prev => {
          const next = [...prev]
          next[msg.index] = url
          return next
        })
        return
      }

      if (msg.type === 'error') {
        // No img element is created for this page — count it settled immediately.
        settled++
        failCount++
        checkDone()
        return
      }

      if (msg.type === 'progress') {
        onLoadingStateRef.current?.({ loading: true, loaded: msg.loaded, total: msg.total })
        return
      }

      if (msg.type === 'done') {
        workerDone = true
        checkDone()
      }
    }

    worker.onerror = (e) => {
      if (cancelled) return
      setErrorMsg(e.message ?? 'Worker error')
      setStatus('error')
      onLoadingStateRef.current?.({ loading: false, loaded: 0, total: urls.length })
    }

    worker.postMessage({ urls })

    return () => {
      cancelled = true
      settleRef.current = null
      worker.terminate()
      createdUrls.forEach(u => URL.revokeObjectURL(u))
    }
  }, [urls])

  return (
    <>
      {status === 'error' && (
        <div className="m-6 rounded-lg border border-red-800 bg-red-950 px-4 py-3 text-sm text-red-300">
          {errorMsg}
        </div>
      )}
      {/* visibility:hidden keeps imgs in layout flow (so offsetHeight is accurate)
          while the loading overlay covers the screen. */}
      <div
        ref={containerRef}
        style={{ visibility: status === 'ready' ? 'visible' : 'hidden' }}
      >
        {pageUrls.map((url, i) =>
          url ? (
            <img
              key={i}
              ref={el => { imgRefs.current[i] = el }}
              src={url}
              className="block w-full"
              alt={`Page ${i + 1}`}
              draggable={false}
              onLoad={() => settleRef.current?.(false)}
              onError={() => settleRef.current?.(true)}
            />
          ) : null,
        )}
      </div>
    </>
  )
}
