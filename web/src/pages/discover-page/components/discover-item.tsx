import type { DomainDictionaryResponse } from '@/types'
import { Button } from '@/components/button'
import { Image } from '@/components/ui/image'
import { useApiRefreshDictionary } from '@/hooks/use-api-refresh-dictionary'

export function DiscoverItem({ entry }: { entry: DomainDictionaryResponse }) {
  const { state: refreshState, refresh } = useApiRefreshDictionary()
  const isRefreshing = refreshState === 'loading'
  const chaptersByLang = entry.chapters_by_lang ?? {}
  const totalChapters = Object.values(chaptersByLang).reduce((a, b) => a + b, 0)
  const languages = Object.keys(chaptersByLang).map(l => l.toUpperCase()).join(', ')

  function handleAddToLibrary() {
    if (!entry.id) return
    refresh(entry.id)
  }

  return (
    <div className="flex flex-col overflow-hidden rounded-lg border border-gray-800 bg-gray-900 transition hover:border-gray-700">
      {/* Cover image */}
      <div className="relative aspect-[2/3] w-full bg-gray-800">
        <Image
          src={entry.cover_url}
          alt={entry.title}
          size="xl"
          aspect="portrait"
          className="h-full w-full"
        />
      </div>

      {/* Card content */}
      <div className="flex flex-1 flex-col p-3">
        <h3 className="line-clamp-2 text-sm font-semibold text-gray-100">{entry.title}</h3>

        <div className="mt-1 flex flex-wrap items-center gap-x-2 gap-y-0.5">
          {languages && (
            <span className="text-xs text-gray-500">{languages}</span>
          )}
          {totalChapters > 0 && (
            <span className="text-xs text-gray-500">{totalChapters} ch</span>
          )}
        </div>

        <div className="mt-auto pt-3">
          <Button
            variant="primary"
            className="w-full"
            onClick={handleAddToLibrary}
            disabled={isRefreshing}
          >
            {isRefreshing ? 'Adding…' : 'Add to Library'}
          </Button>
        </div>
      </div>
    </div>
  )
}
