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
          <svg className="h-10 w-10 text-gray-700" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={1}
              d="M12 6.253v13m0-13C10.832 5.477 9.246 5 7.5 5S4.168 5.477 3 6.253v13C4.168 18.477 5.754 18 7.5 18s3.332.477 4.5 1.253m0-13C13.168 5.477 14.754 5 16.5 5c1.747 0 3.332.477 4.5 1.253v13C19.832 18.477 18.247 18 16.5 18c-1.746 0-3.332.477-4.5 1.253"
            />
          </svg>
          <p className="text-xs leading-snug text-gray-600">
            Cover on the way — please be patient
          </p>
        </div>
      )}
    </div>
  )
}
