import { useState, useEffect } from 'react'
import type { AppChapterRead } from '../types/app'
import { fnv1a } from '../lib/hashCodec'
import { throttle } from '../lib/throttle'


/** Stable, opaque anchor ID derived from mangaId + chapter + page number. */
export function pageAnchor(mangaId: string, chapter: number, page: number): string {
  return fnv1a(`${mangaId}--ch${chapter}--p${page}`)
}

export type OverlayState = 'show' | 'fade' | 'gone'

/**
 * Manages scroll-position persistence via the URL hash.
 *
 * Hash format: `#{pageAnchor}.{offsetMilli}`
 *   - pageAnchor  — fnv1a(mangaId--chN--pN), 8-char hex
 *   - offsetMilli — viewport-centre offset within the image, 0–1000 (per-mille)
 *
 * Reader settings are persisted independently through localStorage
 * (see usePersistedState / useReaderSettings) and are not part of the hash.
 *
 * Responsibilities:
 *   1. Restore — on data arrival, scroll to the position encoded in the hash.
 *   2. Track   — on scroll (throttled to 2/s), keep the hash current.
 */
export function usePageAnchor(
  data: AppChapterRead | null,
  mangaId: string | undefined,
  chapter: number,
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
    if (!data || !mangaId) return
    const raw = window.location.hash.slice(1)
    if (!raw) { dismissOverlay(); return }

    const dot         = raw.indexOf('.')
    const hashPart    = dot !== -1 ? raw.slice(0, dot) : raw
    const offsetMilli = dot !== -1 ? parseInt(raw.slice(dot + 1), 10) : 0
    const offset      = isNaN(offsetMilli) ? 0 : offsetMilli / 1000

    const pageIdx = data.pages.findIndex((_, i) => pageAnchor(mangaId, chapter, i + 1) === hashPart)
    if (pageIdx !== -1) {
      const t = setTimeout(() => {
        const el = document.getElementById(hashPart)
        if (el) {
          const top    = el.getBoundingClientRect().top + window.scrollY
          const target = top + el.offsetHeight * offset - window.innerHeight / 2
          window.scrollTo({ top: Math.max(0, target) })
        }
        dismissOverlay()
      }, 80)
      return () => clearTimeout(t)
    } else {
      dismissOverlay()
    }
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [data])

  // ── 2. Track ──────────────────────────────────────────────────────────────
  // Chrome throttles history.replaceState after ~100 calls/30 s (≈3.3/s).
  // throttle() caps writes at 2/s (500 ms) — well under the limit.
  useEffect(() => {
    if (!data || !mangaId) return

    let rafId: number
    const writeHash = throttle(
      (url: string) => history.replaceState(null, '', url),
      500,
    )

    function update() {
      const viewCenter = window.scrollY + window.innerHeight / 2
      for (let i = 0; i < data!.pages.length; i++) {
        const id = pageAnchor(mangaId!, chapter, i + 1)
        const el = document.getElementById(id)
        if (!el) continue
        const top    = el.getBoundingClientRect().top + window.scrollY
        const height = el.offsetHeight
        if (height > 0 && viewCenter >= top && viewCenter < top + height) {
          const offsetMilli = Math.round(((viewCenter - top) / height) * 1000)
          writeHash(`#${id}.${offsetMilli}`)
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
  }, [data, mangaId, chapter])

  return overlay
}
