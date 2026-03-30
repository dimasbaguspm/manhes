import { Heart } from 'lucide-react'
import { Icon } from '@/components'

interface MangaDetailHeaderProps {
  title: string
  authors: string[]
  isFavorite: boolean
  onToggleFavorite: () => void
}

export function MangaDetailHeader({ title, authors, isFavorite, onToggleFavorite }: MangaDetailHeaderProps) {
  return (
    <div>
      <div className="flex items-start justify-center gap-2 sm:justify-start">
        <h1 className="text-2xl font-bold text-gray-100">{title}</h1>
        <button
          onClick={onToggleFavorite}
          aria-label={isFavorite ? 'Remove from favorites' : 'Add to favorites'}
          className={`mt-1 shrink-0 transition ${
            isFavorite ? 'text-red-400 hover:text-red-300' : 'text-gray-600 hover:text-gray-400'
          }`}
        >
          <Icon as={Heart} fill={isFavorite ? 'currentColor' : 'none'} />
        </button>
      </div>

      {authors.length > 0 && (
        <p className="mt-1 text-sm text-gray-400">{authors.join(', ')}</p>
      )}
    </div>
  )
}
