import { Badge } from '@/components/ui'

const STATUS_COLORS: Record<string, string> = {
  ongoing:   'bg-green-900 text-green-300',
  completed: 'bg-blue-900 text-blue-300',
  hiatus:    'bg-yellow-900 text-yellow-300',
}

const STATE_COLORS: Record<string, string> = {
  available:   'bg-emerald-900 text-emerald-300',
  fetching:    'bg-sky-900 text-sky-300',
  uploading:   'bg-violet-900 text-violet-300',
  unavailable: 'bg-gray-800 text-gray-500',
}

export function StatusBadge({ status }: { status: string }) {
  return (
    <Badge className={STATUS_COLORS[status] ?? 'bg-gray-800 text-gray-400'}>
      {status}
    </Badge>
  )
}

export function StateBadge({ state }: { state: string }) {
  return (
    <Badge className={STATE_COLORS[state] ?? 'bg-gray-800 text-gray-400'}>
      {state}
    </Badge>
  )
}

export function GenreBadge({ genre }: { genre: string }) {
  return (
    <Badge className="bg-gray-800 text-gray-400">
      {genre}
    </Badge>
  )
}
