import { BookOpen } from 'lucide-react'
import { Icon } from '../../../components/Icon'

interface CoverImageProps {
  src?: string
  alt: string
}

export function CoverImage({ src, alt }: CoverImageProps) {
  return (
    <div className="h-56 w-36 flex-shrink-0 overflow-hidden rounded-lg bg-gray-800 sm:h-64 sm:w-44">
      {src ? (
        <img src={src} alt={alt} className="h-full w-full object-cover" />
      ) : (
        <div className="flex h-full flex-col items-center justify-center gap-2 px-3 text-center">
          <Icon as={BookOpen} size="large" className="text-gray-700" />
          <p className="text-xs leading-snug text-gray-600">
            Cover on the way — please be patient
          </p>
        </div>
      )}
    </div>
  )
}
