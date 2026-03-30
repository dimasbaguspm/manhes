import { api } from './client'
import type { DomainDictionaryResponse } from '../types'

export const dictionaryApi = {
  search(q: string) {
    return api.get<DomainDictionaryResponse[]>(`/dictionary?q=${encodeURIComponent(q)}`)
  },

  // Swagger: POST /dictionary/refresh with body { dictionaryId }
  refresh(dictionaryId: string) {
    return api.post<{ status: string }>('/dictionary/refresh', { dictionaryId })
  },
}
