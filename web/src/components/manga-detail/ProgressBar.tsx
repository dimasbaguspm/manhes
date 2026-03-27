interface LangProgressBarProps {
  totalChapters: number
  fetchedChapters: number
  uploadedChapters: number
}

export function LangProgressBar({ totalChapters, fetchedChapters, uploadedChapters }: LangProgressBarProps) {
  if (totalChapters === 0) return null

  return (
    <div className="mb-4 space-y-1">
      <div className="h-1.5 w-full overflow-hidden rounded-full bg-gray-800">
        <div className="flex h-full">
          <div
            className="h-full bg-indigo-500 transition-all"
            style={{ width: `${Math.round((uploadedChapters / totalChapters) * 100)}%` }}
          />
          <div
            className="h-full bg-yellow-700 transition-all"
            style={{ width: `${Math.round(((fetchedChapters - uploadedChapters) / totalChapters) * 100)}%` }}
          />
        </div>
      </div>
      <p className="text-xs text-gray-500">
        <span className="text-indigo-400">{uploadedChapters}</span>
        {fetchedChapters > uploadedChapters && (
          <span className="text-yellow-600"> +{fetchedChapters - uploadedChapters} fetching</span>
        )}
        <span> / {totalChapters} chapters from source</span>
      </p>
    </div>
  )
}
