import { Progress } from '@/components/ui'

interface LangProgressBarProps {
  totalChapters: number
  availableChapters: number
}

export function LangProgressBar({ totalChapters, availableChapters }: LangProgressBarProps) {
  if (totalChapters === 0) return null

  const value = Math.round((availableChapters / totalChapters) * 100)

  return (
    <div className="mb-4 space-y-1">
      <Progress value={availableChapters} max={totalChapters} color="primary" size="sm" />
      <p className="text-xs text-gray-500">
        <span className="text-indigo-400">{availableChapters}</span>
        <span> / {totalChapters} chapters available</span>
      </p>
    </div>
  )
}
