import { useState, useEffect } from 'react'
import { useSearchParams } from 'react-router-dom'
import { useApiSearchDictionary } from '../../hooks/useApiSearchDictionary'
import DiscoverItem from './components/DiscoverItem'

export default function DiscoverPage() {
  const [searchParams, setSearchParams] = useSearchParams()
  const [query, setQuery] = useState(() => searchParams.get('q') ?? '')
  const { results, loading, error, search } = useApiSearchDictionary()

  useEffect(() => {
    const q = searchParams.get('q')
    if (q) search(q)
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    const q = query.trim()
    if (!q) return
    setSearchParams({ q }, { replace: true })
    search(q)
  }

  return (
    <div className="max-w-4xl">
      <h1 className="mb-6 text-xl font-semibold text-gray-100">Discover</h1>

      <form onSubmit={handleSubmit} className="mb-8 flex gap-3">
        <input
          type="text"
          value={query}
          onChange={e => setQuery(e.target.value)}
          placeholder="Search for manga across all sources..."
          className="flex-1 rounded-lg border border-gray-700 bg-gray-900 px-4 py-2.5 text-sm text-gray-100 placeholder-gray-500 focus:border-indigo-500 focus:outline-none"
          autoFocus
        />
        <button
          type="submit"
          disabled={loading}
          className="rounded-lg bg-indigo-600 px-5 py-2.5 text-sm font-medium text-white transition hover:bg-indigo-500 disabled:opacity-60"
        >
          {loading ? 'Searching…' : 'Search'}
        </button>
      </form>

      {error && (
        <div className="mb-4 rounded-lg border border-red-800 bg-red-950 px-4 py-3 text-sm text-red-300">
          {error}
        </div>
      )}

      {results !== null && results.length === 0 && (
        <p className="text-center text-gray-500">No results for "{query}"</p>
      )}

      {results && results.length > 0 && (
        <div className="space-y-3">
          <p className="text-sm text-gray-500">{results.length} results</p>
          {results.map(entry => (
            <DiscoverItem key={entry.id} entry={entry} />
          ))}
        </div>
      )}
    </div>
  )
}
