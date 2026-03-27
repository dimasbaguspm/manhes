import { useState, useRef, useEffect, useCallback } from 'react'
import { Link } from 'react-router-dom'
import { useMangaReader } from '../providers/MangaReaderDataProvider'
import { usePageAnchor, type CanvasPageLayout } from '../hooks/usePageAnchor'
import { DEEP_LINKS } from '../lib/deepLinks'
import { InteractiveProvider, useInteractive } from '../components/reader/InteractiveProvider'
import { useReaderSettings } from '../components/reader/useReaderSettings'
import { useAutoScroll } from '../hooks/useAutoScroll'
import { ReaderHeader } from '../components/reader/ReaderHeader'
import { ReaderMenu } from '../components/reader/ReaderMenu'
import { ReaderSettingsPanel } from '../components/reader/ReaderSettings'
import { ReaderStrip } from '../components/reader/ReaderStrip'
import { ReaderProgressBar } from '../components/reader/ReaderProgressBar'
import { ReaderCanvas } from '../components/reader/ReaderCanvas'
import { ChevronLeft, ChevronRight } from '../components/reader/Icons'

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

  const canvasRef = useRef<HTMLCanvasElement>(null)
  const [pageLayout, setPageLayout] = useState<CanvasPageLayout[] | null>(null)

  // Reset layout whenever the chapter changes so the anchor hook doesn't use
  // stale geometry while the new canvas is being drawn.
  useEffect(() => {
    setPageLayout(null)
  }, [chapter])

  // Persist reading progress (%) to localStorage when chapter changes or on unmount.
  // We use a ref so the cleanup closure always captures the latest scrollPct
  // without needing it in the effect dependency array (which would re-register
  // the cleanup on every scroll event).
  const scrollPctSaveRef = useRef(scrollPct)
  useEffect(() => { scrollPctSaveRef.current = scrollPct })

  const saveProgress = useCallback(() => {
    if (!mangaId || !lang || !chapter) return
    try {
      const raw = localStorage.getItem('manhes_read_progress')
      const prev = (raw ? JSON.parse(raw) : {}) as Record<string, number>
      localStorage.setItem('manhes_read_progress', JSON.stringify({
        ...prev,
        [`${mangaId}/${lang}/${chapter}`]: scrollPctSaveRef.current,
      }))
    } catch {}
  }, [mangaId, lang, chapter])

  useEffect(() => {
    return () => { saveProgress() }
  }, [saveProgress])

  // Double-tap-hold opens settings.
  useEffect(() => {
    doubleTapHoldCallbackRef.current = () => setMenuOpen(true)
    return () => { doubleTapHoldCallbackRef.current = null }
  }, [doubleTapHoldCallbackRef])

  // Keyboard shortcuts: f = fullscreen, s = settings.
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
        setMenuOpen(v => !v)
      }
    }
    window.addEventListener('keydown', handler)
    return () => window.removeEventListener('keydown', handler)
  }, [])

  const overlay = usePageAnchor(data, mangaId, chapter, canvasRef, pageLayout)

  useAutoScroll(
    settings.autoScroll,
    settings.autoScrollSpeed,
    isTouchingRef,
    () => set('autoScroll', false),
  )

  const chaptersHref = DEEP_LINKS.MANGA_CHAPTERS({ mangaId: mangaId!, lang: lang! })

  return (
    <div className={`min-h-screen ${bgClass}`}>

      {/* Scroll-restore overlay — fades out once the target page is scrolled to */}
      {overlay !== 'gone' && (
        <div
          className={`fixed inset-0 z-50 bg-gray-950 transition-opacity duration-300 ${
            overlay === 'fade' ? 'opacity-0' : 'opacity-100'
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
          />
        ) : undefined}
        renderFooter={data ? () => (
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

// ── Chapter navigation footer ─────────────────────────────────────────────────

interface ChapterNavFooterProps {
  chaptersHref: string
  prevDisabled: boolean
  nextDisabled: boolean
  onPrev: () => void
  onNext: () => void
}

const navBtnBase =
  'inline-flex items-center gap-1.5 rounded-lg border border-gray-700 bg-gray-900 py-2.5 text-sm text-gray-300 transition hover:border-gray-500 hover:text-white disabled:cursor-not-allowed disabled:opacity-30'

function ChapterNavFooter({ chaptersHref, prevDisabled, nextDisabled, onPrev, onNext }: ChapterNavFooterProps) {
  return (
    <div className="my-10 flex items-center justify-center gap-2 px-4 sm:gap-4">
      <button
        onClick={onPrev}
        disabled={prevDisabled}
        className={`${navBtnBase} px-3 sm:px-5`}
      >
        <ChevronLeft className="h-4 w-4 shrink-0" />
        <span className="hidden sm:inline">Previous</span>
        <span className="sm:hidden">Prev</span>
      </button>

      <Link
        to={chaptersHref}
        className={`${navBtnBase} px-3 sm:px-5`}
      >
        <span className="hidden sm:inline">Chapter List</span>
        <span className="sm:hidden">List</span>
      </Link>

      <button
        onClick={onNext}
        disabled={nextDisabled}
        className={`${navBtnBase} px-3 sm:px-5`}
      >
        <span className="hidden sm:inline">Next</span>
        <span className="sm:hidden">Next</span>
        <ChevronRight className="h-4 w-4 shrink-0" />
      </button>
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
