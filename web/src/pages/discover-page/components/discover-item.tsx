import type { DomainDictionaryResponse } from '@/types'

function DiscoverItem({ entry }: { entry: DomainDictionaryResponse }) {
  const chaptersByLang = entry.chapters_by_lang ?? {}
  const totalChapters = Object.values(chaptersByLang).reduce((a, b) => a + b, 0)

  return (
    <div className="flex gap-4 rounded-lg border border-gray-800 bg-gray-900 p-4">
      {entry.cover_url && (
        <img
          src={entry.cover_url}
          alt={entry.title}
          className="h-20 w-14 flex-shrink-0 rounded object-cover"
        />
      )}
      <div className="min-w-0 flex-1">
        <h3 className="font-medium text-gray-100">{entry.title}</h3>
        <div className="mt-1 flex flex-wrap gap-x-4 gap-y-1 text-xs text-gray-500">
          {entry.sources && Object.keys(entry.sources).length > 0 && (
            <span>Sources: {Object.keys(entry.sources).join(', ')}</span>
          )}
          {totalChapters > 0 && <span>{totalChapters} total chapters</span>}
          {Object.entries(chaptersByLang).map(([lang, count]) => (
            <span key={lang}>{lang.toUpperCase()}: {count}</span>
          ))}
        </div>
      </div>
    </div>
  )
}

export default DiscoverItem
