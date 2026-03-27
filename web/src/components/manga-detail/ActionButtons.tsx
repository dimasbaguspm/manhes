import { Button } from '../Button'
import { formatDate, DateFormat } from '../../lib/formatDate'

type ActionState = 'idle' | 'loading' | 'done' | 'error'

interface ActionButtonsProps {
  state: string
  addState: ActionState
  onAddToWatchlist: () => void
  refreshState: ActionState
  onRefresh: () => void
  updatedAt: string
}

export function ActionButtons({
  state,
  addState,
  onAddToWatchlist,
  refreshState,
  onRefresh,
  updatedAt,
}: ActionButtonsProps) {
  if (state === 'unavailable') {
    return (
      <Button
        variant={addState === 'error' ? 'danger' : 'primary'}
        disabled={addState !== 'idle'}
        onClick={onAddToWatchlist}
        className="mt-4"
      >
        {addState === 'error' ? 'Failed to add'
          : addState === 'loading' ? 'Adding…'
          : '+ Add to Library'}
      </Button>
    )
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
