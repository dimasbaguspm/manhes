import type { DomainDictionaryResponse } from '@/types'
import { Heading, Text } from '@/components/ui'

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
        <Heading level="h3">{entry.title}</Heading>
        <div className="mt-1 flex flex-wrap gap-x-4 gap-y-1">
          {entry.sources && Object.keys(entry.sources).length > 0 && (
            <Text size="xs" color="muted">Sources: {Object.keys(entry.sources).join(', ')}</Text>
          )}
          {totalChapters > 0 && <Text size="xs" color="muted">{totalChapters} total chapters</Text>}
          {Object.entries(chaptersByLang).map(([lang, count]) => (
            <Text size="xs" color="muted" key={lang}>{lang.toUpperCase()}: {count}</Text>
          ))}
        </div>
      </div>
    </div>
  )
}

export default DiscoverItem
