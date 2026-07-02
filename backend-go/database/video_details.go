package database

import (
	"database/sql"
	"fmt"
	"time"

	"bilibili-history-go/models"
	"bilibili-history-go/utils"
)

const (
	videoBaseInfoTable = "video_base_info"
	videoTagsTable     = "video_tags"
	uploaderInfoTable  = "uploader_info"
)

// EnsureVideoDetailsTables 确保视频详情相关表存在
func EnsureVideoDetailsTables() error {
	db := GetSQLiteDB()
	conn := db.GetDB()
	if conn == nil {
		return fmt.Errorf("database not initialized")
	}

	db.mu.Lock()
	defer db.mu.Unlock()

	createVideoBaseInfoTableSQL := `
	CREATE TABLE IF NOT EXISTS video_base_info (
		id INTEGER PRIMARY KEY,
		bvid TEXT NOT NULL UNIQUE,
		aid INTEGER NOT NULL,
		videos INTEGER DEFAULT 1,
		tid INTEGER,
		tname TEXT,
		copyright INTEGER,
		pic TEXT,
		title TEXT NOT NULL,
		pubdate INTEGER,
		ctime INTEGER,
		desc TEXT,
		duration INTEGER,
		cid INTEGER,
		owner_mid INTEGER,
		owner_name TEXT,
		owner_face TEXT,
		stat_view INTEGER DEFAULT 0,
		stat_danmaku INTEGER DEFAULT 0,
		stat_reply INTEGER DEFAULT 0,
		stat_favorite INTEGER DEFAULT 0,
		stat_coin INTEGER DEFAULT 0,
		stat_share INTEGER DEFAULT 0,
		stat_like INTEGER DEFAULT 0,
		fetch_time INTEGER NOT NULL,
		update_time INTEGER DEFAULT 0
	);`

	createVideoTagsTableSQL := `
	CREATE TABLE IF NOT EXISTS video_tags (
		id INTEGER PRIMARY KEY,
		bvid TEXT NOT NULL,
		tag_id INTEGER NOT NULL,
		tag_name TEXT NOT NULL,
		UNIQUE(bvid, tag_id)
	);`

	createUploaderInfoTableSQL := `
	CREATE TABLE IF NOT EXISTS uploader_info (
		mid INTEGER PRIMARY KEY,
		name TEXT NOT NULL,
		sex TEXT,
		face TEXT,
		sign TEXT,
		level INTEGER DEFAULT 0,
		fans INTEGER DEFAULT 0,
		attention INTEGER DEFAULT 0,
		archive_count INTEGER DEFAULT 0,
		fetch_time INTEGER NOT NULL,
		update_time INTEGER DEFAULT 0
	);`

	indexSQLs := []string{
		"CREATE INDEX IF NOT EXISTS idx_video_base_info_owner_mid ON video_base_info (owner_mid);",
		"CREATE INDEX IF NOT EXISTS idx_video_base_info_title ON video_base_info (title);",
		"CREATE INDEX IF NOT EXISTS idx_video_base_info_fetch_time ON video_base_info (fetch_time);",
		"CREATE INDEX IF NOT EXISTS idx_video_tags_bvid ON video_tags (bvid);",
		"CREATE INDEX IF NOT EXISTS idx_video_tags_tag_name ON video_tags (tag_name);",
	}

	statements := []string{
		createVideoBaseInfoTableSQL,
		createVideoTagsTableSQL,
		createUploaderInfoTableSQL,
	}
	statements = append(statements, indexSQLs...)

	for _, stmt := range statements {
		_, err := conn.Exec(stmt)
		if err != nil {
			return fmt.Errorf("failed to create table/index: %v", err)
		}
	}

	return nil
}

// UpsertVideoBaseInfo 插入或更新视频基本信息
func UpsertVideoBaseInfo(video *models.VideoBaseInfo) error {
	db := GetSQLiteDB()
	conn := db.GetDB()
	if conn == nil {
		return fmt.Errorf("database not initialized")
	}

	if err := EnsureVideoDetailsTables(); err != nil {
		return err
	}

	db.mu.Lock()
	defer db.mu.Unlock()

	var existingID int64
	err := conn.QueryRow("SELECT id FROM video_base_info WHERE bvid = ?", video.Bvid).Scan(&existingID)
	if err == nil {
		_, err = conn.Exec(`
			UPDATE video_base_info SET
				aid = ?, videos = ?, tid = ?, tname = ?, copyright = ?, pic = ?,
				title = ?, pubdate = ?, ctime = ?, desc = ?, duration = ?, cid = ?,
				owner_mid = ?, owner_name = ?, owner_face = ?,
				stat_view = ?, stat_danmaku = ?, stat_reply = ?, stat_favorite = ?,
				stat_coin = ?, stat_share = ?, stat_like = ?,
				update_time = ?
			WHERE bvid = ?
		`,
			video.Aid, video.Videos, video.Tid, video.Tname, video.Copyright, video.Pic,
			video.Title, video.Pubdate, video.Ctime, video.Desc, video.Duration, video.Cid,
			video.OwnerMid, video.OwnerName, video.OwnerFace,
			video.StatView, video.StatDanmaku, video.StatReply, video.StatFavorite,
			video.StatCoin, video.StatShare, video.StatLike,
			video.UpdateTime, video.Bvid,
		)
		return err
	}

	if err != sql.ErrNoRows {
		return err
	}

	_, err = conn.Exec(`
		INSERT INTO video_base_info (
			bvid, aid, videos, tid, tname, copyright, pic, title, pubdate, ctime,
			desc, duration, cid, owner_mid, owner_name, owner_face,
			stat_view, stat_danmaku, stat_reply, stat_favorite, stat_coin, stat_share, stat_like,
			fetch_time, update_time
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		video.Bvid, video.Aid, video.Videos, video.Tid, video.Tname, video.Copyright, video.Pic,
		video.Title, video.Pubdate, video.Ctime, video.Desc, video.Duration, video.Cid,
		video.OwnerMid, video.OwnerName, video.OwnerFace,
		video.StatView, video.StatDanmaku, video.StatReply, video.StatFavorite,
		video.StatCoin, video.StatShare, video.StatLike,
		video.FetchTime, video.UpdateTime,
	)

	return err
}

// GetVideoBaseInfoByBvid 根据bvid获取视频基本信息
func GetVideoBaseInfoByBvid(bvid string) (*models.VideoBaseInfo, error) {
	db := GetSQLiteDB()
	conn := db.GetDB()
	if conn == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	exists, _ := db.TableExists(videoBaseInfoTable)
	if !exists {
		return nil, fmt.Errorf("视频详情表不存在")
	}

	db.mu.RLock()
	defer db.mu.RUnlock()

	video := &models.VideoBaseInfo{}
	err := conn.QueryRow(`
		SELECT id, bvid, aid, videos, tid, tname, copyright, pic, title, pubdate,
			ctime, desc, duration, cid, owner_mid, owner_name, owner_face,
			stat_view, stat_danmaku, stat_reply, stat_favorite, stat_coin, stat_share, stat_like,
			fetch_time, update_time
		FROM video_base_info WHERE bvid = ?
	`, bvid).Scan(
		&video.ID, &video.Bvid, &video.Aid, &video.Videos, &video.Tid, &video.Tname,
		&video.Copyright, &video.Pic, &video.Title, &video.Pubdate, &video.Ctime,
		&video.Desc, &video.Duration, &video.Cid, &video.OwnerMid, &video.OwnerName,
		&video.OwnerFace, &video.StatView, &video.StatDanmaku, &video.StatReply,
		&video.StatFavorite, &video.StatCoin, &video.StatShare, &video.StatLike,
		&video.FetchTime, &video.UpdateTime,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return video, nil
}

// SearchVideos 搜索视频（按关键词搜索标题/UP主）
func SearchVideos(keyword string, page, size int) (*models.PagedResponse, error) {
	db := GetSQLiteDB()
	conn := db.GetDB()
	if conn == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	exists, _ := db.TableExists(videoBaseInfoTable)
	if !exists {
		return &models.PagedResponse{
			Records: []interface{}{},
			Total:   0,
			Size:    size,
			Current: page,
		}, nil
	}

	db.mu.RLock()
	defer db.mu.RUnlock()

	searchKeyword := "%" + keyword + "%"

	countQuery := `
		SELECT COUNT(*) FROM video_base_info
		WHERE title LIKE ? OR owner_name LIKE ?
	`
	var total int64
	err := conn.QueryRow(countQuery, searchKeyword, searchKeyword).Scan(&total)
	if err != nil {
		return nil, err
	}

	query := `
		SELECT id, bvid, aid, videos, tid, tname, copyright, pic, title, pubdate,
			ctime, desc, duration, cid, owner_mid, owner_name, owner_face,
			stat_view, stat_danmaku, stat_reply, stat_favorite, stat_coin, stat_share, stat_like,
			fetch_time, update_time
		FROM video_base_info
		WHERE title LIKE ? OR owner_name LIKE ?
		ORDER BY pubdate DESC
		LIMIT ? OFFSET ?
	`

	rows, err := conn.Query(query, searchKeyword, searchKeyword, size, (page-1)*size)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var videos []map[string]interface{}
	for rows.Next() {
		video := &models.VideoBaseInfo{}
		err := rows.Scan(
			&video.ID, &video.Bvid, &video.Aid, &video.Videos, &video.Tid, &video.Tname,
			&video.Copyright, &video.Pic, &video.Title, &video.Pubdate, &video.Ctime,
			&video.Desc, &video.Duration, &video.Cid, &video.OwnerMid, &video.OwnerName,
			&video.OwnerFace, &video.StatView, &video.StatDanmaku, &video.StatReply,
			&video.StatFavorite, &video.StatCoin, &video.StatShare, &video.StatLike,
			&video.FetchTime, &video.UpdateTime,
		)
		if err != nil {
			continue
		}
		videos = append(videos, videoBaseInfoToMap(video))
	}

	return &models.PagedResponse{
		Records: videos,
		Total:   total,
		Size:    size,
		Current: page,
	}, nil
}

// GetVideoDetailStats 获取视频详情统计信息
func GetVideoDetailStats() (*models.VideoDetailStats, error) {
	db := GetSQLiteDB()
	conn := db.GetDB()
	if conn == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	stats := &models.VideoDetailStats{}

	videoTableExists, _ := db.TableExists(videoBaseInfoTable)
	if videoTableExists {
		conn.QueryRow("SELECT COUNT(*) FROM video_base_info").Scan(&stats.FetchedVideos)
	}

	uploaderTableExists, _ := db.TableExists(uploaderInfoTable)
	if uploaderTableExists {
		conn.QueryRow("SELECT COUNT(*) FROM uploader_info").Scan(&stats.TotalUploaders)
	}

	tagsTableExists, _ := db.TableExists(videoTagsTable)
	if tagsTableExists {
		conn.QueryRow("SELECT COUNT(DISTINCT tag_name) FROM video_tags").Scan(&stats.TotalTags)
	}

	return stats, nil
}

// GetDatabaseStats 获取数据库详细统计
func GetDatabaseStats() (map[string]interface{}, error) {
	db := GetSQLiteDB()
	conn := db.GetDB()
	if conn == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	result := make(map[string]interface{})

	videoTableExists, _ := db.TableExists(videoBaseInfoTable)
	if videoTableExists {
		var totalVideos int64
		conn.QueryRow("SELECT COUNT(*) FROM video_base_info").Scan(&totalVideos)
		result["total_videos"] = totalVideos

		var totalViews int64
		conn.QueryRow("SELECT COALESCE(SUM(stat_view), 0) FROM video_base_info").Scan(&totalViews)
		result["total_views"] = totalViews

		var totalDanmaku int64
		conn.QueryRow("SELECT COALESCE(SUM(stat_danmaku), 0) FROM video_base_info").Scan(&totalDanmaku)
		result["total_danmaku"] = totalDanmaku

		var totalLikes int64
		conn.QueryRow("SELECT COALESCE(SUM(stat_like), 0) FROM video_base_info").Scan(&totalLikes)
		result["total_likes"] = totalLikes

		var totalCoins int64
		conn.QueryRow("SELECT COALESCE(SUM(stat_coin), 0) FROM video_base_info").Scan(&totalCoins)
		result["total_coins"] = totalCoins

		var totalFavorites int64
		conn.QueryRow("SELECT COALESCE(SUM(stat_favorite), 0) FROM video_base_info").Scan(&totalFavorites)
		result["total_favorites"] = totalFavorites

		var totalShares int64
		conn.QueryRow("SELECT COALESCE(SUM(stat_share), 0) FROM video_base_info").Scan(&totalShares)
		result["total_shares"] = totalShares

		var totalDuration int64
		conn.QueryRow("SELECT COALESCE(SUM(duration), 0) FROM video_base_info").Scan(&totalDuration)
		result["total_duration_seconds"] = totalDuration
	}

	uploaderTableExists, _ := db.TableExists(uploaderInfoTable)
	if uploaderTableExists {
		var totalUploaders int64
		conn.QueryRow("SELECT COUNT(*) FROM uploader_info").Scan(&totalUploaders)
		result["total_uploaders"] = totalUploaders
	}

	tagsTableExists, _ := db.TableExists(videoTagsTable)
	if tagsTableExists {
		var totalTags int64
		conn.QueryRow("SELECT COUNT(DISTINCT tag_name) FROM video_tags").Scan(&totalTags)
		result["total_unique_tags"] = totalTags
	}

	return result, nil
}

// GetUploaderList 获取UP主列表（分页、按观看数/投稿数排序）
func GetUploaderList(page, size int, sortBy string) (*models.PagedResponse, error) {
	db := GetSQLiteDB()
	conn := db.GetDB()
	if conn == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	exists, _ := db.TableExists(videoBaseInfoTable)
	if !exists {
		return &models.PagedResponse{
			Records: []interface{}{},
			Total:   0,
			Size:    size,
			Current: page,
		}, nil
	}

	db.mu.RLock()
	defer db.mu.RUnlock()

	var sortColumn string
	switch sortBy {
	case "video_count":
		sortColumn = "video_count"
	case "views":
		sortColumn = "total_views"
	case "likes":
		sortColumn = "total_likes"
	default:
		sortColumn = "video_count"
	}

	countQuery := `
		SELECT COUNT(DISTINCT owner_mid) FROM video_base_info WHERE owner_mid > 0
	`
	var total int64
	err := conn.QueryRow(countQuery).Scan(&total)
	if err != nil {
		return nil, err
	}

	query := fmt.Sprintf(`
		SELECT 
			owner_mid, 
			owner_name, 
			owner_face,
			COUNT(*) as video_count,
			COALESCE(SUM(stat_view), 0) as total_views,
			COALESCE(SUM(stat_like), 0) as total_likes,
			COALESCE(SUM(stat_coin), 0) as total_coins,
			COALESCE(SUM(stat_favorite), 0) as total_favorites
		FROM video_base_info
		WHERE owner_mid > 0
		GROUP BY owner_mid, owner_name, owner_face
		ORDER BY %s DESC
		LIMIT ? OFFSET ?
	`, sortColumn)

	rows, err := conn.Query(query, size, (page-1)*size)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var uploaders []map[string]interface{}
	for rows.Next() {
		stat := &models.UploaderStats{}
		err := rows.Scan(
			&stat.Mid, &stat.Name, &stat.Face,
			&stat.VideoCount, &stat.TotalViews, &stat.TotalLikes,
			&stat.TotalCoins, &stat.TotalFavorites,
		)
		if err != nil {
			continue
		}
		uploaders = append(uploaders, uploaderStatsToMap(stat))
	}

	return &models.PagedResponse{
		Records: uploaders,
		Total:   total,
		Size:    size,
		Current: page,
	}, nil
}

// GetUploaderDetail 获取UP主详细信息
func GetUploaderDetail(mid int) (map[string]interface{}, error) {
	db := GetSQLiteDB()
	conn := db.GetDB()
	if conn == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	videoTableExists, _ := db.TableExists(videoBaseInfoTable)
	if !videoTableExists {
		return nil, fmt.Errorf("视频详情表不存在")
	}

	db.mu.RLock()
	defer db.mu.RUnlock()

	result := make(map[string]interface{})
	result["mid"] = mid

	var name, face string
	err := conn.QueryRow(`
		SELECT owner_name, owner_face FROM video_base_info 
		WHERE owner_mid = ? LIMIT 1
	`, mid).Scan(&name, &face)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	result["name"] = name
	result["face"] = face

	var videoCount int64
	var totalViews, totalLikes, totalCoins, totalFavorites, totalShares, totalDanmaku int64
	err = conn.QueryRow(`
		SELECT 
			COUNT(*),
			COALESCE(SUM(stat_view), 0),
			COALESCE(SUM(stat_like), 0),
			COALESCE(SUM(stat_coin), 0),
			COALESCE(SUM(stat_favorite), 0),
			COALESCE(SUM(stat_share), 0),
			COALESCE(SUM(stat_danmaku), 0)
		FROM video_base_info WHERE owner_mid = ?
	`, mid).Scan(
		&videoCount, &totalViews, &totalLikes, &totalCoins,
		&totalFavorites, &totalShares, &totalDanmaku,
	)
	if err != nil {
		return nil, err
	}

	result["video_count"] = videoCount
	result["total_views"] = totalViews
	result["total_likes"] = totalLikes
	result["total_coins"] = totalCoins
	result["total_favorites"] = totalFavorites
	result["total_shares"] = totalShares
	result["total_danmaku"] = totalDanmaku

	uploaderTableExists, _ := db.TableExists(uploaderInfoTable)
	if uploaderTableExists {
		var uploader models.UploaderInfo
		err = conn.QueryRow(`
			SELECT mid, name, sex, face, sign, level, fans, attention, archive_count, fetch_time, update_time
			FROM uploader_info WHERE mid = ?
		`, mid).Scan(
			&uploader.Mid, &uploader.Name, &uploader.Sex, &uploader.Face, &uploader.Sign,
			&uploader.Level, &uploader.Fans, &uploader.Attention, &uploader.ArchiveCount,
			&uploader.FetchTime, &uploader.UpdateTime,
		)
		if err == nil {
			result["sign"] = uploader.Sign
			result["level"] = uploader.Level
			result["fans"] = uploader.Fans
			result["attention"] = uploader.Attention
			result["archive_count"] = uploader.ArchiveCount
		}
	}

	return result, nil
}

// GetTagList 获取标签列表
func GetTagList(page, size int) (*models.PagedResponse, error) {
	db := GetSQLiteDB()
	conn := db.GetDB()
	if conn == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	exists, _ := db.TableExists(videoTagsTable)
	if !exists {
		return &models.PagedResponse{
			Records: []interface{}{},
			Total:   0,
			Size:    size,
			Current: page,
		}, nil
	}

	db.mu.RLock()
	defer db.mu.RUnlock()

	countQuery := `
		SELECT COUNT(DISTINCT tag_name) FROM video_tags
	`
	var total int64
	err := conn.QueryRow(countQuery).Scan(&total)
	if err != nil {
		return nil, err
	}

	query := `
		SELECT tag_name, COUNT(*) as video_count
		FROM video_tags
		GROUP BY tag_name
		ORDER BY video_count DESC
		LIMIT ? OFFSET ?
	`

	rows, err := conn.Query(query, size, (page-1)*size)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags []map[string]interface{}
	for rows.Next() {
		var tagName string
		var videoCount int64
		err := rows.Scan(&tagName, &videoCount)
		if err != nil {
			continue
		}
		tags = append(tags, map[string]interface{}{
			"tag_name":    tagName,
			"video_count": videoCount,
		})
	}

	return &models.PagedResponse{
		Records: tags,
		Total:   total,
		Size:    size,
		Current: page,
	}, nil
}

// GetUniqueBvidsFromHistory 从历史记录中获取所有不重复的bvid
func GetUniqueBvidsFromHistory() ([]string, error) {
	db := GetSQLiteDB()
	conn := db.GetDB()
	if conn == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	availableYears, err := db.GetAvailableYears()
	if err != nil {
		return nil, err
	}

	db.mu.RLock()
	defer db.mu.RUnlock()

	var bvidSet = make(map[string]bool)

	for _, year := range availableYears {
		tableName := fmt.Sprintf("bilibili_history_%d", year)
		exists, _ := db.TableExists(tableName)
		if !exists {
			continue
		}

		query := fmt.Sprintf("SELECT DISTINCT bvid FROM %s WHERE bvid != '' AND business = 'archive'", tableName)
		rows, err := conn.Query(query)
		if err != nil {
			continue
		}

		for rows.Next() {
			var bvid string
			if err := rows.Scan(&bvid); err == nil && bvid != "" {
				bvidSet[bvid] = true
			}
		}
		rows.Close()
	}

	bvids := make([]string, 0, len(bvidSet))
	for bvid := range bvidSet {
		bvids = append(bvids, bvid)
	}

	return bvids, nil
}

// GetFetchedBvids 获取已获取详情的bvid列表
func GetFetchedBvids() (map[string]bool, error) {
	db := GetSQLiteDB()
	conn := db.GetDB()
	if conn == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	exists, _ := db.TableExists(videoBaseInfoTable)
	if !exists {
		return make(map[string]bool), nil
	}

	db.mu.RLock()
	defer db.mu.RUnlock()

	bvidSet := make(map[string]bool)
	rows, err := conn.Query("SELECT bvid FROM video_base_info")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var bvid string
		if err := rows.Scan(&bvid); err == nil {
			bvidSet[bvid] = true
		}
	}

	return bvidSet, nil
}

// UpsertVideoTags 批量插入视频标签
func UpsertVideoTags(bvid string, tags []models.VideoTag) error {
	db := GetSQLiteDB()
	conn := db.GetDB()
	if conn == nil {
		return fmt.Errorf("database not initialized")
	}

	if err := EnsureVideoDetailsTables(); err != nil {
		return err
	}

	db.mu.Lock()
	defer db.mu.Unlock()

	_, err := conn.Exec("DELETE FROM video_tags WHERE bvid = ?", bvid)
	if err != nil {
		return err
	}

	for _, tag := range tags {
		_, err = conn.Exec(`
			INSERT OR IGNORE INTO video_tags (bvid, tag_id, tag_name)
			VALUES (?, ?, ?)
		`, bvid, tag.TagID, tag.TagName)
		if err != nil {
			return err
		}
	}

	return nil
}

// UpsertUploaderInfo 插入或更新UP主信息
func UpsertUploaderInfo(uploader *models.UploaderInfo) error {
	db := GetSQLiteDB()
	conn := db.GetDB()
	if conn == nil {
		return fmt.Errorf("database not initialized")
	}

	if err := EnsureVideoDetailsTables(); err != nil {
		return err
	}

	db.mu.Lock()
	defer db.mu.Unlock()

	var existingMid int
	err := conn.QueryRow("SELECT mid FROM uploader_info WHERE mid = ?", uploader.Mid).Scan(&existingMid)
	if err == nil {
		_, err = conn.Exec(`
			UPDATE uploader_info SET
				name = ?, sex = ?, face = ?, sign = ?, level = ?,
				fans = ?, attention = ?, archive_count = ?, update_time = ?
			WHERE mid = ?
		`,
			uploader.Name, uploader.Sex, uploader.Face, uploader.Sign, uploader.Level,
			uploader.Fans, uploader.Attention, uploader.ArchiveCount, uploader.UpdateTime,
			uploader.Mid,
		)
		return err
	}

	if err != sql.ErrNoRows {
		return err
	}

	_, err = conn.Exec(`
		INSERT INTO uploader_info (
			mid, name, sex, face, sign, level, fans, attention, archive_count,
			fetch_time, update_time
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		uploader.Mid, uploader.Name, uploader.Sex, uploader.Face, uploader.Sign,
		uploader.Level, uploader.Fans, uploader.Attention, uploader.ArchiveCount,
		uploader.FetchTime, uploader.UpdateTime,
	)

	return err
}

func videoBaseInfoToMap(video *models.VideoBaseInfo) map[string]interface{} {
	return map[string]interface{}{
		"id":              video.ID,
		"bvid":            video.Bvid,
		"aid":             video.Aid,
		"videos":          video.Videos,
		"tid":             video.Tid,
		"tname":           video.Tname,
		"copyright":       video.Copyright,
		"pic":             video.Pic,
		"title":           video.Title,
		"pubdate":         video.Pubdate,
		"pubdate_str":     utils.FormatDateTime(time.Unix(video.Pubdate, 0)),
		"ctime":           video.Ctime,
		"desc":            video.Desc,
		"duration":        video.Duration,
		"duration_str":    formatDuration(video.Duration),
		"cid":             video.Cid,
		"owner_mid":       video.OwnerMid,
		"owner_name":      video.OwnerName,
		"owner_face":      video.OwnerFace,
		"stat_view":       video.StatView,
		"stat_danmaku":    video.StatDanmaku,
		"stat_reply":      video.StatReply,
		"stat_favorite":   video.StatFavorite,
		"stat_coin":       video.StatCoin,
		"stat_share":      video.StatShare,
		"stat_like":       video.StatLike,
		"fetch_time":      video.FetchTime,
		"update_time":     video.UpdateTime,
		"original_url":    fmt.Sprintf("https://www.bilibili.com/video/%s", video.Bvid),
	}
}

func uploaderStatsToMap(stat *models.UploaderStats) map[string]interface{} {
	return map[string]interface{}{
		"mid":             stat.Mid,
		"name":            stat.Name,
		"face":            stat.Face,
		"video_count":     stat.VideoCount,
		"total_views":     stat.TotalViews,
		"total_likes":     stat.TotalLikes,
		"total_coins":     stat.TotalCoins,
		"total_favorites": stat.TotalFavorites,
	}
}

func formatDuration(seconds int) string {
	if seconds <= 0 {
		return "00:00"
	}
	hours := seconds / 3600
	minutes := (seconds % 3600) / 60
	secs := seconds % 60
	if hours > 0 {
		return fmt.Sprintf("%02d:%02d:%02d", hours, minutes, secs)
	}
	return fmt.Sprintf("%02d:%02d", minutes, secs)
}
