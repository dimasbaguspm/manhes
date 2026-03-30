import { useState, useCallback } from 'react'
import { dictionaryApi } from '../api/dictionary'

type RefreshState = 'idle' | 'loading' | 'done' | 'error'

export function useApiRefreshDictionary() {
  const [state, setState] = useState<RefreshState>('idle')

  const refresh = useCallback(async (dictionaryId: string) => {
    setState('loading')
    try {
      await dictionaryApi.refresh(dictionaryId)
      setState('done')
    } catch {
      setState('error')
    }
  }, [])

  return { state, refresh }
}
