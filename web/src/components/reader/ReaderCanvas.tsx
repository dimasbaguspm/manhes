import { useEffect, useState, type RefObject } from 'react'
import type { CanvasPageLayout } from '../../hooks/usePageAnchor'

interface ReaderCanvasProps {
  urls: string[]
  /** Forwarded ref — the parent holds this so anchor tracking can read canvas geometry. */
  canvasRef: RefObject<HTMLCanvasElement>
  onLayout: (layout: CanvasPageLayout[]) => void
}

/**
 * Loads all page images, stitches them into a single canvas, then calls
 * onLayout with per-page pixel positions so the anchor hook can scroll
 * without needing individual DOM elements.
 *
 * Drawing is lazy: the first few pages are painted immediately so content
 * appears at once; the remaining pages are painted on demand as the user
 * scrolls them into (or near) the viewport, avoiding a full-canvas draw
 * that would hang the main thread.
 */
export function ReaderCanvas({ urls, canvasRef, onLayout }: ReaderCanvasProps) {
  const [status, setStatus] = useState<'loading' | 'ready' | 'error'>('loading')
  const [errorMsg, setErrorMsg] = useState<string | null>(null)

  useEffect(() => {
    if (!urls.length) return
    setStatus('loading')
    setErrorMsg(null)

    let cancelled = false
    let rafId = 0
    let cleanupScroll: (() => void) | null = null

    const load = (url: string) =>
      new Promise<HTMLImageElement>((resolve, reject) => {
        const img = new Image()
        img.onload = () => resolve(img)
        img.onerror = () => reject(new Error(`Failed to load page`))
        img.src = url
      })

    Promise.all(urls.map(load))
      .then(loaded => {
        if (cancelled) return
        const canvas = canvasRef.current
        if (!canvas) return

        // Chrome's GPU texture limit is ~32 767 px in a single dimension.
        // First compute page heights at the desired width, then scale the whole
        // canvas down if the total height would exceed the safe threshold.
        const MAX_WIDTH = 1200
        const MAX_HEIGHT = 32000

        const rawWidth = Math.min(
          Math.max(...loaded.map(img => img.naturalWidth), 800),
          MAX_WIDTH,
        )
        const rawHeights = loaded.map(img =>
          img.naturalWidth > 0
            ? Math.round(img.naturalHeight * (rawWidth / img.naturalWidth))
            : 0,
        )
        const rawTotal = rawHeights.reduce((a, b) => a + b, 0)

        // Uniform scale-down when height would overflow the GPU limit.
        const scale = rawTotal > MAX_HEIGHT ? MAX_HEIGHT / rawTotal : 1
        const canvasWidth = Math.round(rawWidth * scale)

        const layout: CanvasPageLayout[] = []
        let totalHeight = 0
        for (let i = 0; i < loaded.length; i++) {
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
          return
        }

        // ── Lazy drawing ──────────────────────────────────────────────────────
        // Track which pages have already been painted so we never redraw them.
        const drawn = new Set<number>()

        // Pre-draw: paint the first few pages synchronously before the canvas
        // is shown. Drawing 3 pages is fast (~ms) and ensures there is
        // immediate content visible — no blank canvas flash.
        const PREFETCH = Math.min(3, loaded.length)
        for (let i = 0; i < PREFETCH; i++) {
          ctx.drawImage(loaded[i], 0, layout[i].top, canvasWidth, layout[i].height)
          drawn.add(i)
        }

        onLayout(layout)
        setStatus('ready')

        // drawVisible paints every undrawn page whose canvas-pixel region
        // overlaps the current viewport ± one viewport of buffer.
        // getBoundingClientRect works once the canvas is display:block
        // (after React commits the setStatus('ready') update).
        function drawVisible() {
          if (cancelled || !canvas || !ctx) return
          const scaleF = canvas.width > 0 ? canvas.offsetWidth / canvas.width : 1
          const canvasTopAbs = canvas.getBoundingClientRect().top + window.scrollY
          const buffer = window.innerHeight
          const vpTop = window.scrollY - buffer
          const vpBottom = window.scrollY + window.innerHeight + buffer

          for (let i = 0; i < layout.length; i++) {
            if (drawn.has(i)) continue
            const pageTop = canvasTopAbs + layout[i].top * scaleF
            const pageBottom = pageTop + layout[i].height * scaleF
            if (pageBottom >= vpTop && pageTop <= vpBottom) {
              ctx.drawImage(loaded[i], 0, layout[i].top, canvasWidth, layout[i].height)
              drawn.add(i)
            }
          }
        }

        // Defer initial drawVisible until after React commits (canvas visible).
        rafId = requestAnimationFrame(drawVisible)

        // Draw more pages as the user scrolls.
        const onScroll = () => {
          cancelAnimationFrame(rafId)
          rafId = requestAnimationFrame(drawVisible)
        }
        window.addEventListener('scroll', onScroll, { passive: true })
        cleanupScroll = () => {
          window.removeEventListener('scroll', onScroll)
          cancelAnimationFrame(rafId)
        }
      })
      .catch(err => {
        if (cancelled) return
        setErrorMsg(String(err?.message ?? err))
        setStatus('error')
      })

    return () => {
      cancelled = true
      cleanupScroll?.()
      cancelAnimationFrame(rafId)
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [urls])

  return (
    <>
      {status === 'loading' && (
        <div className="flex flex-col items-center justify-center gap-3 py-24">
          <SpinnerIcon className="h-7 w-7 animate-spin text-indigo-500" />
          <p className="text-sm text-gray-500">
            Loading {urls.length} page{urls.length !== 1 ? 's' : ''}…
          </p>
        </div>
      )}
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

function SpinnerIcon({ className }: { className?: string }) {
  return (
    <svg className={className} fill="none" viewBox="0 0 24 24">
      <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" />
      <path
        className="opacity-75"
        fill="currentColor"
        d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"
      />
    </svg>
  )
}
