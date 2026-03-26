package mangadex

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"manga-engine/internal/domain"
)

func (a *Adapter) FetchMangaDetail(ctx context.Context, id string) (*domain.Manga, error) {
	u := fmt.Sprintf("%s/manga/%s?includes[]=author&includes[]=cover_art", a.baseURL, id)
	var resp mangaDetailResp
	if err := a.get(ctx, u, &resp); err != nil {
		return nil, fmt.Errorf("mangadex fetch manga: %w", err)
	}

	attr := resp.Data.Attributes
	title := firstOf(attr.Title, "en", "ja-ro", "ja")

	var authors []string
	var coverFilename string
	for _, rel := range resp.Data.Relationships {
		switch rel.Type {
		case "author":
			var authorAttr struct {
				Name string `json:"name"`
			}
			if json.Unmarshal(rel.Attributes, &authorAttr) == nil && authorAttr.Name != "" {
				authors = append(authors, authorAttr.Name)
			}
		case "cover_art":
			var coverAttr struct {
				FileName string `json:"fileName"`
			}
			if json.Unmarshal(rel.Attributes, &coverAttr) == nil {
				coverFilename = coverAttr.FileName
			}
		}
	}

	var genres []string
	for _, tag := range attr.Tags {
		if name := firstOf(tag.Attributes.Name, "en"); name != "" {
			genres = append(genres, name)
		}
	}

	coverURL := ""
	if coverFilename != "" {
		coverURL = fmt.Sprintf("%s/%s/%s", coverBaseURL, id, coverFilename)
	}

	return &domain.Manga{
		Title:       title,
		Description: firstOf(attr.Description, "en"),
		Status:      attr.Status,
		Authors:     authors,
		Genres:      genres,
		CoverURL:    coverURL,
		Sources:     map[string]string{"mangadex": id},
		ScrapedAt:   time.Now(),
	}, nil
}

func (a *Adapter) FetchChapterList(ctx context.Context, mangaID string) ([]domain.Chapter, error) {
	const limit = 500
	var all []domain.Chapter
	offset := 0

	for {
		params := url.Values{}
		params.Set("limit", strconv.Itoa(limit))
		params.Set("offset", strconv.Itoa(offset))
		params.Set("order[chapter]", "asc")
		params.Add("contentRating[]", "safe")
		params.Add("contentRating[]", "suggestive")
		params.Add("contentRating[]", "erotica")
		params.Add("contentRating[]", "pornographic")

		u := fmt.Sprintf("%s/manga/%s/feed?%s", a.baseURL, mangaID, params.Encode())
		var resp chapterFeedResp
		if err := a.get(ctx, u, &resp); err != nil {
			return nil, fmt.Errorf("mangadex fetch chapters: %w", err)
		}

		for _, ch := range resp.Data {
			if ch.Attributes.Chapter == nil {
				continue
			}
			num, err := strconv.ParseFloat(*ch.Attributes.Chapter, 64)
			if err != nil {
				continue
			}
			all = append(all, domain.Chapter{
				Number:    num,
				Title:     ch.Attributes.Title,
				Language:  ch.Attributes.TranslatedLanguage,
				Source:    "mangadex",
				SourceID:  ch.ID,
				ScrapedAt: time.Now(),
			})
		}

		offset += len(resp.Data)
		if offset >= resp.Total {
			break
		}
	}

	seen := make(map[float64]struct{})
	deduped := all[:0]
	for _, ch := range all {
		if _, ok := seen[ch.Number]; ok {
			continue
		}
		seen[ch.Number] = struct{}{}
		deduped = append(deduped, ch)
	}
	return deduped, nil
}

func (a *Adapter) FetchPageURLs(ctx context.Context, chapterID string) ([]string, error) {
	u := fmt.Sprintf("%s/at-home/server/%s", a.baseURL, chapterID)
	var resp atHomeResp
	if err := a.get(ctx, u, &resp); err != nil {
		return nil, fmt.Errorf("mangadex fetch pages: %w", err)
	}

	urls := make([]string, len(resp.Chapter.Data))
	for i, filename := range resp.Chapter.Data {
		urls[i] = fmt.Sprintf("%s/data/%s/%s", resp.BaseURL, resp.Chapter.Hash, filename)
	}
	return urls, nil
}
