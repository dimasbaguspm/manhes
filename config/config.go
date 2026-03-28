package config

import (
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Env         string // dev | prod
	ListenAddr  string
	LogLevel    string
	LibraryPath string

	IngestInterval            time.Duration
	IngestWorkers             int
	SyncInterval              time.Duration
	DictionaryRefreshInterval time.Duration

	DownloaderTimeout time.Duration

	Database DatabaseConfig
	S3       S3Config
	Kafka    KafkaConfig
	Mangadex MangadexConfig
	Atsu     AtsuConfig
}

// Load reads configuration from environment variables, pre-loaded from .env.
// The file is optional — if absent, only actual env vars are used.
func Load() (*Config, error) {
	if err := godotenv.Load(".env"); err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("load .env: %w", err)
	}

	return &Config{
		Env:         envStr("APP_ENV", "dev"),
		ListenAddr:  ":8080",
		LogLevel:    "info",
		LibraryPath: "./library",

		IngestInterval:            envDuration("INGEST_INTERVAL", 30*time.Minute),
		IngestWorkers:             envInt("INGEST_WORKERS", 5),
		SyncInterval:              envDuration("SYNC_INTERVAL", 15*time.Minute),
		DictionaryRefreshInterval: envDuration("DICTIONARY_REFRESH_INTERVAL", 4*time.Hour),

		DownloaderTimeout: envDuration("DOWNLOADER_TIMEOUT", 30*time.Second),

		Database: loadDatabaseConfig(),
		S3:       loadS3Config(),
		Kafka:    loadKafkaConfig(),
		Mangadex: loadMangadexConfig(),
		Atsu:     loadAtsuConfig(),
	}, nil
}
