import { api } from './client'

export const watchlistApi = {
  add(dictionaryId: string) {
    return api.post<{ status: string; slug: string; dictionaryId: string }>(
      '/watchlist',
      { dictionaryId },
    )
  },
}
