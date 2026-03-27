package persistence

import (
	"database/sql"
	_ "embed"
	"fmt"

	"github.com/google/uuid"
	_ "modernc.org/sqlite"
)

//go:embed schema.sql
var schema string

// SQLiteRepository implements domain.Repository using SQLite.
type SQLiteRepository struct {
	db *sql.DB
}

// NewSQLite opens (or creates) a SQLite DB at path and applies the schema and migrations.
func NewSQLite(path string) (*SQLiteRepository, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	db.Exec("PRAGMA journal_mode=WAL")
	db.Exec("PRAGMA busy_timeout = 5000")
	db.Exec("PRAGMA foreign_keys=ON")
	if _, err := db.Exec(schema); err != nil {
		return nil, fmt.Errorf("migrate: %w", err)
	}

	// Migrate chapter tables from REAL to TEXT chapter_num (v2 schema).
	var chNumType string
	db.QueryRow(`SELECT type FROM pragma_table_info('ingest_chapters') WHERE name='chapter_num'`).Scan(&chNumType)
	if chNumType == "REAL" || chNumType == "" {
		db.Exec(`DROP TABLE IF EXISTS chapter_pages`)
		db.Exec(`DROP TABLE IF EXISTS manga_chapters`)
		db.Exec(`DROP TABLE IF EXISTS ingest_chapters`)
		db.Exec(`CREATE TABLE ingest_chapters (
			slug TEXT NOT NULL, language TEXT NOT NULL, chapter_num TEXT NOT NULL,
			sort_key REAL NOT NULL DEFAULT 0,
			downloaded_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (slug, language, chapter_num))`)
		db.Exec(`CREATE TABLE manga_chapters (
			slug TEXT NOT NULL, language TEXT NOT NULL, chapter_num TEXT NOT NULL,
			sort_key REAL NOT NULL DEFAULT 0,
			page_count INT NOT NULL DEFAULT 0, uploaded BOOLEAN NOT NULL DEFAULT FALSE,
			uploaded_at TEXT, PRIMARY KEY (slug, language, chapter_num),
			FOREIGN KEY (slug, language) REFERENCES manga_langs(slug, language) ON DELETE CASCADE)`)
		db.Exec(`CREATE TABLE chapter_pages (
			slug TEXT NOT NULL, language TEXT NOT NULL, chapter_num TEXT NOT NULL,
			page_index INT NOT NULL, s3_url TEXT NOT NULL,
			PRIMARY KEY (slug, language, chapter_num, page_index),
			FOREIGN KEY (slug, language, chapter_num)
				REFERENCES manga_chapters(slug, language, chapter_num) ON DELETE CASCADE)`)
	}
	// Idempotent sort_key addition for tables that exist but may be missing it
	db.Exec(`ALTER TABLE manga_chapters ADD COLUMN sort_key REAL NOT NULL DEFAULT 0`)
	db.Exec(`ALTER TABLE ingest_chapters ADD COLUMN sort_key REAL NOT NULL DEFAULT 0`)

	// Idempotent column additions — errors are intentionally ignored (column already exists).
	db.Exec(`ALTER TABLE manga ADD COLUMN uuid        TEXT NOT NULL DEFAULT ''`)
	db.Exec(`ALTER TABLE manga ADD COLUMN description TEXT NOT NULL DEFAULT ''`)
	db.Exec(`ALTER TABLE watchlist ADD COLUMN dictionary_id TEXT REFERENCES dictionary(id)`)
	db.Exec(`ALTER TABLE watchlist ADD COLUMN id TEXT`)
	db.Exec(`ALTER TABLE manga_chapters ADD COLUMN uploaded_at TEXT`)
	db.Exec(`ALTER TABLE dictionary ADD COLUMN cover_url TEXT NOT NULL DEFAULT ''`)
	db.Exec(`UPDATE manga_chapters SET uploaded_at = CURRENT_TIMESTAMP WHERE uploaded = TRUE AND uploaded_at IS NULL`)
	db.Exec(`UPDATE watchlist SET id = lower(hex(randomblob(4)))||'-'||lower(hex(randomblob(2)))||'-4'||substr(lower(hex(randomblob(2))),2)||'-'||substr('89ab',abs(random())%4+1,1)||substr(lower(hex(randomblob(2))),2)||'-'||lower(hex(randomblob(6))) WHERE id IS NULL`)
	db.Exec(`UPDATE manga SET uuid = lower(hex(randomblob(4)))||'-'||lower(hex(randomblob(2)))||'-4'||substr(lower(hex(randomblob(2))),2)||'-'||substr('89ab',abs(random())%4+1,1)||substr(lower(hex(randomblob(2))),2)||'-'||lower(hex(randomblob(6))) WHERE uuid = ''`)
	db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS idx_manga_uuid ON manga(uuid) WHERE uuid != ''`)

	// Performance indexes.
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_manga_chapters_pending ON manga_chapters(slug, language, chapter_num) WHERE uploaded = FALSE`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_dictionary_state ON dictionary(state)`)

	// Back-fill dictionary entries for manga ingested before the dictionary existed.
	rows, err := db.Query(`SELECT slug, title FROM manga WHERE slug NOT IN (SELECT slug FROM dictionary)`)
	if err == nil {
		type row struct{ slug, title string }
		var toInsert []row
		for rows.Next() {
			var r row
			if rows.Scan(&r.slug, &r.title) == nil {
				toInsert = append(toInsert, r)
			}
		}
		rows.Close()
		for _, r := range toInsert {
			db.Exec(`INSERT OR IGNORE INTO dictionary (id, slug, title, sources, source_stats, best_source, state, created_at)
				VALUES (?, ?, ?, '{}', '{}', '{}', 'available', CURRENT_TIMESTAMP)`, uuid.New().String(), r.slug, r.title)
		}
	}

	return &SQLiteRepository{db: db}, nil
}

func (r *SQLiteRepository) Close() error { return r.db.Close() }

func nullStr(s string) any {
	if s == "" {
		return nil
	}
	return s
}
