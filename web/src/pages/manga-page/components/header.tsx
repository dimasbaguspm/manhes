import { Heading } from '@/components/ui'

interface MangaDetailHeaderProps {
  title: string
  authors: string[]
}

export function MangaDetailHeader({ title, authors }: MangaDetailHeaderProps) {
  return (
    <div>
      <Heading level="h1" className="text-2xl font-bold text-gray-100 text-center sm:text-left">
        {title}
      </Heading>

      {authors.length > 0 && (
        <p className="mt-1 text-sm text-gray-400 text-center sm:text-left">{authors.join(', ')}</p>
      )}
    </div>
  )
}
