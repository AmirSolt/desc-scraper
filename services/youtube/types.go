package youtube

type VideoResult struct {
	videoPrimaryInfoRenderer   VideoPrimaryInfoRenderer
	videoSecondaryInfoRenderer VideoSecondaryInfoRenderer
	compactVideoRenderers      []CompactVideoRenderer
}

type CompactVideoRenderer struct {
	VideoID string `json:"videoId"`
}

type VideoPrimaryInfoRenderer struct {
	Title     Title      `json:"title"`
	ViewCount ViewCount  `json:"viewCount"`
	DateText  SimpleText `json:"dateText"`
}

type VideoSecondaryInfoRenderer struct {
	Owner                 Owner   `json:"owner"`
	AttributedDescription Content `json:"attributedDescription"`
}

type Title struct {
	Runs []TextRun `json:"runs"`
}

type Content struct {
	Content string `json:"content"`
}

type TextRun struct {
	Text string `json:"text"`
}

type ViewCount struct {
	VideoViewCountRenderer VideoViewCountRenderer `json:"videoViewCountRenderer"`
}

type VideoViewCountRenderer struct {
	ViewCount         SimpleText `json:"viewCount"`
	ShortViewCount    SimpleText `json:"shortViewCount"`
	OriginalViewCount string     `json:"originalViewCount"`
}

type SimpleText struct {
	SimpleText string `json:"simpleText"`
}

type Owner struct {
	VideoOwnerRenderer VideoOwnerRenderer `json:"videoOwnerRenderer"`
}

type VideoOwnerRenderer struct {
	Thumbnail Thumbnail `json:"thumbnail"`
	Title     TitleRun  `json:"title"`
}

type Thumbnail struct {
	Thumbnails []ThumbnailItem `json:"thumbnails"`
}

type ThumbnailItem struct {
	URL    string `json:"url"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

type TitleRun struct {
	Runs []Run `json:"runs"`
}

type Run struct {
	Text               string             `json:"text"`
	NavigationEndpoint NavigationEndpoint `json:"navigationEndpoint"`
}

type NavigationEndpoint struct {
	ClickTrackingParams string         `json:"clickTrackingParams"`
	BrowseEndpoint      BrowseEndpoint `json:"browseEndpoint"`
}

type BrowseEndpoint struct {
	BrowseId         string `json:"browseId"`
	CanonicalBaseUrl string `json:"canonicalBaseUrl"`
}
