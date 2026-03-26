import { api } from './client'
import type { DomainDictionaryResponse } from '../types'

export const dictionaryApi = {
  search(q: string) {
    return api.get<DomainDictionaryResponse[]>(`/dictionary?q=${encodeURIComponent(q)}`)
  },
  refresh(dictionaryId: string) {
    return api.post<DomainDictionaryResponse>(`/dictionary/${encodeURIComponent(dictionaryId)}`, null)
  },
}
