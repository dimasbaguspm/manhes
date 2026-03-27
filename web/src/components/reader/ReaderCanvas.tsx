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
  /** Forwarded ref — the parent holds this so anchor tracking can read canvas geometry. */
  canvasRef: RefObject<HTMLCanvasElement>
  onLayout: (layout: CanvasPageLayout[]) => void
  /**
   * Called whenever loading state changes. Parent is responsible for showing
   * a loading overlay; this component only renders the canvas and inline errors.
   */
  onLoadingState?: (info: CanvasLoadingInfo) => void
}

/**
 * Loads all page images off the main thread via a Web Worker, then stitches
 * them into a single canvas. The worker fetches and decodes each image as an
 * ImageBitmap, transferring ownership back to the main thread so the CPU-heavy
 * decode never blocks the UI.
 *
 * Loading progress and state are reported upward via onLoadingState — the
 * parent owns the loading overlay so it can position it anywhere in the tree.
 *
 * Drawing remains lazy: the first few pages are painted immediately, the rest
 * are painted on demand as the user scrolls near them.
 */
export function ReaderCanvas({ urls, canvasRef, onLayout, onLoadingState }: ReaderCanvasProps) {
  const [status, setStatus] = useState<'loading' | 'ready' | 'error'>('loading')
  const [errorMsg, setErrorMsg] = useState<string | null>(null)

  // Always-fresh ref so the worker handler never uses a stale callback.
  const onLoadingStateRef = useRef(onLoadingState)
  useEffect(() => { onLoadingStateRef.current = onLoadingState })

  useEffect(() => {
    if (!urls.length) return
    setStatus('loading')
    setErrorMsg(null)
    onLoadingStateRef.current?.({ loading: true, loaded: 0, total: urls.length })

    let cancelled = false
    let rafId = 0
    let cleanupScroll: (() => void) | null = null

    const bitmaps = new Map<number, ImageBitmap>()
    const errors = new Set<number>()

    const worker = new Worker(
      new URL('../../workers/imageLoader.worker.ts', import.meta.url),
      { type: 'module' },
    )

    worker.onmessage = (e: MessageEvent<WorkerOutMessage>) => {
      if (cancelled) return
      const msg = e.data

      if (msg.type === 'page') {
        bitmaps.set(msg.index, msg.bitmap)
        return
      }

      if (msg.type === 'error') {
        errors.add(msg.index)
        return
      }

      if (msg.type === 'progress') {
        onLoadingStateRef.current?.({ loading: true, loaded: msg.loaded, total: msg.total })
        return
      }

      if (msg.type === 'done') {
        if (errors.size === urls.length) {
          setErrorMsg('Failed to load any pages')
          setStatus('error')
          onLoadingStateRef.current?.({ loading: false, loaded: urls.length, total: urls.length })
          return
        }

        const canvas = canvasRef.current
        if (!canvas) return

        const MAX_WIDTH = 1200
        const MAX_HEIGHT = 32000

        const rawWidth = Math.min(
          Math.max(...[...bitmaps.values()].map(b => b.width), 800),
          MAX_WIDTH,
        )

        const rawHeights = urls.map((_, i) => {
          const b = bitmaps.get(i)
          if (!b || b.width === 0) return 0
          return Math.round(b.height * (rawWidth / b.width))
        })
        const rawTotal = rawHeights.reduce((a, b) => a + b, 0)

        const scale = rawTotal > MAX_HEIGHT ? MAX_HEIGHT / rawTotal : 1
        const canvasWidth = Math.round(rawWidth * scale)

        const layout: CanvasPageLayout[] = []
        let totalHeight = 0
        for (let i = 0; i < urls.length; i++) {
          const h = Math.round(rawHeights[i] * scale)
          layout.push({ top: totalHeight, height: h })
          totalHeight += h
        }

        canvas.width = canvasWidth
        canvas.height = totalHeight

        const ctx = canvas.getContext('2d')
        if (!ctx) {
          setErrorMsg('Could not get canvas context')
          setStatus('error')
          onLoadingStateRef.current?.({ loading: false, loaded: urls.length, total: urls.length })
          return
        }

        // Track painted pages — never redraw an already-drawn page.
        const drawn = new Set<number>()

        // Pre-paint the first few pages synchronously so there is immediate
        // content when the overlay fades out.
        const PREFETCH = Math.min(3, urls.length)
        for (let i = 0; i < PREFETCH; i++) {
          const bitmap = bitmaps.get(i)
          if (bitmap) {
            ctx.drawImage(bitmap, 0, layout[i].top, canvasWidth, layout[i].height)
            drawn.add(i)
          }
        }

        onLayout(layout)
        setStatus('ready')
        onLoadingStateRef.current?.({ loading: false, loaded: urls.length, total: urls.length })

        // Paint every undrawn page whose canvas region overlaps the viewport ± buffer.
        function drawVisible() {
          if (cancelled || !canvas || !ctx) return
          const scaleF = canvas.width > 0 ? canvas.offsetWidth / canvas.width : 1
          const canvasTopAbs = canvas.getBoundingClientRect().top + window.scrollY
          const buffer = window.innerHeight
          const vpTop = window.scrollY - buffer
          const vpBottom = window.scrollY + window.innerHeight + buffer

          for (let i = 0; i < layout.length; i++) {
            if (drawn.has(i)) continue
            const bitmap = bitmaps.get(i)
            if (!bitmap) continue
            const pageTop = canvasTopAbs + layout[i].top * scaleF
            const pageBottom = pageTop + layout[i].height * scaleF
            if (pageBottom >= vpTop && pageTop <= vpBottom) {
              ctx.drawImage(bitmap, 0, layout[i].top, canvasWidth, layout[i].height)
              drawn.add(i)
            }
          }
        }

        rafId = requestAnimationFrame(drawVisible)

        const onScroll = () => {
          cancelAnimationFrame(rafId)
          rafId = requestAnimationFrame(drawVisible)
        }
        window.addEventListener('scroll', onScroll, { passive: true })
        cleanupScroll = () => {
          window.removeEventListener('scroll', onScroll)
          cancelAnimationFrame(rafId)
        }
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
      cleanupScroll?.()
      cancelAnimationFrame(rafId)
      worker.terminate()
      bitmaps.forEach(b => b.close())
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [urls])

  return (
    <>
      {status === 'error' && (
        <div className="m-6 rounded-lg border border-red-800 bg-red-950 px-4 py-3 text-sm text-red-300">
          {errorMsg}
        </div>
      )}
      <canvas
        ref={canvasRef}
        className="block w-full"
        style={{ display: status === 'ready' ? 'block' : 'none' }}
      />
    </>
  )
}
