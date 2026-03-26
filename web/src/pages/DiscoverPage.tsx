import { useState, useEffect } from 'react'
import { useSearchParams } from 'react-router-dom'
import { useDictionaryList } from '../providers/DictionaryListProvider'
import { watchlistApi } from '../api/watchlist'
import type { AppDictionaryEntry } from '../types/app'

type AddState = 'idle' | 'loading' | 'done' | 'error'

function DiscoverItem({ entry }: { entry: AppDictionaryEntry }) {
  const [addState, setAddState] = useState<AddState>('idle')

  async function handleAdd() {
    setAddState('loading')
    try {
      await watchlistApi.add(entry.id)
      setAddState('done')
    } catch {
      setAddState('error')
    }
  }

  return (
    <div className="flex gap-4 rounded-lg border border-gray-800 bg-gray-900 p-4">
      {entry.coverUrl && (
        <img
          src={entry.coverUrl}
          alt={entry.title}
          className="h-20 w-14 flex-shrink-0 rounded object-cover"
        />
      )}
      <div className="min-w-0 flex-1">
        <h3 className="font-medium text-gray-100">{entry.title}</h3>
        <div className="mt-1 flex flex-wrap gap-x-4 gap-y-1 text-xs text-gray-500">
          {Object.keys(entry.sources).length > 0 && (
            <span>Sources: {Object.keys(entry.sources).join(', ')}</span>
          )}
          {entry.totalChapters > 0 && <span>{entry.totalChapters} total chapters</span>}
          {Object.entries(entry.chaptersByLang).map(([lang, count]) => (
            <span key={lang}>{lang.toUpperCase()}: {count}</span>
          ))}
        </div>
      </div>
      <div className="flex flex-shrink-0 items-start">
        <button
          onClick={handleAdd}
          disabled={addState !== 'idle'}
          className={`rounded-lg px-3 py-1.5 text-xs font-medium transition disabled:cursor-not-allowed ${
            addState === 'done' ? 'bg-green-900 text-green-300'
            : addState === 'error' ? 'bg-red-900 text-red-300'
            : addState === 'loading' ? 'bg-gray-700 text-gray-400'
            : 'bg-indigo-600 text-white hover:bg-indigo-500'
          }`}
        >
          {addState === 'done' ? '✓ Added'
            : addState === 'error' ? 'Failed'
            : addState === 'loading' ? 'Adding…'
            : '+ Library'}
        </button>
      </div>
    </div>
  )
}

export default function DiscoverPage() {
  const [searchParams, setSearchParams] = useSearchParams()
  const [query, setQuery] = useState(() => searchParams.get('q') ?? '')
  const { results, loading, error, search } = useDictionaryList()

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
