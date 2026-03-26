package application

import (
	"fmt"
	"path/filepath"
	"sort"
	"time"

	"manga-engine/internal/domain"
)

func (s *IngestService) writeMetadataLocked(slug string, manga *domain.Manga, availByLang map[string]int) {
	s.metaMu.Lock()
	defer s.metaMu.Unlock()
	if err := s.writeMetadata(slug, manga, availByLang); err != nil {
		s.log.Warn("write metadata failed", "err", err)
	}
}

func (s *IngestService) writeLangMetadataLocked(slug, lang string, available int) {
	s.metaMu.Lock()
	defer s.metaMu.Unlock()
	if err := s.writeLangMetadata(slug, lang, available); err != nil {
		s.log.Warn("write lang metadata failed", "lang", lang, "err", err)
	}
}

func (s *IngestService) writeLangMetadata(slug, lang string, available int) error {
	chapters, err := s.repo.GetDownloadedChaptersByLang(slug, lang)
	if err != nil {
		return fmt.Errorf("get chapters: %w", err)
	}
	m := &domain.LangMetadata{
		Slug:       slug,
		Language:   lang,
		Available:  available,
		Downloaded: len(chapters),
		Chapters:   chapters,
		UpdatedAt:  time.Now(),
	}
	return s.disk.WriteLangMetadata(slug, lang, m)
}

func (s *IngestService) writeMetadata(slug string, manga *domain.Manga, availByLang map[string]int) error {
	downloaded, err := s.repo.GetDownloadedByLang(slug)
	if err != nil {
		return fmt.Errorf("get downloaded stats: %w", err)
	}

	langs := make(map[string]struct{})
	for l := range availByLang {
		langs[l] = struct{}{}
	}
	for l := range downloaded {
		langs[l] = struct{}{}
	}
	langStats := make(map[string]domain.LanguageStat, len(langs))
	for l := range langs {
		langStats[l] = domain.LanguageStat{
			Available:  availByLang[l],
			Downloaded: downloaded[l],
		}
	}

	existing, _ := s.disk.ReadMetadata(slug)

	meta := &domain.Metadata{
		Slug:      slug,
		Languages: langStats,
		UpdatedAt: time.Now(),
	}

	if manga != nil {
		meta.Title = manga.Title
		meta.Description = manga.Description
		meta.Status = manga.Status
		meta.Authors = manga.Authors
		meta.Genres = manga.Genres
		meta.Sources = manga.Sources
		if manga.CoverURL != "" {
			meta.Cover = "cover" + filepath.Ext(manga.CoverURL)
		}
	} else if existing != nil {
		meta.Title = existing.Title
		meta.Description = existing.Description
		meta.Status = existing.Status
		meta.Authors = existing.Authors
		meta.Genres = existing.Genres
		meta.Sources = existing.Sources
		meta.Cover = existing.Cover
	}

	return s.disk.WriteMetadata(slug, meta)
}

func orderedLangs(m map[string][]domain.Chapter) []string {
	langs := make([]string, 0, len(m))
	for l := range m {
		langs = append(langs, l)
	}
	sort.Slice(langs, func(i, j int) bool {
		if langs[i] == "en" {
			return true
		}
		if langs[j] == "en" {
			return false
		}
		return langs[i] < langs[j]
	})
	return langs
}
