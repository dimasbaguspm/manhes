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
  chapter?: string;
  page_count?: number;
  uploaded_at?: string;
}

export interface DomainChapterListResponse {
  chapters?: DomainChapterItem[];
  id?: string;
  lang?: string;
}

export interface DomainChapterReadResponse {
  chapter?: string;
  id?: string;
  lang?: string;
  next_chapter?: string;
  pages?: string[];
  prev_chapter?: string;
}

export interface DomainDictionaryResponse {
  best_source?: Record<string, string>;
  chapters_by_lang?: Record<string, number>;
  cover_url?: string;
  id?: string;
  refreshed_at?: string;
  slug?: string;
  source_stats?: Record<string, DomainSourceStat>;
  sources?: Record<string, string>;
  state?: string;
  title?: string;
}

export interface DomainMangaDetailResponse {
  authors?: string[];
  cover_url?: string;
  description?: string;
  genres?: string[];
  id?: string;
  languages?: DomainMangaLangResponse[];
  sources?: Record<string, string>;
  state?: string;
  status?: string;
  title?: string;
  updated_at?: string;
}

export interface DomainMangaLangResponse {
  fetched_chapters?: number;
  lang?: string;
  latest_update?: string;
  total_chapters?: number;
  uploaded_chapters?: number;
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
  chapters_by_lang?: Record<string, number>;
  cover_url?: string;
  description?: string;
  genres?: string[];
  id?: string;
  languages?: string[];
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

export interface HandlerWatchlistRequest {
  dictionaryId?: string;
}

export interface HttputilErrorResponse {
  code?: string;
  details?: string[];
  message?: string;
}
