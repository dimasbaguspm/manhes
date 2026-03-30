import { Icon } from '@/components/icon'
import { useApiChapterRead } from '@/hooks/use-api-chapter-read'
import { useAutoScroll } from '@/hooks/use-auto-scroll'
import { useChapterName } from '@/hooks/use-chapter-name'
import { usePageAnchor, type CanvasPageLayout } from '@/hooks/use-page-anchor'
import { useProgressSave } from '@/hooks/use-progress-save'
import { DEEP_LINKS } from '@/lib/deep-links'
import { parseChapterIdFromUrl } from '@/lib/format-data'
import { AutoScrollControls, ChapterNavFooter, InteractiveProvider, ReaderCanvas, ReaderHeader, ReaderMenu, ReaderProgressBar, ReaderSettingsPanel, ReaderStrip, ShortcutsOverlay, useInteractive, useReaderSettings, type CanvasLoadingInfo } from '@/pages/reader-page/components'
import { Loader2 } from 'lucide-react'
import { useCallback, useEffect, useRef, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'

type OverlayState = 'visible' | 'fade' | 'gone'

function ReaderContent() {
  const { chapterId } = useParams<{ chapterId: string }>()
  const navigate = useNavigate()
  const { data, loading, error } = useApiChapterRead(chapterId)
  const { name: chapterName, loading: chapterNameLoading } = useChapterName(data?.manga_id, chapterId)
  const { settings, set, stripMaxWidthClass, bgClass, isActive: autoScrollActive, speed: autoScrollSpeed, toggle: toggleAutoScroll, cycleSpeed } = useReaderSettings()
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

  const containerRef = useRef<HTMLDivElement>(null)
  const [pageLayout, setPageLayout] = useState<CanvasPageLayout[] | null>(null)

  const goNext = useCallback(() => {
    const next = parseChapterIdFromUrl(data?.next_chapter)
    if (next) navigate(`/read/${encodeURIComponent(next)}`)
  }, [data?.next_chapter, navigate])

  const goPrev = useCallback(() => {
    const prev = parseChapterIdFromUrl(data?.prev_chapter)
    if (prev) navigate(`/read/${encodeURIComponent(prev)}`)
  }, [data?.prev_chapter, navigate])

  // ── Canvas loading overlay ─────────────────────────────────────────────────
  const [canvasOverlay, setCanvasOverlay] = useState<OverlayState>('gone')
  const [canvasLoadInfo, setCanvasLoadInfo] = useState({ loaded: 0, total: 0 })
  const fadeTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null)

  const lastDataChapterRef = useRef<string | null>(null)
  useEffect(() => {
    if (!data) return
    if (chapterId !== lastDataChapterRef.current) {
      lastDataChapterRef.current = chapterId ?? null
      if (fadeTimerRef.current !== null) clearTimeout(fadeTimerRef.current)
      setCanvasOverlay('visible')
      setCanvasLoadInfo({ loaded: 0, total: 0 })
    }
  }, [data, chapterId])

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

  useEffect(() => {
    return () => {
      if (fadeTimerRef.current !== null) clearTimeout(fadeTimerRef.current)
    }
  }, [])

  // Reset layout whenever the chapter changes so the anchor hook doesn't use
  // stale geometry while the new canvas is being drawn.
  useEffect(() => {
    setPageLayout(null)
  }, [chapterId])

  useProgressSave(data?.manga_id, chapterId, chapterId, scrollPct)

  // Double-tap-hold opens settings.
  useEffect(() => {
    doubleTapHoldCallbackRef.current = () => setMenuOpen(true)
    return () => { doubleTapHoldCallbackRef.current = null }
  }, [doubleTapHoldCallbackRef])

  // Keyboard shortcuts: f = fullscreen, s = auto-scroll, / = shortcuts, Esc = close.
  const toggleAutoScrollRef = useRef(toggleAutoScroll)
  useEffect(() => { toggleAutoScrollRef.current = toggleAutoScroll }, [toggleAutoScroll])

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
      if (e.key === 's' || e.key === 'S') {
        e.preventDefault()
        setMenuOpen(true)
      }
      if (e.key === 'p' || e.key === 'P') {
        e.preventDefault()
        toggleAutoScrollRef.current()
      }
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

  usePageAnchor(data, data?.manga_id, chapterId ?? '', containerRef, pageLayout)

  useAutoScroll(
    autoScrollActive,
    autoScrollSpeed,
    isTouchingRef,
    () => set('autoScroll', false),
  )

  const chaptersHref = data?.manga_id
    ? DEEP_LINKS.MANGA_DETAIL({ mangaId: data.manga_id })
    : '/'

  const prevDisabled = data?.prev_chapter == null
  const nextDisabled = data?.next_chapter == null

  return (
    <div className={`min-h-screen ${bgClass}`}>

      {/* Worker loading overlay */}
      {canvasOverlay !== 'gone' && (
        <div
          className={`fixed inset-0 z-40 flex flex-col items-center justify-center gap-3 bg-gray-950 transition-opacity duration-300 ${
            canvasOverlay === 'fade' ? 'pointer-events-none opacity-0' : 'opacity-100'
          }`}
        >
          <Icon as={Loader2} size="large" className="animate-spin text-indigo-500" />
          <p className="text-sm text-gray-500">
            {canvasLoadInfo.total > 0
              ? canvasLoadInfo.loaded > 0
                ? `Loading ${canvasLoadInfo.loaded} / ${canvasLoadInfo.total} page${canvasLoadInfo.total !== 1 ? 's' : ''}…`
                : `Loading ${canvasLoadInfo.total} page${canvasLoadInfo.total !== 1 ? 's' : ''}…`
              : 'Loading…'}
          </p>
        </div>
      )}


      <ReaderHeader
        visible={headerVisible}
        chapter={chapterName ?? data?.chapter_id ?? chapterId ?? ''}
        chapterNameLoading={chapterNameLoading}
        pageCount={data?.pages?.length ?? null}
        chaptersHref={chaptersHref}
        menuOpen={menuOpen}
        onMenuToggle={() => setMenuOpen(o => !o)}
        onShortcutsToggle={() => setShortcutsOpen(o => !o)}
        prevDisabled={prevDisabled}
        nextDisabled={nextDisabled}
        onPrev={goPrev}
        onNext={goNext}
      />

      <ReaderMenu open={menuOpen} onClose={() => setMenuOpen(false)}>
        <ReaderSettingsPanel
          settings={settings}
          set={set}
          headerVisible={headerVisible}
          onHeaderToggle={() => setHeaderVisible(v => !v)}
          prevDisabled={prevDisabled}
          nextDisabled={nextDisabled}
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
            key={chapterId}
            urls={data.pages ?? []}
            containerRef={containerRef}
            onLayout={setPageLayout}
            onLoadingState={handleCanvasLoadingState}
          />
        ) : undefined}
        renderFooter={data && canvasOverlay === 'gone' ? () => (
          <ChapterNavFooter
            chaptersHref={chaptersHref}
            prevDisabled={prevDisabled}
            nextDisabled={nextDisabled}
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

      <AutoScrollControls
        speed={autoScrollSpeed}
        onCycle={cycleSpeed}
      />

    </div>
  )
}

export default function ReaderPage() {
  return (
    <InteractiveProvider>
      <ReaderContent />
    </InteractiveProvider>
  )
}

