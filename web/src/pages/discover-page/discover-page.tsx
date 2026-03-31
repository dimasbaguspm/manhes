import { useState, useEffect } from 'react'
import { useSearchParams } from 'react-router-dom'
import { useApiSearchDictionary } from '@/hooks/use-api-search-dictionary'
import { Badge, Button, Heading, NoResults, Text } from '@/components/ui'
import { DiscoverItem } from './components/discover-item'

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
    <div className="w-full">
      <Heading level="h1" className="mb-6">Discover</Heading>

      <form onSubmit={handleSubmit} className="mb-8 flex gap-3">
        <input
          type="text"
          value={query}
          onChange={e => setQuery(e.target.value)}
          placeholder="Search for manga across all sources..."
          className="flex-1 rounded-lg border border-gray-700 bg-gray-900 px-4 py-2.5 text-sm text-gray-100 placeholder-gray-500 focus:border-indigo-500 focus:outline-none"
          autoFocus
        />
        <Button
          type="submit"
          disabled={loading}
          color="primary"
        >
          {loading ? 'Searching…' : 'Search'}
        </Button>
      </form>

      {error && (
        <Badge variant="error" className="mb-4">{error}</Badge>
      )}

      {results !== null && results.length === 0 && (
        <NoResults message={`No results for "${query}"`} />
      )}

      {results && results.length > 0 && (
        <div>
          <Text size="sm" color="muted" className="mb-4">{results.length} results</Text>
          <div className="grid grid-cols-2 gap-3 sm:grid-cols-3 md:grid-cols-5 lg:grid-cols-6">
            {results.map(entry => (
              <DiscoverItem key={entry.id} entry={entry} />
            ))}
          </div>
        </div>
      )}
    </div>
  )
}
