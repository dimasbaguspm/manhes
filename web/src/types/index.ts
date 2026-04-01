/* eslint-disable */
/* tslint:disable */
// @ts-nocheck
/*
 * ---------------------------------------------------------------
 * ## THIS FILE WAS GENERATED VIA SWAGGER-TYPESCRIPT-API        ##
 * ##                                                           ##
 * ## AUTHOR: acacode                                           ##
 * ## SOURCE: https://github.com/acacode/swagger-typescript-api ##
 * ---------------------------------------------------------------
 */

export interface DomainChapterItem {
  id?: string;
  name?: string;
  order?: number;
  page_count?: number;
  updated_at?: string;
}

export interface DomainChapterListResponse {
  chapters?: DomainChapterItem[];
  id?: string;
  lang?: string;
}

export interface DomainChapterReadResponse {
  chapter_id?: string;
  manga_id?: string;
  next_chapter?: string;
  pages?: string[];
  prev_chapter?: string;
}

export interface DomainDictionaryRefreshRequest {
  dictionaryId?: string;
}

export interface DomainDictionaryResponse {
  best_source?: Record<string, string>;
  chapters_by_lang?: Record<string, number>;
  cover_url?: string;
  id?: string;
  slug?: string;
  source_stats?: Record<string, DomainSourceStat>;
  sources?: Record<string, string>;
  title?: string;
  updated_at?: string;
}

export interface DomainMangaDetailResponse {
  authors?: string[];
  cover_url?: string;
  created_at?: string;
  description?: string;
  dictionary_id?: string;
  genres?: string[];
  id?: string;
  languages?: DomainMangaLangResponse[];
  state?: string;
  status?: string;
  title?: string;
  updated_at?: string;
}

export interface DomainMangaLangResponse {
  available_chapters?: number;
  lang?: string;
  latest_update?: string;
  total_chapters?: number;
}

export interface DomainMangaListResponse {
  itemCount?: number;
  items?: DomainMangaSummary[];
  pageNumber?: number;
  pageSize?: number;
  pageTotal?: number;
}

export interface DomainMangaSummary {
  authors?: string[];
  cover_url?: string;
  created_at?: string;
  description?: string;
  dictionary_id?: string;
  genres?: string[];
  id?: string;
  languages?: DomainMangaLangResponse[];
  state?: string;
  status?: string;
  title?: string;
  updated_at?: string;
}

export interface DomainSourceStat {
  chapters_by_lang?: Record<string, number>;
  err?: string;
  fetched_at?: string;
}

export interface DomainTrackerResponse {
  chapter_id?: string;
  created_at?: string;
  id?: string;
  is_read?: boolean;
  manga_id?: string;
  /** JSON string, use json.RawMessage on the domain Tracker */
  metadata?: string;
  updated_at?: string;
}

export type DomainUpsertTrackerRequest = object;

export interface HttputilErrorResponse {
  code?: string;
  details?: string[];
  message?: string;
}
