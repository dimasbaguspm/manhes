package config

type MangadexConfig struct {
	BaseURL   string
	RateLimit float64
}

type AtsuConfig struct {
	BaseURL   string
	RateLimit float64
}

func loadMangadexConfig() MangadexConfig {
	return MangadexConfig{
		BaseURL:   envStr("MANGADEX_BASE_URL", "https://api.mangadex.org"),
		RateLimit: envFloat("MANGADEX_RATE_LIMIT", 5),
	}
}

func loadAtsuConfig() AtsuConfig {
	return AtsuConfig{
		BaseURL:   envStr("ATSU_BASE_URL", "https://atsu.moe"),
		RateLimit: envFloat("ATSU_RATE_LIMIT", 3),
	}
}
