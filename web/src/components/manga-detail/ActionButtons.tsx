import { Button } from '../Button'
import { formatDate, DateFormat } from '../../lib/formatDate'

type RefreshState = 'idle' | 'loading' | 'done' | 'error'

interface ActionButtonsProps {
  state: string
  refreshState: RefreshState
  onRefresh: () => void
  updatedAt: string
}

export function ActionButtons({
  state,
  refreshState,
  onRefresh,
  updatedAt,
}: ActionButtonsProps) {
  if (state === 'unavailable') {
    return null
  }

  return (
    <div className="mt-4 flex items-center justify-center gap-3 sm:justify-start">
      <Button
        variant={
          refreshState === 'error' ? 'outline-danger'
          : refreshState === 'done' || refreshState === 'loading' ? 'muted'
          : 'default'
        }
        disabled={refreshState === 'loading' || refreshState === 'done'}
        onClick={onRefresh}
      >
        {refreshState === 'loading' ? 'Refreshing…'
          : refreshState === 'error' ? 'Refresh failed'
          : refreshState === 'done' ? 'Queued'
          : 'Refresh'}
      </Button>

      {updatedAt && (
        <span className="text-xs text-gray-500">
          Updated {formatDate(updatedAt, DateFormat.ShortDateTime)}
        </span>
      )}
    </div>
  )
}
