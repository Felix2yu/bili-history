package models

// The types in this file are the wire contract of the external client channel
// (/api/client/*), used by the 满屏 macOS player. They are deliberately small:
// a client drawing a list only needs what fits on a row, and every extra field
// is another thing that can disagree with the web UI.

// ClientVideo is one archive history row or one 继续观看 entry.
type ClientVideo struct {
	Bvid      string `json:"bvid"`
	Title     string `json:"title"`
	Cover     string `json:"cover"`
	Author    string `json:"author"`
	AuthorMid int64  `json:"author_mid"`
	Duration  int    `json:"duration"`
	// Progress is seconds watched; a completed play is normalized to Duration so
	// the client never has to know that -1 means finished.
	Progress int    `json:"progress"`
	ViewAt   int64  `json:"view_at"`
	Finished bool   `json:"finished"`
	Business string `json:"business,omitempty"`
}

// ClientHistoryPage is one cursor page of the archive. Next is 0 when the last
// page has been reached, so a client stops without an extra request.
type ClientHistoryPage struct {
	Records []ClientVideo `json:"records"`
	Next    int64         `json:"next"`
}

// ClientSummary is the compact "what did I watch" strip. Every number uses the
// same definition as the web monthly report, so the two never disagree:
// seconds is the sum of the videos' lengths, not the time spent watching.
type ClientSummary struct {
	Year           int     `json:"year"`
	Month          int     `json:"month"` // 0 means the whole year
	Videos         int     `json:"videos"`
	Seconds        int64   `json:"seconds"`
	UniqueAuthors  int     `json:"unique_authors"`
	UniqueDays     int     `json:"unique_days"`
	Finished       int     `json:"finished"`
	FinishedRate   float64 `json:"finished_rate"`
	LateNightRatio float64 `json:"late_night_ratio"`
}

// ClientWatchLaterItem mirrors one archived 稍后再看 entry.
type ClientWatchLaterItem struct {
	Bvid     string `json:"bvid"`
	Title    string `json:"title"`
	Cover    string `json:"cover"`
	Author   string `json:"author"`
	Duration int    `json:"duration"`
	AddAt    int64  `json:"add_at"`
	View     int    `json:"view"`
}

// ClientFavoriteFolder mirrors one archived 收藏夹.
type ClientFavoriteFolder struct {
	MediaID    int64  `json:"media_id"`
	Title      string `json:"title"`
	Cover      string `json:"cover"`
	MediaCount int    `json:"media_count"`
	Mtime      int64  `json:"mtime"`
}

// ClientFavoriteItem mirrors one archived 收藏内容.
type ClientFavoriteItem struct {
	Bvid     string `json:"bvid"`
	Title    string `json:"title"`
	Cover    string `json:"cover"`
	Author   string `json:"author"`
	UpperMid int64  `json:"author_mid"`
	Duration int    `json:"duration"`
	FavTime  int64  `json:"fav_time"`
}
