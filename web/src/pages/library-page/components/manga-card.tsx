import { Link } from 'react-router-dom'
import { RefreshCw } from 'lucide-react'
import { Badge } from '@/components/ui/badge/badge'
import { Button } from '@/components/ui/button/button'
import { Text } from '@/components/ui/text/text'
import { ButtonIcon } from '@/components/ui/button-icon'
import { Image } from '@/components/ui/image'
import { useApiRefreshDictionary } from '@/hooks/use-api-refresh-dictionary'
import { DEEP_LINKS } from '@/lib/deep-links'
import { formatDate, DateFormat } from '@/lib/format-date'
import type { DomainMangaSummary } from '@/types'

const STATE_LABEL: Record<string, string> = {
  available: 'Available',
  fetching: 'Fetching',
  uploading: 'Uploading',
}

const STATE_BADGE_VARIANT: Record<string, 'success' | 'warning' | 'primary'> = {
  available: 'success',
  fetching: 'warning',
  uploading: 'primary',
}

const NEEDS_REFRESH_THRESHOLD_MS = 24 * 60 * 60 * 1000

function needsRefresh(updatedAt: string | undefined): boolean {
  if (!updatedAt) return false
  return Date.now() - new Date(updatedAt).getTime() > NEEDS_REFRESH_THRESHOLD_MS
}

export function MangaCard({ manga }: { manga: DomainMangaSummary }) {
  const { state: refreshState, refresh } = useApiRefreshDictionary()
  const isRefreshing = refreshState === 'loading'
  const isStale = needsRefresh(manga.updated_at)
  const languages = manga.languages?.map(l => l.lang?.toUpperCase()).filter(Boolean).join(', ')

  return (
    <div className="flex flex-col overflow-hidden rounded-lg border border-gray-800 bg-gray-900 transition hover:border-gray-700">
      {/* Cover image with state badge overlay */}
      <Link to={DEEP_LINKS.MANGA_DETAIL({ mangaId: manga.id ?? '' })} className="block">
        <div className="relative aspect-[2/3] w-full bg-gray-800">
          <Image
            src={manga.cover_url}
            alt={manga.title}
            size="sm"
            aspect="portrait"
            className="h-full w-full"
          />
          <div className="absolute top-2 right-2">
            <Badge
              variant={STATE_BADGE_VARIANT[manga.state ?? ''] ?? 'default'}
              size="sm"
            >
              {STATE_LABEL[manga.state ?? ''] ?? manga.state}
            </Badge>
          </div>
        </div>
      </Link>

      {/* Card content */}
      <div className="flex flex-1 flex-col p-3">
        <Link to={DEEP_LINKS.MANGA_DETAIL({ mangaId: manga.id ?? '' })} className="block">
          <h3 className="line-clamp-2 text-sm font-semibold text-gray-100">{manga.title}</h3>
        </Link>

        <div className="mt-1 flex flex-wrap items-center gap-x-2 gap-y-0.5">
          {languages && (
            <Text size="xs" color="muted">{languages}</Text>
          )}
          {manga.updated_at && (
            <Text size="xs" color="muted">{formatDate(manga.updated_at, DateFormat.Relative)}</Text>
          )}
        </div>

        {isStale && manga.dictionary_id && (
          <div className="mt-2">
            <ButtonIcon
              variant="ghost"
              size="sm"
              onClick={() => refresh(manga.dictionary_id!)}
              disabled={isRefreshing}
              aria-label="Refresh manga"
              className={isRefreshing ? 'animate-spin text-gray-400' : 'text-gray-400 hover:text-indigo-400'}
            >
              <RefreshCw size={14} />
            </ButtonIcon>
          </div>
        )}

        <div className="mt-auto pt-2">
          <Link to={DEEP_LINKS.MANGA_DETAIL({ mangaId: manga.id ?? '' })} className="block">
            <Button size="sm" className="w-full">View</Button>
          </Link>
        </div>
      </div>
    </div>
  )
}
