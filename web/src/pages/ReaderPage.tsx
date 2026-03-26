import { useState } from 'react'
import { Link } from 'react-router-dom'
import { useMangaReader } from '../providers/MangaReaderDataProvider'
import { usePageAnchor, pageAnchor } from '../hooks/usePageAnchor'
import { DEEP_LINKS } from '../lib/deepLinks'
import { InteractiveProvider, useInteractive } from '../components/reader/InteractiveProvider'
import { useReaderSettings } from '../components/reader/useReaderSettings'
import { useAutoScroll } from '../hooks/useAutoScroll'
import { ReaderHeader } from '../components/reader/ReaderHeader'
import { ReaderMenu } from '../components/reader/ReaderMenu'
import { ReaderSettingsPanel } from '../components/reader/ReaderSettings'
import { ReaderStrip } from '../components/reader/ReaderStrip'
import { ReaderProgressBar } from '../components/reader/ReaderProgressBar'

// ── Inner content — requires InteractiveProvider in the tree ──────────────────

function ReaderContent() {
  const { data, loading, error, chapter, mangaId, lang, goNext, goPrev } = useMangaReader()
  const { settings, set, stripMaxWidthClass, bgClass } = useReaderSettings()
  const { scrollPct, headerVisible, setHeaderVisible, isTouchingRef, onStripTap } = useInteractive()
  const [menuOpen, setMenuOpen] = useState(false)
  const overlay = usePageAnchor(data, mangaId, chapter)

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
        gap={settings.stripGap}
        onClick={onStripTap}
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
        renderPages={data ? () => data.pages.map((url, i) => (
          <img
            key={i}
            id={pageAnchor(mangaId!, chapter, i + 1)}
            src={url}
            alt={`Page ${i + 1}`}
            className="w-full"
            loading={i < 3 ? 'eager' : 'lazy'}
          />
        )) : undefined}
        renderFooter={data ? () => (
          <div className="my-10 flex flex-wrap justify-center gap-4 px-4">
            <button
              onClick={goPrev}
              disabled={data.prevChapter == null}
              className="rounded-lg border border-gray-700 bg-gray-900 px-5 py-2.5 text-sm text-gray-300 transition hover:border-gray-500 hover:text-white disabled:cursor-not-allowed disabled:opacity-30"
            >
              ← Previous Chapter
            </button>
            <Link
              to={chaptersHref}
              className="rounded-lg border border-gray-700 bg-gray-900 px-5 py-2.5 text-sm text-gray-300 transition hover:border-gray-500 hover:text-white"
            >
              Chapter List
            </Link>
            <button
              onClick={goNext}
              disabled={data.nextChapter == null}
              className="rounded-lg border border-gray-700 bg-gray-900 px-5 py-2.5 text-sm text-gray-300 transition hover:border-gray-500 hover:text-white disabled:cursor-not-allowed disabled:opacity-30"
            >
              Next Chapter →
            </button>
          </div>
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
