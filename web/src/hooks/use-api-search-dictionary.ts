import { useState, useCallback } from 'react'
import { dictionaryApi } from '@/api/dictionary'
import { formatDictionary } from '@/lib/format-data'
import type { DomainDictionaryResponse } from '@/types'

export function useApiSearchDictionary() {
  const [results, setResults] = useState<DomainDictionaryResponse[] | null>(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const search = useCallback(async (q: string) => {
    setLoading(true)
    setError(null)
    setResults(null)
    try {
      const raw = await dictionaryApi.search(q)
      setResults(raw.map(formatDictionary))
    } catch (err) {
      setError(err instanceof Error ? err.message : String(err))
    } finally {
      setLoading(false)
    }
  }, [])

  return { results, loading, error, search }
}
