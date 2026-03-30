import { RefreshCw } from 'lucide-react'
import type { DomainDictionaryResponse } from '@/types'
import { ButtonIcon, Heading, Text } from '@/components/ui'
import { formatDate, DateFormat } from '@/lib/format-date'
import { useApiRefreshDictionary } from '@/hooks/use-api-refresh-dictionary'

function DiscoverItem({ entry }: { entry: DomainDictionaryResponse }) {
  const { state: refreshState, refresh } = useApiRefreshDictionary()
  const chaptersByLang = entry.chapters_by_lang ?? {}
  const totalChapters = Object.values(chaptersByLang).reduce((a, b) => a + b, 0)
  const sourceNames = entry.sources ? Object.values(entry.sources).join(', ') : null
  const languages = Object.keys(chaptersByLang).map(l => l.toUpperCase()).join(', ')
  const isRefreshing = refreshState === 'loading'
  const isQueued = refreshState === 'done'

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
          {sourceNames && <Text size="xs" color="muted">{sourceNames}</Text>}
          {languages && <Text size="xs" color="muted">{languages}</Text>}
          {totalChapters > 0 && <Text size="xs" color="muted">{totalChapters} chapters</Text>}
        </div>
        {entry.updated_at && (
          <Text size="xs" color="muted" className="mt-1">
            Last checked {formatDate(entry.updated_at, DateFormat.ShortDateTime)}
          </Text>
        )}
      </div>
      <div className="flex flex-col justify-between">
        <ButtonIcon
          variant="ghost"
          size="sm"
          aria-label="Refresh dictionary"
          disabled={isRefreshing || isQueued}
          onClick={() => refresh(entry.id!)}
        >
          <RefreshCw
            size={14}
            className={isRefreshing ? 'animate-spin' : isQueued ? 'text-indigo-400' : ''}
          />
        </ButtonIcon>
      </div>
    </div>
  )
}

export default DiscoverItem
