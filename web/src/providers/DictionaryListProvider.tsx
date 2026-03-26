import { createContext, useContext, useState, type ReactNode } from 'react'
import { dictionaryApi } from '../api/dictionary'
import { formatDictionaryEntry } from '../lib/formatters'
import type { AppDictionaryEntry } from '../types/app'

interface DictionaryListContextValue {
  results: AppDictionaryEntry[] | null
  loading: boolean
  error: string | null
  search: (query: string) => Promise<void>
}

const DictionaryListContext = createContext<DictionaryListContextValue | null>(null)

export function DictionaryListProvider({ children }: { children: ReactNode }) {
  const [results, setResults] = useState<AppDictionaryEntry[] | null>(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  async function search(query: string) {
    setLoading(true)
    setError(null)
    setResults(null)
    try {
      const raw = await dictionaryApi.search(query)
      setResults(raw.map(formatDictionaryEntry))
    } catch (err) {
      setError(err instanceof Error ? err.message : String(err))
    } finally {
      setLoading(false)
    }
  }

  return (
    <DictionaryListContext.Provider value={{ results, loading, error, search }}>
      {children}
    </DictionaryListContext.Provider>
  )
}

export function useDictionaryList() {
  const ctx = useContext(DictionaryListContext)
  if (!ctx) throw new Error('useDictionaryList must be used within DictionaryListProvider')
  return ctx
}
