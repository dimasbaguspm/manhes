package domain

import (
	"encoding/json"
	"time"
)

// Tracker represents a reading progress entry for a manga chapter.
type Tracker struct {
	ID        string          `json:"id"`
	MangaID   string          `json:"manga_id"`
	ChapterID string          `json:"chapter_id"`
	IsRead    bool            `json:"is_read"`
	Metadata  json.RawMessage `json:"metadata"`
	UpdatedAt time.Time       `json:"updated_at"`
	CreatedAt time.Time       `json:"created_at"`
}

// UpsertTrackerRequest is the request body for PUT /api/v1/tracker.
type UpsertTrackerRequest struct {
	ID        string          `json:"id"`
	MangaID   string          `json:"manga_id"`
	ChapterID string          `json:"chapter_id"`
	IsRead    bool            `json:"is_read"`
	Metadata  json.RawMessage `json:"metadata"`
}

// TrackerResponse is the response for tracker endpoints.
type TrackerResponse struct {
	ID        string `json:"id"`
	MangaID   string `json:"manga_id"`
	ChapterID string `json:"chapter_id"`
	IsRead    bool   `json:"is_read"`
	Metadata  string `json:"metadata"` // JSON string, use json.RawMessage on the domain Tracker
	UpdatedAt time.Time `json:"updated_at"`
	CreatedAt time.Time `json:"created_at"`
}
