interface NoResultsProps {
  message: string
  error?: boolean
}

export function NoResults({ message, error = false }: NoResultsProps) {
  if (error) {
    return (
      <div className="rounded-lg border border-red-800 bg-red-950 px-4 py-3 text-sm text-red-300">
        {message}
      </div>
    )
  }
  return (
    <div className="rounded-lg border border-gray-800 bg-gray-900 px-4 py-6 text-center text-gray-500">
      {message}
    </div>
  )
}
