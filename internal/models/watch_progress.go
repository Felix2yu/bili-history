package models

// WatchProgress is the compact view of a video's viewing state used by external
// clients (for example the 满屏 macOS player) to decorate video covers.
type WatchProgress struct {
	Bvid     string `json:"bvid"`
	Watched  bool   `json:"watched"`
	Progress int    `json:"progress"` // seconds already watched; -1 is normalized to duration
	Duration int    `json:"duration"` // seconds, 0 when the history row has none
	Finished bool   `json:"finished"`
	ViewAt   int64  `json:"view_at"`
}
