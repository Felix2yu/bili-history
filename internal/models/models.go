package models

import "time"

// HistoryRecord represents a single bilibili watch history entry.
// Schema matches Python's bilibili_history_YYYY table exactly.
type HistoryRecord struct {
	ID         int64  `json:"id" db:"id"`
	BVID       string `json:"bvid" db:"bvid"`
	AID        int64  `json:"aid" db:"aid"`
	Title      string `json:"title" db:"title"`
	Desc       string `json:"desc" db:"desc"`
	Pic        string `json:"pic" db:"pic"`
	Duration   int    `json:"duration" db:"duration"`
	OwnerName  string `json:"owner_name" db:"owner_name"`
	OwnerMid   int64  `json:"owner_mid" db:"owner_mid"`
	TagName    string `json:"tag_name" db:"tag_name"`
	Tid        int    `json:"tid" db:"tid"`
	ViewAt     int64  `json:"view_at" db:"view_at"`
	Progress   int    `json:"progress" db:"progress"`
	Business   string `json:"business" db:"business"`
	View       int    `json:"view" db:"view"`
	Danmaku    int    `json:"danmaku" db:"danmaku"`
	Coin       int    `json:"coin" db:"coin"`
	Favorite   int    `json:"favorite" db:"favorite"`
	Like       int    `json:"like" db:"like"`
	Reply      int    `json:"reply" db:"reply"`
	Share      int    `json:"share" db:"share"`

	// Aliases for backward compatibility with JSON output
	Thumbnail  string `json:"thumbnail,omitempty" db:"-"`
	OwnerID    int64  `json:"owner_id,omitempty" db:"-"`
	Cid        int64  `json:"cid,omitempty" db:"-"`
	Repeat     int    `json:"repeat,omitempty" db:"-"`
	PartName   string `json:"part_name,omitempty" db:"-"`
	PubDate    int64  `json:"pub_date,omitempty" db:"-"`
	CategoryID int    `json:"category_id,omitempty" db:"-"`
}

// VideoDetail represents extended video information.
type VideoDetail struct {
	BVID         string `json:"bvid"`
	Title        string `json:"title"`
	Desc         string `json:"desc"`
	Duration     int    `json:"duration"`
	PubDate      int64  `json:"pub_date"`
	View         int    `json:"view"`
	Danmaku      int    `json:"danmaku"`
	Reply        int    `json:"reply"`
	Favorite     int    `json:"favorite"`
	Coin         int    `json:"coin"`
	Share        int    `json:"share"`
	Like         int    `json:"like"`
	OwnerID      int64  `json:"owner_id"`
	OwnerName    string `json:"owner_name"`
	OwnerFace    string `json:"owner_face"`
	Tid          int    `json:"tid"`
	TName        string `json:"t_name"`
	Pic          string `json:"pic"`
}

// DailyCount represents daily watch counts.
type DailyCount struct {
	Date  string `json:"date"`
	Count int    `json:"count"`
}

// MonthlyCount represents monthly watch counts.
type MonthlyCount struct {
	Month string `json:"month"`
	Count int    `json:"count"`
}

// HealthResponse is returned by the /health endpoint.
type HealthResponse struct {
	Status          string `json:"status"`
	Timestamp       string `json:"timestamp"`
	SchedulerStatus string `json:"scheduler_status"`
}

// QRCodeResponse represents a QR code login response.
type QRCodeResponse struct {
	URL    string `json:"url"`
	Key    string `json:"key"`
	ImgURL string `json:"img_url"`
}

// LoginPollResponse represents the polling result of QR code login.
type LoginPollResponse struct {
	Status  string            `json:"status"` // "pending", "success", "expired"
	Message string            `json:"message"`
	Cookies map[string]string `json:"cookies,omitempty"`
}

// SSEEvent is a Server-Sent Event.
type SSEEvent struct {
	Event string
	Data  interface{}
}

// Task represents a scheduler task.
type Task struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Endpoint  string    `json:"endpoint"`
	Method    string    `json:"method"`
	Schedule  string    `json:"schedule"`
	Params    string    `json:"params"`
	Requires  string    `json:"requires"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TaskExecution represents a single task execution record.
type TaskExecution struct {
	ID        int64     `json:"id"`
	TaskID    int64     `json:"task_id"`
	TaskName  string    `json:"task_name"`
	Status    string    `json:"status"` // "running", "success", "failed"
	StartTime time.Time `json:"start_time"`
	EndTime   *time.Time `json:"end_time,omitempty"`
	Error     string    `json:"error,omitempty"`
	Result    string    `json:"result,omitempty"`
}

// AnalysisResult holds aggregated analysis data.
type AnalysisResult struct {
	TotalVideos   int            `json:"total_videos"`
	TotalDuration int64          `json:"total_duration"`
	DailyCounts   []DailyCount   `json:"daily_counts"`
	MonthlyCounts []MonthlyCount `json:"monthly_counts"`
}
