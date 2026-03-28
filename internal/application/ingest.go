package application

import (
	"context"
	"fmt"
	"log/slog"
	"mime"
	"os"
	"path/filepath"
	"sync"

	"github.com/google/uuid"

	"manga-engine/config"
	"manga-engine/internal/domain"
)

// IngestServiceConfig holds dependencies for IngestService.
type IngestServiceConfig struct {
	Repo     domain.Repository
	Registry domain.SourceRegistry
	DL       domain.Downloader
	Disk     domain.Storer
	S3       domain.ObjectStore
	Bus      domain.EventBus
	Cfg      *config.Config
}

// IngestService handles the 3-stage pipeline: download → upload → DB write.
type IngestService struct {
	repo     domain.Repository
	registry domain.SourceRegistry
	dl       domain.Downloader
	disk     domain.Storer
	s3c      domain.ObjectStore
	bus      domain.EventBus
	cfg      *config.Config
	log      *slog.Logger
}

func NewIngestService(cfg IngestServiceConfig) *IngestService {
	return &IngestService{
		repo:     cfg.Repo,
		registry: cfg.Registry,
		dl:       cfg.DL,
		disk:     cfg.Disk,
		s3c:      cfg.S3,
		bus:      cfg.Bus,
		cfg:      cfg.Cfg,
		log:      slog.With("service", "ingest"),
	}
}

// Ingest processes an IngestRequested event: sets state to fetching and runs the pipeline.
func (s *IngestService) Ingest(ctx context.Context, e domain.IngestRequested) error {
	// Set manga state to fetching
	if err := s.repo.SetDictionaryStateBySlug(ctx, e.Slug, domain.StateFetching); err != nil {
		s.log.Warn("ingest: set state fetching", "slug", e.Slug, "err", err)
	}

	sources := s.registry.Ordered()
	if len(sources) == 0 {
		return fmt.Errorf("no sources registered")
	}

	// Gather manga detail and chapter list from the best available source
	var primaryManga *domain.Manga
	mergedAvail := make(map[string]int)

	for _, src := range sources {
		sourceID, ok := e.Sources[src.Source()]
		if !ok {
			continue
		}

		manga, err := src.FetchMangaDetail(ctx, sourceID)
		if err != nil {
			s.log.Warn("ingest: fetch detail failed", "source", src.Source(), "slug", e.Slug, "err", err)
			continue
		}
		manga.Slug = e.Slug

		chapters, err := src.FetchChapterList(ctx, sourceID)
		if err != nil {
			s.log.Warn("ingest: fetch chapters failed", "source", src.Source(), "slug", e.Slug, "err", err)
			continue
		}

		// Group by language, respecting LangToSource
		byLang := make(map[string][]domain.Chapter)
		for _, ch := range chapters {
			if len(e.LangToSource) > 0 {
				if bestSrc, ok := e.LangToSource[ch.Language]; ok && bestSrc != src.Source() {
					continue
				}
			}
			ch.MangaSlug = e.Slug
			byLang[ch.Language] = append(byLang[ch.Language], ch)
		}

		availByLang := make(map[string]int)
		for lang, chs := range byLang {
			availByLang[lang] = len(chs)
			if availByLang[lang] > mergedAvail[lang] {
				mergedAvail[lang] = availByLang[lang]
			}
		}

		if primaryManga == nil {
			primaryManga = manga
		}

		// Run the 3-stage pipeline for this source's chapters
		if err := s.runPipeline(ctx, e.DictionaryID, e.Slug, byLang); err != nil {
			s.log.Warn("ingest: pipeline error", "source", src.Source(), "slug", e.Slug, "err", err)
		}

		break // Process one source at a time for now
	}

	// Stamp manga metadata
	if primaryManga != nil {
		if err := s.repo.UpsertManga(ctx, domain.Manga{
			Slug:        e.Slug,
			Title:       primaryManga.Title,
			Description: primaryManga.Description,
			Status:      primaryManga.Status,
			Authors:     primaryManga.Authors,
			Genres:      primaryManga.Genres,
		}); err != nil {
			s.log.Warn("ingest: upsert manga", "slug", e.Slug, "err", err)
		}
	}

	// Check if all chapters are ingested → state available
	if err := s.checkAndSetAvailable(ctx, e.Slug, mergedAvail); err != nil {
		s.log.Warn("ingest: check available", "slug", e.Slug, "err", err)
	}

	return nil
}

// Pipeline job types (defined at package level)
type dlJob struct {
	dictionaryID string
	lang         string
	ch           domain.Chapter
	pages        []string
	localPaths   []string
}

type upJob struct {
	dl     *dlJob
	key    string
	s3URLs []string
}

type dbJob struct {
	up *upJob
}

type chapterResult struct {
	slug  string
	lang  string
	chap  string
	pages int
	err   error
}

// runPipeline runs the 3-stage download → upload → DB write pipeline.
func (s *IngestService) runPipeline(ctx context.Context, dictionaryID, slug string, byLang map[string][]domain.Chapter) error {
	// Collect all download jobs
	jobs := make([]*dlJob, 0)
	for lang, chapters := range byLang {
		for _, ch := range chapters {
			pages := ch.PageURLs
			if len(pages) == 0 {
				continue
			}
			jobs = append(jobs, &dlJob{dictionaryID: dictionaryID, lang: lang, ch: ch, pages: pages})
		}
	}

	if len(jobs) == 0 {
		return nil
	}

	downloadCh := make(chan *dlJob, len(jobs))
	uploadCh := make(chan *upJob, len(jobs))
	dbCh := make(chan *dbJob, len(jobs))
	doneCh := make(chan *chapterResult, len(jobs))

	for _, j := range jobs {
		downloadCh <- j
	}
	close(downloadCh)

	var wg sync.WaitGroup

	// Stage 1: Download (4 concurrent)
	dlSem := make(chan struct{}, 4)
	wg.Add(1)
	go func() {
		defer wg.Done()
		for job := range downloadCh {
			dlSem <- struct{}{}
			go func(j *dlJob) {
				defer func() { <-dlSem }()
				s.downloadChapter(j)
				uploadCh <- &upJob{dl: j}
			}(job)
		}
	}()

	// Stage 2: Upload (4 concurrent)
	upSem := make(chan struct{}, 4)
	wg.Add(1)
	go func() {
		defer wg.Done()
		for job := range uploadCh {
			upSem <- struct{}{}
			go func(j *upJob) {
				defer func() { <-upSem }()
				s.uploadChapter(j)
				dbCh <- &dbJob{up: j}
			}(job)
		}
	}()

	// Stage 3: DB write (4 concurrent)
	dbSem := make(chan struct{}, 4)
	wg.Add(1)
	go func() {
		defer wg.Done()
		for job := range dbCh {
			dbSem <- struct{}{}
			go func(j *dbJob) {
				defer func() { <-dbSem }()
				doneCh <- s.writeChapter(j)
			}(job)
		}
	}()

	go func() {
		wg.Wait()
		close(uploadCh)
		close(dbCh)
		close(doneCh)
	}()

	var errors []error
	for res := range doneCh {
		if res.err != nil {
			errors = append(errors, res.err)
		}
	}

	if len(errors) > 0 {
		return errors[0]
	}
	return nil
}

func (s *IngestService) downloadChapter(j *dlJob) {
	pages := j.pages
	j.localPaths = make([]string, 0, len(pages))
	for i, u := range pages {
		data, err := s.dl.Download(context.Background(), u)
		if err != nil {
			s.log.Warn("download page failed", "slug", j.ch.MangaSlug, "lang", j.lang, "num", j.ch.Number, "page", i, "err", err)
			continue
		}
		ext := filepath.Ext(u)
		if ext == "" {
			ext = ".jpg"
		}
		path, err := s.disk.SavePage(j.ch.MangaSlug, j.lang, j.ch.Number, i+1, data, ext)
		if err != nil {
			s.log.Warn("save page failed", "slug", j.ch.MangaSlug, "err", err)
			continue
		}
		j.localPaths = append(j.localPaths, path)
	}
}

func (s *IngestService) uploadChapter(j *upJob) {
	key := fmt.Sprintf("%s/%s/%s", j.dl.ch.MangaSlug, j.dl.lang, j.dl.ch.Number)
	j.key = key
	j.s3URLs = make([]string, 0, len(j.dl.localPaths))
	for i, localPath := range j.dl.localPaths {
		data, err := os.ReadFile(localPath)
		if err != nil {
			s.log.Warn("upload chapter: read file", "path", localPath, "err", err)
			continue
		}
		ext := filepath.Ext(localPath)
		pageKey := fmt.Sprintf("%s/%03d%s", key, i+1, ext)
		ct := "image/jpeg"
		if ext != "" {
			if ct2 := mime.TypeByExtension(ext); ct2 != "" {
				ct = ct2
			}
		}
		url, err := s.s3c.Upload(context.Background(), pageKey, data, ct)
		if err != nil {
			s.log.Warn("upload chapter: s3 upload", "key", pageKey, "err", err)
			continue
		}
		j.s3URLs = append(j.s3URLs, url)
	}
}

func (s *IngestService) writeChapter(j *dbJob) *chapterResult {
	res := &chapterResult{
		slug:  j.up.dl.ch.MangaSlug,
		lang:  j.up.dl.lang,
		chap:  j.up.dl.ch.Number,
		pages: len(j.up.dl.pages),
	}

	id := uuid.New().String()
	err := s.repo.UpsertChapter(context.Background(),
		id,
		j.up.dl.ch.MangaSlug,
		j.up.dl.ch.Number,
		int(j.up.dl.ch.SortKey),
		j.up.dl.lang,
		j.up.key,
	)
	if err != nil {
		res.err = fmt.Errorf("upsert chapter: %w", err)
		return res
	}

	s.bus.Publish(context.Background(), s.cfg.Bus.ChapterUploaded, domain.ChapterUploaded{
		DictionaryID: j.up.dl.dictionaryID,
		Slug:         j.up.dl.ch.MangaSlug,
		Language:     j.up.dl.lang,
		ChapterNum:   j.up.dl.ch.Number,
		SortKey:      j.up.dl.ch.SortKey,
		PageCount:    len(j.up.dl.pages),
	})

	return res
}

// checkAndSetAvailable checks if all expected chapters are ingested and sets state to available.
func (s *IngestService) checkAndSetAvailable(ctx context.Context, slug string, expectedByLang map[string]int) error {
	for lang, expected := range expectedByLang {
		count, err := s.repo.GetChapterCountByLang(ctx, slug, lang)
		if err != nil {
			return err
		}
		if count < expected {
			return nil // Not all chapters ingested yet
		}
	}

	// All chapters ingested → set available
	return s.repo.SetDictionaryStateBySlug(ctx, slug, domain.StateAvailable)
}
