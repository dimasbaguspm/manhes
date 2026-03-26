package mangadex

import "encoding/json"

type mangaDetailResp struct {
	Data struct {
		ID         string `json:"id"`
		Attributes struct {
			Title       map[string]string `json:"title"`
			Description map[string]string `json:"description"`
			Status      string            `json:"status"`
			Tags        []struct {
				Attributes struct {
					Name map[string]string `json:"name"`
				} `json:"attributes"`
			} `json:"tags"`
		} `json:"attributes"`
		Relationships []relationship `json:"relationships"`
	} `json:"data"`
}

type relationship struct {
	Type       string          `json:"type"`
	Attributes json.RawMessage `json:"attributes"`
}

type chapterFeedResp struct {
	Data []struct {
		ID         string `json:"id"`
		Attributes struct {
			Chapter            *string `json:"chapter"`
			Title              string  `json:"title"`
			TranslatedLanguage string  `json:"translatedLanguage"`
			Pages              int     `json:"pages"`
		} `json:"attributes"`
	} `json:"data"`
	Total  int `json:"total"`
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

type atHomeResp struct {
	BaseURL string `json:"baseUrl"`
	Chapter struct {
		Hash      string   `json:"hash"`
		Data      []string `json:"data"`
		DataSaver []string `json:"dataSaver"`
	} `json:"chapter"`
}

type searchResp struct {
	Data []struct {
		ID         string `json:"id"`
		Attributes struct {
			Title  map[string]string `json:"title"`
			Status string            `json:"status"`
			Tags   []struct {
				Attributes struct {
					Name map[string]string `json:"name"`
				} `json:"attributes"`
			} `json:"tags"`
		} `json:"attributes"`
	} `json:"data"`
}
