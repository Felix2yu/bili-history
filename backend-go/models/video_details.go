package models

// VideoBaseInfo 视频基本信息模型
type VideoBaseInfo struct {
	ID            int64  `json:"id" db:"id"`
	Bvid          string `json:"bvid" db:"bvid"`
	Aid           int    `json:"aid" db:"aid"`
	Videos        int    `json:"videos" db:"videos"`
	Tid           int    `json:"tid" db:"tid"`
	Tname         string `json:"tname" db:"tname"`
	Copyright     int    `json:"copyright" db:"copyright"`
	Pic           string `json:"pic" db:"pic"`
	Title         string `json:"title" db:"title"`
	Pubdate       int64  `json:"pubdate" db:"pubdate"`
	Ctime         int64  `json:"ctime" db:"ctime"`
	Desc          string `json:"desc" db:"desc"`
	Duration      int    `json:"duration" db:"duration"`
	Cid           int    `json:"cid" db:"cid"`
	OwnerMid      int    `json:"owner_mid" db:"owner_mid"`
	OwnerName     string `json:"owner_name" db:"owner_name"`
	OwnerFace     string `json:"owner_face" db:"owner_face"`
	StatView      int    `json:"stat_view" db:"stat_view"`
	StatDanmaku   int    `json:"stat_danmaku" db:"stat_danmaku"`
	StatReply     int    `json:"stat_reply" db:"stat_reply"`
	StatFavorite  int    `json:"stat_favorite" db:"stat_favorite"`
	StatCoin      int    `json:"stat_coin" db:"stat_coin"`
	StatShare     int    `json:"stat_share" db:"stat_share"`
	StatLike      int    `json:"stat_like" db:"stat_like"`
	FetchTime     int64  `json:"fetch_time" db:"fetch_time"`
	UpdateTime    int64  `json:"update_time" db:"update_time"`
}

// UploaderInfo UP主信息模型
type UploaderInfo struct {
	Mid           int    `json:"mid" db:"mid"`
	Name          string `json:"name" db:"name"`
	Sex           string `json:"sex" db:"sex"`
	Face          string `json:"face" db:"face"`
	Sign          string `json:"sign" db:"sign"`
	Level         int    `json:"level" db:"level"`
	Fans          int    `json:"fans" db:"fans"`
	Attention     int    `json:"attention" db:"attention"`
	ArchiveCount  int    `json:"archive_count" db:"archive_count"`
	FetchTime     int64  `json:"fetch_time" db:"fetch_time"`
	UpdateTime    int64  `json:"update_time" db:"update_time"`
}

// VideoTag 视频标签模型
type VideoTag struct {
	ID      int64  `json:"id" db:"id"`
	Bvid    string `json:"bvid" db:"bvid"`
	TagID   int    `json:"tag_id" db:"tag_id"`
	TagName string `json:"tag_name" db:"tag_name"`
}

// VideoDetailStats 视频详情统计
type VideoDetailStats struct {
	TotalVideos      int64 `json:"total_videos"`
	FetchedVideos    int64 `json:"fetched_videos"`
	TotalUploaders   int64 `json:"total_uploaders"`
	TotalTags        int64 `json:"total_tags"`
}

// UploaderStats UP主统计信息
type UploaderStats struct {
	Mid          int    `json:"mid"`
	Name         string `json:"name"`
	Face         string `json:"face"`
	VideoCount   int64  `json:"video_count"`
	TotalViews   int64  `json:"total_views"`
	TotalLikes   int64  `json:"total_likes"`
	TotalCoins   int64  `json:"total_coins"`
	TotalFavorites int64 `json:"total_favorites"`
}

// VideoDetailProgress 视频详情获取进度
type VideoDetailProgress struct {
	IsProcessing    bool   `json:"is_processing"`
	IsComplete      bool   `json:"is_complete"`
	IsStopped       bool   `json:"is_stopped"`
	TotalVideos     int    `json:"total_videos"`
	ProcessedVideos int    `json:"processed_videos"`
	SuccessCount    int    `json:"success_count"`
	FailedCount     int    `json:"failed_count"`
	StartTime       int64  `json:"start_time"`
	LastUpdateTime  int64  `json:"last_update_time"`
	Status          string `json:"status"`
	ErrorMessage    string `json:"error_message,omitempty"`
}

// BatchFetchRequest 批量获取请求
type BatchFetchRequest struct {
	Bvids []string `json:"bvids" binding:"required"`
}
