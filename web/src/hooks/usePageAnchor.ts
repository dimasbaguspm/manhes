import { useState, useEffect, type RefObject } from 'react'
import type { AppChapterRead } from '../types/app'
import { fnv1a } from '../lib/hashCodec'
import { throttle } from '../lib/throttle'

/** Per-page position within the stitched canvas, in canvas pixels. */
export interface CanvasPageLayout {
  top: number
  height: number
}

/** Stable, opaque anchor ID derived from mangaId + chapter + 1-based page number. */
export function pageAnchor(mangaId: string, chapter: string, page: number): string {
  return fnv1a(`${mangaId}--ch${chapter}--p${page}`)
}

export type OverlayState = 'show' | 'fade' | 'gone'

/**
 * Manages scroll-position persistence via the URL hash.
 *
 * Hash format: `#{pageAnchor}.{offsetMilli}`
 *   - pageAnchor  — fnv1a(mangaId--chN--pN), 8-char hex
 *   - offsetMilli — viewport-centre offset within the page, 0–1000 (per-mille)
 *
 * Responsibilities:
 *   1. Restore — on layout arrival, scroll to the position encoded in the hash.
 *   2. Track   — on scroll (throttled to 2/s), keep the hash current.
 *
 * Both operations use canvas-pixel geometry via canvasRef + pageLayout instead
 * of individual DOM element IDs.
 */
export function usePageAnchor(
  data: AppChapterRead | null,
  mangaId: string | undefined,
  chapter: string,
  containerRef: RefObject<HTMLElement> | null,
  pageLayout: CanvasPageLayout[] | null,
): OverlayState {
  const [overlay, setOverlay] = useState<OverlayState>(() =>
    window.location.hash.length > 1 ? 'show' : 'gone',
  )

  function dismissOverlay() {
    setOverlay('fade')
    setTimeout(() => setOverlay('gone'), 300)
  }

  // ── 1. Restore ────────────────────────────────────────────────────────────
  useEffect(() => {
    if (!data || !mangaId || !containerRef?.current || !pageLayout) return
    const raw = window.location.hash.slice(1)
    if (!raw) { dismissOverlay(); return }

    const dot = raw.indexOf('.')
    const hashPart = dot !== -1 ? raw.slice(0, dot) : raw
    const offsetMilli = dot !== -1 ? parseInt(raw.slice(dot + 1), 10) : 0
    const offset = isNaN(offsetMilli) ? 0 : offsetMilli / 1000

    const pageIdx = data.pages.findIndex((_, i) => pageAnchor(mangaId, chapter, i + 1) === hashPart)

    if (pageIdx !== -1 && pageLayout[pageIdx]) {
      const t = setTimeout(() => {
        const container = containerRef.current
        if (!container) { dismissOverlay(); return }
        // pageLayout is in CSS pixels — no canvas-pixel scale conversion needed.
        const containerTop = container.getBoundingClientRect().top + window.scrollY
        const pg = pageLayout[pageIdx]
        const target = containerTop + pg.top + pg.height * offset - window.innerHeight / 2
        window.scrollTo({ top: Math.max(0, target) })
        dismissOverlay()
      }, 80)
      return () => clearTimeout(t)
    } else {
      dismissOverlay()
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [data, containerRef, pageLayout])

  // ── 2. Track ──────────────────────────────────────────────────────────────
  // Chrome throttles history.replaceState after ~100 calls/30 s (≈3.3/s).
  // throttle() caps writes at 2/s (500 ms) — well under the limit.
  useEffect(() => {
    if (!data || !mangaId || !containerRef?.current || !pageLayout) return

    let rafId: number
    const writeHash = throttle(
      (url: string) => history.replaceState(null, '', url),
      500,
    )

    function update() {
      const container = containerRef!.current
      if (!container || !pageLayout) return
      // pageLayout is in CSS pixels — no canvas-pixel scale conversion needed.
      const containerTop = container.getBoundingClientRect().top + window.scrollY
      const viewCenter = window.scrollY + window.innerHeight / 2

      for (let i = 0; i < pageLayout.length; i++) {
        const pg = pageLayout[i]
        const top = containerTop + pg.top
        const height = pg.height
        if (height > 0 && viewCenter >= top && viewCenter < top + height) {
          const offsetMilli = Math.round(((viewCenter - top) / height) * 1000)
          writeHash(`#${pageAnchor(mangaId!, chapter, i + 1)}.${offsetMilli}`)
          break
        }
      }
    }

    function onScroll() {
      cancelAnimationFrame(rafId)
      rafId = requestAnimationFrame(update)
    }

    window.addEventListener('scroll', onScroll, { passive: true })
    update()

    return () => {
      window.removeEventListener('scroll', onScroll)
      cancelAnimationFrame(rafId)
      history.replaceState(null, '', location.pathname + location.search)
    }
  }, [data, mangaId, chapter, containerRef, pageLayout])

  return overlay
}
