package atsu

import (
	"context"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"manga-engine/internal/domain"
)

var chapterNumRe = regexp.MustCompile(`(?i)chapter\s+([\d]+(?:\.[\d]+)?)`)

func (a *Adapter) FetchMangaDetail(ctx context.Context, id string) (*domain.Manga, error) {
	u := fmt.Sprintf("%s/manga/page?id=%s", a.apiBase, id)
	var resp mangaPageResp
	if err := a.get(ctx, u, &resp); err != nil {
		return nil, fmt.Errorf("atsu fetch manga: %w", err)
	}

	p := resp.MangaPage
	var genres []string
	for _, g := range p.Genres {
		genres = append(genres, g.Name)
	}
	var authors []string
	for _, au := range p.Authors {
		authors = append(authors, au.Name)
	}

	coverURL := ""
	if p.Poster != nil && p.Poster.LargeImage != "" {
		coverURL = a.staticBase + "/" + p.Poster.LargeImage
	}

	return &domain.Manga{
		Title:       p.Title,
		Description: p.Synopsis,
		Status:      strings.ToLower(p.Status),
		Authors:     authors,
		Genres:      genres,
		CoverURL:    coverURL,
		Sources:     map[string]string{"atsu": id},
		ScrapedAt:   time.Now(),
	}, nil
}

func (a *Adapter) FetchChapterList(ctx context.Context, mangaID string) ([]domain.Chapter, error) {
	raw, err := a.fetchRawChapters(ctx, mangaID)
	if err != nil {
		return nil, err
	}

	allZeroPages := len(raw) > 0
	for _, ch := range raw {
		if ch.PageCount > 0 {
			allZeroPages = false
			break
		}
	}
	// wasExpanded tracks whether we fetched chapters through session/scanlation
	// sub-IDs. When true, ScanlationMangaID distinguishes chapters from different
	// sessions and must be part of the dedup key. When false (direct chapters),
	// dedup by sort key alone to avoid duplicates when ScanlationMangaID varies.
	wasExpanded := false
	if allZeroPages {
		var expanded []rawChapterEntry
		for _, container := range raw {
			sub, err := a.fetchRawChapters(ctx, container.ID)
			if err != nil {
				return nil, err
			}
			expanded = append(expanded, sub...)
		}
		raw = expanded
		wasExpanded = true
	} else {
		subIDs := make(map[string]struct{})
		for _, ch := range raw {
			if ch.ScanlationMangaID != "" && ch.ScanlationMangaID != mangaID {
				subIDs[ch.ScanlationMangaID] = struct{}{}
			}
		}
		if len(subIDs) > 0 && len(subIDs) == len(raw) {
			var expanded []rawChapterEntry
			for subID := range subIDs {
				sub, err := a.fetchRawChapters(ctx, subID)
				if err != nil {
					return nil, err
				}
				expanded = append(expanded, sub...)
			}
			raw = expanded
			wasExpanded = true
		}
	}

	numberFreq := make(map[float64]int, len(raw))
	for _, ch := range raw {
		numberFreq[ch.Number]++
	}
	useSeasonEncoding := false
	for _, count := range numberFreq {
		if count > 1 {
			useSeasonEncoding = true
			break
		}
	}

	// resolveNumber returns the float64 sort key. For season-encoded mangas
	// it parses the chapter from the Title and offsets by season*10000 so that
	// chapters from different seasons never collide (e.g. S2-Ch5 → 20005).
	resolveNumber := func(ch rawChapterEntry) float64 {
		if !useSeasonEncoding {
			return ch.Number
		}
		m := chapterNumRe.FindStringSubmatch(ch.Title)
		if len(m) < 2 {
			return ch.Number
		}
		n, err := strconv.ParseFloat(m[1], 64)
		if err != nil {
			return ch.Number
		}
		return ch.Number*10000 + n
	}

	sort.Slice(raw, func(i, j int) bool {
		ni := resolveNumber(raw[i])
		nj := resolveNumber(raw[j])
		if ni != nj {
			return ni < nj
		}
		return raw[i].CreatedAt < raw[j].CreatedAt
	})

	type dedupKey struct {
		mangaID string
		sortKey float64
	}
	seen := make(map[dedupKey]struct{})
	var chapters []domain.Chapter
	for _, ch := range raw {
		sortKey := resolveNumber(ch)
		// For expanded (session) manga, ScanlationMangaID distinguishes chapters
		// from different scanlation groups — include it so same-numbered chapters
		// from different sessions are kept. For direct (no-session) manga, dedup
		// by sort key alone; including ScanlationMangaID would let the same chapter
		// slip through when different sub-IDs are present.
		sessionID := ""
		if wasExpanded {
			sessionID = ch.ScanlationMangaID
		}
		key := dedupKey{sessionID, sortKey}
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		var numStr string
		if useSeasonEncoding {
			numStr = ch.Title
		} else {
			numStr = strconv.FormatFloat(ch.Number, 'f', -1, 64)
		}
		chapters = append(chapters, domain.Chapter{
			Number:    numStr,
			SortKey:   sortKey,
			Title:     ch.Title,
			Language:  "en",
			Source:    "atsu",
			SourceID:  ch.ScanlationMangaID + ":" + ch.ID,
			ScrapedAt: time.Now(),
		})
	}
	return chapters, nil
}

type rawChapterEntry struct {
	ID                string
	ScanlationMangaID string
	Title             string
	Number            float64
	CreatedAt         int64
	PageCount         int
}

func (a *Adapter) fetchRawChapters(ctx context.Context, mangaID string) ([]rawChapterEntry, error) {
	u := fmt.Sprintf("%s/manga/allChapters?mangaId=%s", a.apiBase, mangaID)
	var resp allChaptersResp
	if err := a.get(ctx, u, &resp); err != nil {
		return nil, fmt.Errorf("atsu fetch chapters: %w", err)
	}
	out := make([]rawChapterEntry, len(resp.Chapters))
	for i, ch := range resp.Chapters {
		scanlationID := ch.ScanlationMangaID
		if scanlationID == "" {
			scanlationID = mangaID
		}
		out[i] = rawChapterEntry{
			ID:                ch.ID,
			ScanlationMangaID: scanlationID,
			Title:             ch.Title,
			Number:            ch.Number,
			CreatedAt:         ch.CreatedAt,
			PageCount:         ch.PageCount,
		}
	}
	return out, nil
}

func (a *Adapter) FetchPageURLs(ctx context.Context, sourceID string) ([]string, error) {
	parts := strings.SplitN(sourceID, ":", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("atsu: invalid sourceID %q", sourceID)
	}
	mangaID, chapterID := parts[0], parts[1]

	u := fmt.Sprintf("%s/read/chapter?mangaId=%s&chapterId=%s", a.apiBase, mangaID, chapterID)
	var resp readChapterResp
	if err := a.get(ctx, u, &resp); err != nil {
		return nil, fmt.Errorf("atsu fetch pages: %w", err)
	}

	pages := resp.ReadChapter.Pages
	sort.Slice(pages, func(i, j int) bool {
		return pages[i].Number < pages[j].Number
	})

	urls := make([]string, len(pages))
	for i, p := range pages {
		if strings.HasPrefix(p.Image, "http") {
			urls[i] = p.Image
		} else {
			urls[i] = a.baseURL + p.Image
		}
	}
	return urls, nil
}
