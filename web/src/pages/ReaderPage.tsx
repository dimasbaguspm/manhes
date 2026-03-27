import { useState, useRef, useEffect, useCallback } from 'react'
import { useMangaReader } from '../providers/MangaReaderDataProvider'
import { usePageAnchor, type CanvasPageLayout } from '../hooks/usePageAnchor'
import { useProgressSave } from '../hooks/useProgressSave'
import { DEEP_LINKS } from '../lib/deepLinks'
import { InteractiveProvider, useInteractive } from '../components/reader/InteractiveProvider'
import { useReaderSettings } from '../components/reader/useReaderSettings'
import { useAutoScroll } from '../hooks/useAutoScroll'
import { ReaderHeader } from '../components/reader/ReaderHeader'
import { ReaderMenu } from '../components/reader/ReaderMenu'
import { ReaderSettingsPanel } from '../components/reader/ReaderSettings'
import { ReaderStrip } from '../components/reader/ReaderStrip'
import { ReaderProgressBar } from '../components/reader/ReaderProgressBar'
import { ReaderCanvas, type CanvasLoadingInfo } from '../components/reader/ReaderCanvas'
import { ChapterNavFooter } from '../components/reader/ChapterNavFooter'
import { ShortcutsOverlay } from '../components/reader/ShortcutsOverlay'

type OverlayState = 'visible' | 'fade' | 'gone'

// ── Inner content — requires InteractiveProvider in the tree ──────────────────

function ReaderContent() {
  const { data, loading, error, chapter, mangaId, lang, goNext, goPrev } = useMangaReader()
  const { settings, set, stripMaxWidthClass, bgClass } = useReaderSettings()
  const {
    scrollPct,
    headerVisible,
    setHeaderVisible,
    isTouchingRef,
    onStripPointerDown,
    onStripPointerUp,
    onStripPointerCancel,
    doubleTapHoldCallbackRef,
  } = useInteractive()

  const [menuOpen, setMenuOpen] = useState(false)
  const [shortcutsOpen, setShortcutsOpen] = useState(false)

  const canvasRef = useRef<HTMLCanvasElement>(null)
  const [pageLayout, setPageLayout] = useState<CanvasPageLayout[] | null>(null)

  // ── Canvas loading overlay ─────────────────────────────────────────────────
  // Shown as a full-viewport dim screen while the worker fetches images.
  // Transitions: visible → fade (opacity-0, 300 ms) → gone (unmounted).
  const [canvasOverlay, setCanvasOverlay] = useState<OverlayState>('gone')
  const [canvasLoadInfo, setCanvasLoadInfo] = useState({ loaded: 0, total: 0 })
  const fadeTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null)

  // When a new chapter's data arrives, show the overlay before the canvas mounts.
  const lastDataChapterRef = useRef<string | null>(null)
  useEffect(() => {
    if (!data) return
    if (chapter !== lastDataChapterRef.current) {
      lastDataChapterRef.current = chapter
      if (fadeTimerRef.current !== null) clearTimeout(fadeTimerRef.current)
      setCanvasOverlay('visible')
      setCanvasLoadInfo({ loaded: 0, total: 0 })
    }
  }, [data, chapter])

  const handleCanvasLoadingState = useCallback((info: CanvasLoadingInfo) => {
    setCanvasLoadInfo({ loaded: info.loaded, total: info.total })
    if (!info.loading) {
      setCanvasOverlay('fade')
      fadeTimerRef.current = setTimeout(() => {
        setCanvasOverlay('gone')
        fadeTimerRef.current = null
      }, 300)
    }
  }, [])

  // Clear the fade timer on unmount.
  useEffect(() => {
    return () => {
      if (fadeTimerRef.current !== null) clearTimeout(fadeTimerRef.current)
    }
  }, [])

  // Reset layout whenever the chapter changes so the anchor hook doesn't use
  // stale geometry while the new canvas is being drawn.
  useEffect(() => {
    setPageLayout(null)
  }, [chapter])

  useProgressSave(mangaId, lang, chapter, scrollPct)

  // Double-tap-hold opens settings.
  useEffect(() => {
    doubleTapHoldCallbackRef.current = () => setMenuOpen(true)
    return () => { doubleTapHoldCallbackRef.current = null }
  }, [doubleTapHoldCallbackRef])

  // Keyboard shortcuts: f = fullscreen, s = settings, / = shortcuts, Esc = close.
  useEffect(() => {
    const handler = (e: KeyboardEvent) => {
      const tag = (e.target as HTMLElement)?.tagName
      if (tag === 'INPUT' || tag === 'TEXTAREA' || tag === 'SELECT') return

      if (e.key === 'f' || e.key === 'F') {
        if (!document.fullscreenElement) {
          document.documentElement.requestFullscreen?.()
        } else {
          document.exitFullscreen?.()
        }
      }
      if (e.key === 's' || e.key === 'S') setMenuOpen(v => !v)
      if (e.key === '/') {
        e.preventDefault()
        setShortcutsOpen(v => !v)
      }
      if (e.key === 'Escape') {
        setShortcutsOpen(false)
        setMenuOpen(false)
      }
    }
    window.addEventListener('keydown', handler)
    return () => window.removeEventListener('keydown', handler)
  }, [])

  const scrollRestoreOverlay = usePageAnchor(data, mangaId, chapter, canvasRef, pageLayout)

  useAutoScroll(
    settings.autoScroll,
    settings.autoScrollSpeed,
    isTouchingRef,
    () => set('autoScroll', false),
  )

  const chaptersHref = DEEP_LINKS.MANGA_CHAPTERS({ mangaId: mangaId!, lang: lang! })

  return (
    <div className={`min-h-screen ${bgClass}`}>

      {/* Worker loading overlay — dim screen with centered progress while images load */}
      {canvasOverlay !== 'gone' && (
        <div
          className={`fixed inset-0 z-40 flex flex-col items-center justify-center gap-3 bg-gray-950 transition-opacity duration-300 ${
            canvasOverlay === 'fade' ? 'pointer-events-none opacity-0' : 'opacity-100'
          }`}
        >
          <SpinnerIcon className="h-7 w-7 animate-spin text-indigo-500" />
          <p className="text-sm text-gray-500">
            {canvasLoadInfo.total > 0
              ? canvasLoadInfo.loaded > 0
                ? `Loading ${canvasLoadInfo.loaded} / ${canvasLoadInfo.total} page${canvasLoadInfo.total !== 1 ? 's' : ''}…`
                : `Loading ${canvasLoadInfo.total} page${canvasLoadInfo.total !== 1 ? 's' : ''}…`
              : 'Loading…'}
          </p>
        </div>
      )}

      {/* Scroll-restore overlay — fades out once the target page is scrolled to */}
      {scrollRestoreOverlay !== 'gone' && (
        <div
          className={`fixed inset-0 z-50 bg-gray-950 transition-opacity duration-300 ${
            scrollRestoreOverlay === 'fade' ? 'opacity-0' : 'opacity-100'
          }`}
        />
      )}

      <ReaderHeader
        visible={headerVisible}
        lang={lang!}
        chapter={chapter}
        pageCount={data?.pages.length ?? null}
        chaptersHref={chaptersHref}
        menuOpen={menuOpen}
        onMenuToggle={() => setMenuOpen(o => !o)}
        onShortcutsToggle={() => setShortcutsOpen(o => !o)}
        prevDisabled={data?.prevChapter == null}
        nextDisabled={data?.nextChapter == null}
        onPrev={goPrev}
        onNext={goNext}
      />

      <ReaderMenu open={menuOpen} onClose={() => setMenuOpen(false)}>
        <ReaderSettingsPanel
          settings={settings}
          set={set}
          headerVisible={headerVisible}
          onHeaderToggle={() => setHeaderVisible(v => !v)}
          prevDisabled={data?.prevChapter == null}
          nextDisabled={data?.nextChapter == null}
          onPrev={goPrev}
          onNext={goNext}
        />
      </ReaderMenu>

      <ShortcutsOverlay open={shortcutsOpen} onClose={() => setShortcutsOpen(false)} />

      <ReaderStrip
        maxWidthClass={stripMaxWidthClass}
        onPointerDown={onStripPointerDown}
        onPointerUp={onStripPointerUp}
        onPointerCancel={onStripPointerCancel}
        renderLoading={loading ? () => (
          <div className="flex h-96 items-center justify-center text-gray-500">
            Loading chapter…
          </div>
        ) : undefined}
        renderError={error ? () => (
          <div className="m-6 rounded-lg border border-red-800 bg-red-950 px-4 py-3 text-sm text-red-300">
            {error}
          </div>
        ) : undefined}
        renderPages={data ? () => (
          <ReaderCanvas
            key={chapter}
            urls={data.pages}
            canvasRef={canvasRef}
            onLayout={setPageLayout}
            onLoadingState={handleCanvasLoadingState}
          />
        ) : undefined}
        // Footer is suppressed while the loading overlay is covering the screen.
        renderFooter={data && canvasOverlay === 'gone' ? () => (
          <ChapterNavFooter
            chaptersHref={chaptersHref}
            prevDisabled={data.prevChapter == null}
            nextDisabled={data.nextChapter == null}
            onPrev={goPrev}
            onNext={goNext}
          />
        ) : undefined}
      />

      <ReaderProgressBar
        pct={scrollPct}
        showBar={settings.showProgress}
        showIndicator={settings.showPageIndicator}
      />

    </div>
  )
}

// ── Page export — InteractiveProvider wraps the content tree ─────────────────

export default function ReaderPage() {
  return (
    <InteractiveProvider>
      <ReaderContent />
    </InteractiveProvider>
  )
}

// ── Shared ────────────────────────────────────────────────────────────────────

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
