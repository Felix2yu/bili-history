package database

import (
	"database/sql"
	"time"
)

// LikesAnalysisResult 点赞分析结果
type LikesAnalysisResult struct {
	TotalCount      int                `json:"total_count"`
	TopCreators     []CreatorCount     `json:"top_creators"`
	CategoryDist    []CategoryCount    `json:"category_dist"`
	DurationDist    []DurationBucket   `json:"duration_dist"`
	AvgViewCount    float64            `json:"avg_view_count"`
	AvgDanmakuCount float64            `json:"avg_danmaku_count"`
}

// FavoritesAnalysisResult 收藏分析结果
type FavoritesAnalysisResult struct {
	TotalCount      int                `json:"total_count"`
	FolderCount     int                `json:"folder_count"`
	FolderDist      []FolderCount      `json:"folder_dist"`
	TopCreators     []CreatorCount     `json:"top_creators"`
	AvgPlayCount    float64            `json:"avg_play_count"`
	AvgCollectCount float64            `json:"avg_collect_count"`
}

// WatchLaterAnalysisResult 稍后再看分析结果
type WatchLaterAnalysisResult struct {
	TotalCount      int                `json:"total_count"`
	CategoryDist    []CategoryCount    `json:"category_dist"`
	DurationDist    []DurationBucket   `json:"duration_dist"`
	AvgViewCount    float64            `json:"avg_view_count"`
	OldestItems     []WatchLaterItem   `json:"oldest_items"`
}

// CreatorCount 创作者计数
type CreatorCount struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

// CategoryCount 分类计数
type CategoryCount struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

// DurationBucket 时长区间
type DurationBucket struct {
	Range string `json:"range"`
	Count int    `json:"count"`
}

// FolderCount 收藏夹计数
type FolderCount struct {
	Title string `json:"title"`
	Count int    `json:"count"`
}

// WatchLaterItem 稍后再看条目
type WatchLaterItem struct {
	Title     string `json:"title"`
	OwnerName string `json:"owner_name"`
	AddAt     int64  `json:"add_at"`
	DaysAgo   int    `json:"days_ago"`
}

// AnalyzeLikes 分析点赞数据
func AnalyzeLikes() (*LikesAnalysisResult, error) {
	db := GetLikesDB()
	if db == nil {
		return &LikesAnalysisResult{}, nil
	}

	result := &LikesAnalysisResult{}

	// 总数
	var totalCount int
	err := db.QueryRow("SELECT COUNT(*) FROM liked_videos").Scan(&totalCount)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	result.TotalCount = totalCount

	// Top 创作者
	rows, err := db.Query(`
		SELECT owner_name, COUNT(*) as cnt 
		FROM liked_videos 
		WHERE owner_name != '' AND owner_name IS NOT NULL
		GROUP BY owner_name 
		ORDER BY cnt DESC 
		LIMIT 10
	`)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var c CreatorCount
			if err := rows.Scan(&c.Name, &c.Count); err == nil {
				result.TopCreators = append(result.TopCreators, c)
			}
		}
	}

	// 分类分布
	rows, err = db.Query(`
		SELECT tname, COUNT(*) as cnt 
		FROM liked_videos 
		WHERE tname != '' AND tname IS NOT NULL
		GROUP BY tname 
		ORDER BY cnt DESC 
		LIMIT 10
	`)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var c CategoryCount
			if err := rows.Scan(&c.Name, &c.Count); err == nil {
				result.CategoryDist = append(result.CategoryDist, c)
			}
		}
	}

	// 时长分布
	rows, err = db.Query(`
		SELECT 
			CASE 
				WHEN duration < 60 THEN '短视频(<1分钟)'
				WHEN duration < 300 THEN '短片(1-5分钟)'
				WHEN duration < 600 THEN '中等(5-10分钟)'
				WHEN duration < 1800 THEN '较长(10-30分钟)'
				ELSE '长视频(30分钟+)'
			END as range,
			COUNT(*) as cnt
		FROM liked_videos
		WHERE duration > 0
		GROUP BY range
		ORDER BY 
			CASE range
				WHEN '短视频(<1分钟)' THEN 1
				WHEN '短片(1-5分钟)' THEN 2
				WHEN '中等(5-10分钟)' THEN 3
				WHEN '较长(10-30分钟)' THEN 4
				WHEN '长视频(30分钟+)' THEN 5
			END
	`)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var d DurationBucket
			if err := rows.Scan(&d.Range, &d.Count); err == nil {
				result.DurationDist = append(result.DurationDist, d)
			}
		}
	}

	// 平均播放和弹幕数
	db.QueryRow(`
		SELECT 
			COALESCE(AVG(CASE WHEN view > 0 THEN view END), 0),
			COALESCE(AVG(CASE WHEN danmaku > 0 THEN danmaku END), 0)
		FROM liked_videos
	`).Scan(&result.AvgViewCount, &result.AvgDanmakuCount)

	return result, nil
}

// AnalyzeFavorites 分析收藏数据
func AnalyzeFavorites() (*FavoritesAnalysisResult, error) {
	db := GetFavoritesDB()
	if db == nil {
		return &FavoritesAnalysisResult{}, nil
	}

	result := &FavoritesAnalysisResult{}

	// 总收藏数
	var totalCount int
	err := db.QueryRow("SELECT COUNT(*) FROM favorites_content").Scan(&totalCount)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	result.TotalCount = totalCount

	// 收藏夹数量
	var folderCount int
	db.QueryRow("SELECT COUNT(*) FROM favorites_folder").Scan(&folderCount)
	result.FolderCount = folderCount

	// 收藏夹分布
	rows, err := db.Query(`
		SELECT f.title, COUNT(c.id) as cnt
		FROM favorites_folder f
		LEFT JOIN favorites_content c ON f.media_id = c.media_id
		GROUP BY f.media_id, f.title
		ORDER BY cnt DESC
		LIMIT 10
	`)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var fc FolderCount
			if err := rows.Scan(&fc.Title, &fc.Count); err == nil {
				result.FolderDist = append(result.FolderDist, fc)
			}
		}
	}

	// Top 创作者
	rows, err = db.Query(`
		SELECT creator_name, COUNT(*) as cnt
		FROM favorites_content
		WHERE creator_name != '' AND creator_name IS NOT NULL
		GROUP BY creator_name
		ORDER BY cnt DESC
		LIMIT 10
	`)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var c CreatorCount
			if err := rows.Scan(&c.Name, &c.Count); err == nil {
				result.TopCreators = append(result.TopCreators, c)
			}
		}
	}

	// 平均播放和收藏数
	db.QueryRow(`
		SELECT 
			COALESCE(AVG(CASE WHEN play > 0 THEN play END), 0),
			COALESCE(AVG(CASE WHEN collect > 0 THEN collect END), 0)
		FROM favorites_content
	`).Scan(&result.AvgPlayCount, &result.AvgCollectCount)

	return result, nil
}

// AnalyzeWatchLater 分析稍后再看数据
func AnalyzeWatchLater() (*WatchLaterAnalysisResult, error) {
	db := GetWatchLaterDB()
	if db == nil {
		return &WatchLaterAnalysisResult{}, nil
	}

	result := &WatchLaterAnalysisResult{}

	// 总数
	var totalCount int
	err := db.QueryRow("SELECT COUNT(*) FROM watchlater_videos").Scan(&totalCount)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	result.TotalCount = totalCount

	// 分类分布
	rows, err := db.Query(`
		SELECT tname, COUNT(*) as cnt
		FROM watchlater_videos
		WHERE tname != '' AND tname IS NOT NULL
		GROUP BY tname
		ORDER BY cnt DESC
		LIMIT 10
	`)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var c CategoryCount
			if err := rows.Scan(&c.Name, &c.Count); err == nil {
				result.CategoryDist = append(result.CategoryDist, c)
			}
		}
	}

	// 时长分布
	rows, err = db.Query(`
		SELECT 
			CASE 
				WHEN duration < 60 THEN '短视频(<1分钟)'
				WHEN duration < 300 THEN '短片(1-5分钟)'
				WHEN duration < 600 THEN '中等(5-10分钟)'
				WHEN duration < 1800 THEN '较长(10-30分钟)'
				ELSE '长视频(30分钟+)'
			END as range,
			COUNT(*) as cnt
		FROM watchlater_videos
		WHERE duration > 0
		GROUP BY range
		ORDER BY 
			CASE range
				WHEN '短视频(<1分钟)' THEN 1
				WHEN '短片(1-5分钟)' THEN 2
				WHEN '中等(5-10分钟)' THEN 3
				WHEN '较长(10-30分钟)' THEN 4
				WHEN '长视频(30分钟+)' THEN 5
			END
	`)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var d DurationBucket
			if err := rows.Scan(&d.Range, &d.Count); err == nil {
				result.DurationDist = append(result.DurationDist, d)
			}
		}
	}

	// 平均播放数
	db.QueryRow(`
		SELECT COALESCE(AVG(CASE WHEN view > 0 THEN view END), 0)
		FROM watchlater_videos
	`).Scan(&result.AvgViewCount)

	// 最早的条目
	now := time.Now().Unix()
	rows, err = db.Query(`
		SELECT title, owner_name, add_at
		FROM watchlater_videos
		WHERE add_at > 0
		ORDER BY add_at ASC
		LIMIT 5
	`)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var item WatchLaterItem
			if err := rows.Scan(&item.Title, &item.OwnerName, &item.AddAt); err == nil {
				item.DaysAgo = int((now - item.AddAt) / 86400)
				result.OldestItems = append(result.OldestItems, item)
			}
		}
	}

	return result, nil
}

// GetExtraStatsOverview 获取额外模块概览统计
func GetExtraStatsOverview() map[string]interface{} {
	overview := map[string]interface{}{}

	// 点赞数
	if db := GetLikesDB(); db != nil {
		var count int
		db.QueryRow("SELECT COUNT(*) FROM liked_videos").Scan(&count)
		overview["likes_count"] = count
	}

	// 稍后再看数
	if db := GetWatchLaterDB(); db != nil {
		var count int
		db.QueryRow("SELECT COUNT(*) FROM watchlater_videos").Scan(&count)
		overview["watchlater_count"] = count
	}

	// 收藏数
	if db := GetFavoritesDB(); db != nil {
		var count int
		db.QueryRow("SELECT COUNT(*) FROM favorites_content").Scan(&count)
		overview["favorites_count"] = count

		var folderCount int
		db.QueryRow("SELECT COUNT(*) FROM favorites_folder").Scan(&folderCount)
		overview["favorites_folder_count"] = folderCount
	}

	return overview
}
