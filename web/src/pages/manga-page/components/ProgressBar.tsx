interface LangProgressBarProps {
  totalChapters: number
  availableChapters: number
}

export function LangProgressBar({ totalChapters, availableChapters }: LangProgressBarProps) {
  if (totalChapters === 0) return null

  return (
    <div className="mb-4 space-y-1">
      <div className="h-1.5 w-full overflow-hidden rounded-full bg-gray-800">
        <div
          className="h-full bg-indigo-500 transition-all"
          style={{ width: `${Math.round((availableChapters / totalChapters) * 100)}%` }}
        />
      </div>
      <p className="text-xs text-gray-500">
        <span className="text-indigo-400">{availableChapters}</span>
        <span> / {totalChapters} chapters available</span>
      </p>
    </div>
  )
}
