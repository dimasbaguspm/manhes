package atsu

type mangaPageResp struct {
	MangaPage struct {
		ID       string `json:"id"`
		Title    string `json:"title"`
		Status   string `json:"status"`
		Synopsis string `json:"synopsis"`
		Genres   []struct {
			Name string `json:"name"`
		} `json:"genres"`
		Authors []struct {
			Name string `json:"name"`
		} `json:"authors"`
		Poster *struct {
			LargeImage string `json:"largeImage"`
		} `json:"poster"`
	} `json:"mangaPage"`
}

type chapterItem struct {
	ID                string  `json:"id"`
	ScanlationMangaID string  `json:"scanlationMangaId"`
	Title             string  `json:"title"`
	Number            float64 `json:"number"`
	CreatedAt         int64   `json:"createdAt"`
	Index             int     `json:"index"`
	PageCount         int     `json:"pageCount"`
}

type allChaptersResp struct {
	Chapters []chapterItem `json:"chapters"`
}

type readChapterResp struct {
	ReadChapter struct {
		ID    string `json:"id"`
		Pages []struct {
			Image  string `json:"image"`
			Number int    `json:"number"`
		} `json:"pages"`
	} `json:"readChapter"`
}

type searchResp struct {
	Hits []struct {
		ID    string `json:"id"`
		Title string `json:"title"`
		Type  string `json:"type"`
	} `json:"hits"`
}
