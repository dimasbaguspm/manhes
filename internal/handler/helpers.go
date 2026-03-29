package handler

import (
	"time"

	"manga-engine/internal/domain"
)

func pickBestSource(stats map[string]domain.SourceStat, prio map[string]int) map[string]string {
	langBest := make(map[string]struct{ src string; count int; prio int })
	for name, stat := range stats {
		if stat.Err != "" {
			continue
		}
		p := prio[name]
		for lang, count := range stat.ChaptersByLang {
			b := langBest[lang]
			if b.count == 0 || count > b.count || (count == b.count && p < b.prio) {
				langBest[lang] = struct{ src string; count int; prio int }{src: name, count: count, prio: p}
			}
		}
	}
	result := make(map[string]string, len(langBest))
	for lang, b := range langBest {
		result[lang] = b.src
	}
	return result
}

func sourcesChanged(old, new map[string]string) bool {
	if len(old) != len(new) {
		return true
	}
	for k, v := range old {
		if new[k] != v {
			return true
		}
	}
	return false
}

func statsChangedPure(old, new map[string]domain.SourceStat) bool {
	oldLangs := make(map[string]int)
	newLangs := make(map[string]int)
	for _, s := range old {
		for l, c := range s.ChaptersByLang {
			oldLangs[l] += c
		}
	}
	for _, s := range new {
		for l, c := range s.ChaptersByLang {
			newLangs[l] += c
		}
	}
	if len(oldLangs) != len(newLangs) {
		return true
	}
	for l, c := range oldLangs {
		if newLangs[l] != c {
			return true
		}
	}
	return false
}

func bestSourceChanged(old, new map[string]string) bool {
	if len(old) != len(new) {
		return true
	}
	for l, s := range old {
		if new[l] != s {
			return true
		}
	}
	return false
}

func toMangaSummary(m domain.Manga, languages []domain.MangaLangResponse) domain.MangaSummary {
	return domain.MangaSummary{
		ID:           m.ID,
		DictionaryID: m.DictionaryID,
		Title:        m.Title,
		Description:  m.Description,
		Status:       m.Status,
		CoverURL:     m.CoverURL,
		State:        string(m.State),
		Authors:      m.Authors,
		Genres:       m.Genres,
		Languages:    languages,
		UpdatedAt:    m.UpdatedAt,
		CreatedAt:    m.CreatedAt,
	}
}

func toChapterItem(ch domain.Chapter) domain.ChapterItem {
	var updatedAt *time.Time
	if !ch.UpdatedAt.IsZero() {
		ua := ch.UpdatedAt
		updatedAt = &ua
	}
	return domain.ChapterItem{
		ID:        ch.ID,
		Order:     int(ch.SortKey),
		Name:      ch.Number,
		UpdatedAt: updatedAt,
		PageCount: ch.PageCount,
	}
}

func buildPrevNext(chapters []domain.Chapter, currentID string) (prev, next *string) {
	for i, ch := range chapters {
		if ch.ID != currentID {
			continue
		}
		if i > 0 {
			prev = &chapters[i-1].ID
		}
		if i < len(chapters)-1 {
			next = &chapters[i+1].ID
		}
		return
	}
	return
}
