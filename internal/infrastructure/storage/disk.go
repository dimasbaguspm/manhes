package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"manga-engine/internal/domain"
)

var _ domain.Storer = (*DiskStore)(nil)

type DiskStore struct {
	root string
}

func NewDisk(root string) *DiskStore {
	return &DiskStore{root: root}
}

func (d *DiskStore) SavePage(slug, lang string, chapterNum float64, pageIdx int, data []byte, ext string) (string, error) {
	dir := filepath.Join(d.root, slug, lang, ChapterDir(chapterNum))
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("mkdir: %w", err)
	}
	path := filepath.Join(dir, PageFile(pageIdx, ext))
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return "", fmt.Errorf("write page: %w", err)
	}
	return path, nil
}

func (d *DiskStore) SaveCover(slug string, data []byte, ext string) (string, error) {
	dir := filepath.Join(d.root, slug)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("mkdir: %w", err)
	}
	path := filepath.Join(dir, "cover"+ext)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return "", fmt.Errorf("write cover: %w", err)
	}
	return path, nil
}

func (d *DiskStore) WriteMetadata(slug string, m *domain.Metadata) error {
	dir := filepath.Join(d.root, slug)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("mkdir: %w", err)
	}
	return writeJSON(filepath.Join(dir, "metadata.json"), m)
}

func (d *DiskStore) ReadMetadata(slug string) (*domain.Metadata, error) {
	path := filepath.Join(d.root, slug, "metadata.json")
	f, err := os.Open(path)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var m domain.Metadata
	if err := json.NewDecoder(f).Decode(&m); err != nil {
		return nil, err
	}
	return &m, nil
}

func (d *DiskStore) WriteLangMetadata(slug, lang string, m *domain.LangMetadata) error {
	dir := filepath.Join(d.root, slug, lang)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("mkdir: %w", err)
	}
	return writeJSON(filepath.Join(dir, "metadata.json"), m)
}

func (d *DiskStore) ReadLangMetadata(slug, lang string) (*domain.LangMetadata, error) {
	path := filepath.Join(d.root, slug, lang, "metadata.json")
	f, err := os.Open(path)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var m domain.LangMetadata
	if err := json.NewDecoder(f).Decode(&m); err != nil {
		return nil, err
	}
	return &m, nil
}

func (d *DiskStore) WriteChapterManifest(slug, lang string, ch *domain.Chapter) error {
	dir := filepath.Join(d.root, slug, lang, ChapterDir(ch.Number))
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("mkdir: %w", err)
	}
	return writeJSON(filepath.Join(dir, "chapter.json"), ch)
}

func writeJSON(path string, v any) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}
