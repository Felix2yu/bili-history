package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"bili-history/internal/config"
	"bili-history/internal/db"
	"bili-history/internal/services/biliapi"

	"github.com/gin-gonic/gin"
)

// GetVideoInfoFromDB returns video info from the database.
func GetVideoInfoFromDB(c *gin.Context) {
	bvid := c.Param("bvid")
	if bvid == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bvid is required"})
		return
	}

	database, err := db.Open(config.GetOutputPath("database", "video_library.db"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	var (
		vidBvid, title, desc, tname string
		duration, tid               int
		pubdate                     int64
		ownerMid                    int64
		ownerName, ownerFace        string
		statView, statDanmaku       int
		statReply, statFavorite     int
		statCoin, statShare         int
		statLike                    int
	)

	err = database.QueryRow(`
		SELECT bvid, title, "desc", duration, pubdate, tid, tname,
			owner_mid, owner_name, owner_face,
			stat_view, stat_danmaku, stat_reply, stat_favorite,
			stat_coin, stat_share, stat_like
		FROM video_base_info WHERE bvid = ?`, bvid).Scan(
		&vidBvid, &title, &desc,
		&duration, &pubdate, &tid, &tname,
		&ownerMid, &ownerName, &ownerFace,
		&statView, &statDanmaku, &statReply,
		&statFavorite, &statCoin, &statShare,
		&statLike,
	)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Video not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": gin.H{
		"bvid": vidBvid, "title": title, "desc": desc,
		"duration": duration, "pubdate": pubdate, "tid": tid, "tname": tname,
		"owner_mid": ownerMid, "owner_name": ownerName, "owner_face": ownerFace,
		"stat_view": statView, "stat_danmaku": statDanmaku, "stat_reply": statReply,
		"stat_favorite": statFavorite, "stat_coin": statCoin, "stat_share": statShare,
		"stat_like": statLike,
	}})
}

// BatchFetchVideoDetails fetches video details for multiple BVIDs.
func BatchFetchVideoDetails(c *gin.Context) {
	var req struct {
		BVIDs []string `json:"bvids"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || len(req.BVIDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bvids array is required"})
		return
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Config error"})
		return
	}

	client := biliapi.NewClient(cfg.SESSDATA, cfg.BiliJCT, cfg.DedeUserID)
	infos, err := client.BatchFetchVideoInfo(req.BVIDs, 10)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": infos, "count": len(infos)})
}

// FetchVideoDetailsFromHistory batch fetches video details for all history BVIDs.
func FetchVideoDetailsFromHistory(c *gin.Context) {
	yearStr := c.DefaultQuery("year", strconv.Itoa(time.Now().Year()))
	year, _ := strconv.Atoi(yearStr)

	database, err := db.OpenHistoryDB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	tableName := db.GetYearTable(year)
	rows, err := database.Query(fmt.Sprintf(`SELECT DISTINCT bvid FROM %s LIMIT 100`, tableName))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Query error"})
		return
	}
	defer rows.Close()

	var bvids []string
	for rows.Next() {
		var bvid string
		rows.Scan(&bvid)
		bvids = append(bvids, bvid)
	}

	if len(bvids) == 0 {
		c.JSON(http.StatusOK, gin.H{"data": []interface{}{}, "count": 0})
		return
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Config error"})
		return
	}

	client := biliapi.NewClient(cfg.SESSDATA, cfg.BiliJCT, cfg.DedeUserID)
	infos, err := client.BatchFetchVideoInfo(bvids, 10)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": infos, "count": len(infos)})
}

// GetVideoStats returns video details statistics.
func GetVideoStats(c *gin.Context) {
	database, err := db.Open(config.GetOutputPath("database", "video_library.db"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	var totalVideos, totalUploaders, totalTags int
	database.QueryRow("SELECT COUNT(*) FROM video_base_info").Scan(&totalVideos)
	database.QueryRow("SELECT COUNT(*) FROM uploader_info").Scan(&totalUploaders)
	database.QueryRow("SELECT COUNT(*) FROM video_tags").Scan(&totalTags)

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"total_videos":    totalVideos,
			"total_uploaders": totalUploaders,
			"total_tags":      totalTags,
		},
	})
}

// GetUploaderList returns the list of uploaders.
func GetUploaderList(c *gin.Context) {
	database, err := db.Open(config.GetOutputPath("database", "video_library.db"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	rows, err := database.Query(`
		SELECT mid, name, face, fans, level FROM uploader_info
		ORDER BY fans DESC LIMIT 50`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Query error"})
		return
	}
	defer rows.Close()

	type uploader struct {
		Mid   int64  `json:"mid"`
		Name  string `json:"name"`
		Face  string `json:"face"`
		Fans  int    `json:"fans"`
		Level int    `json:"level"`
	}
	var uploaders []uploader
	for rows.Next() {
		var u uploader
		rows.Scan(&u.Mid, &u.Name, &u.Face, &u.Fans, &u.Level)
		uploaders = append(uploaders, u)
	}

	c.JSON(http.StatusOK, gin.H{"data": uploaders})
}

// GetVideoTags returns video tag list.
func GetVideoTags(c *gin.Context) {
	database, err := db.Open(config.GetOutputPath("database", "video_library.db"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	rows, err := database.Query(`
		SELECT tag_name, COUNT(*) as cnt FROM video_tags
		GROUP BY tag_name ORDER BY cnt DESC LIMIT 50`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Query error"})
		return
	}
	defer rows.Close()

	type tagCount struct {
		Name  string `json:"name"`
		Count int    `json:"count"`
	}
	var tags []tagCount
	for rows.Next() {
		var t tagCount
		rows.Scan(&t.Name, &t.Count)
		tags = append(tags, t)
	}

	c.JSON(http.StatusOK, gin.H{"data": tags})
}

// SearchVideoDetails searches videos in the database.
func SearchVideoDetails(c *gin.Context) {
	keyword := c.Query("keyword")
	if keyword == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "keyword is required"})
		return
	}

	database, err := db.Open(config.GetOutputPath("database", "video_library.db"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	rows, err := database.Query(`
		SELECT bvid, title, owner_name, duration, stat_view
		FROM video_base_info
		WHERE title LIKE ? OR owner_name LIKE ?
		ORDER BY stat_view DESC LIMIT 50`,
		"%"+keyword+"%", "%"+keyword+"%")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Query error"})
		return
	}
	defer rows.Close()

	type videoResult struct {
		BVID      string `json:"bvid"`
		Title     string `json:"title"`
		OwnerName string `json:"owner_name"`
		Duration  int    `json:"duration"`
		Views     int    `json:"views"`
	}
	var results []videoResult
	for rows.Next() {
		var v videoResult
		rows.Scan(&v.BVID, &v.Title, &v.OwnerName, &v.Duration, &v.Views)
		results = append(results, v)
	}

	c.JSON(http.StatusOK, gin.H{"data": results})
}

// GetUploaderInfo returns detailed uploader information.
func GetUploaderInfo(c *gin.Context) {
	midStr := c.Param("mid")
	if midStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "mid is required"})
		return
	}
	mid, _ := strconv.ParseInt(midStr, 10, 64)

	database, err := db.Open(config.GetOutputPath("database", "video_library.db"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	var (
		dbMid           int64
		dbName, dbFace  string
		dbSign          string
		dbLevel         int
		dbFans, dbFriend, dbAttention int
		dbOfficialRole  int
		dbOfficialTitle string
		dbVipType, dbVipStatus int
		dbArchiveCount, dbLikeNum int
	)

	err = database.QueryRow(`
		SELECT mid, name, face, COALESCE(sign,''), level, fans, friend, attention,
			official_role, COALESCE(official_title,''), vip_type, vip_status,
			archive_count, like_num
		FROM uploader_info WHERE mid = ?`, mid).Scan(
		&dbMid, &dbName, &dbFace, &dbSign, &dbLevel,
		&dbFans, &dbFriend, &dbAttention,
		&dbOfficialRole, &dbOfficialTitle,
		&dbVipType, &dbVipStatus,
		&dbArchiveCount, &dbLikeNum,
	)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Uploader not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": gin.H{
		"mid": dbMid, "name": dbName, "face": dbFace, "sign": dbSign,
		"level": dbLevel, "fans": dbFans, "friend": dbFriend, "attention": dbAttention,
		"official_role": dbOfficialRole, "official_title": dbOfficialTitle,
		"vip_type": dbVipType, "vip_status": dbVipStatus,
		"archive_count": dbArchiveCount, "like_num": dbLikeNum,
	}})
}

// GetMainCategories returns main video categories.
func GetMainCategories(c *gin.Context) {
	database, err := db.OpenHistoryDB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	rows, err := database.Query(`
		SELECT DISTINCT main_category FROM video_categories
		ORDER BY main_category`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Query error"})
		return
	}
	defer rows.Close()

	var categories []string
	for rows.Next() {
		var cat string
		rows.Scan(&cat)
		categories = append(categories, cat)
	}

	c.JSON(http.StatusOK, gin.H{"data": categories})
}

// GetSubCategories returns sub-categories for a main category.
func GetSubCategories(c *gin.Context) {
	mainCategory := c.Param("main_category")
	if mainCategory == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "main_category is required"})
		return
	}

	database, err := db.OpenHistoryDB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	rows, err := database.Query(`
		SELECT sub_category, alias, tid, COALESCE(image,'') FROM video_categories
		WHERE main_category = ?`, mainCategory)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Query error"})
		return
	}
	defer rows.Close()

	type subCat struct {
		SubCategory string `json:"sub_category"`
		Alias       string `json:"alias"`
		Tid         int    `json:"tid"`
		Image       string `json:"image"`
	}
	var cats []subCat
	for rows.Next() {
		var s subCat
		rows.Scan(&s.SubCategory, &s.Alias, &s.Tid, &s.Image)
		cats = append(cats, s)
	}

	c.JSON(http.StatusOK, gin.H{"data": cats})
}

// GetInvalidVideos returns list of invalid videos.
func GetInvalidVideos(c *gin.Context) {
	database, err := db.Open(config.GetOutputPath("database", "video_library.db"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	rows, err := database.Query(`
		SELECT bvid, COALESCE(error_type,''), COALESCE(error_code,0),
			COALESCE(error_message,''), COALESCE(first_check_time,0),
			COALESCE(last_check_time,0), COALESCE(check_count,0)
		FROM invalid_videos ORDER BY last_check_time DESC LIMIT 100`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Query error"})
		return
	}
	defer rows.Close()

	type invalidVideo struct {
		BVID         string `json:"bvid"`
		ErrorType    string `json:"error_type"`
		ErrorCode    int    `json:"error_code"`
		ErrorMessage string `json:"error_message"`
		FirstCheck   int64  `json:"first_check_time"`
		LastCheck    int64  `json:"last_check_time"`
		CheckCount   int    `json:"check_count"`
	}
	var videos []invalidVideo
	for rows.Next() {
		var v invalidVideo
		rows.Scan(&v.BVID, &v.ErrorType, &v.ErrorCode, &v.ErrorMessage,
			&v.FirstCheck, &v.LastCheck, &v.CheckCount)
		videos = append(videos, v)
	}

	c.JSON(http.StatusOK, gin.H{"data": videos})
}

// BatchDeleteHistory deletes multiple history records.
func BatchDeleteHistory(c *gin.Context) {
	var req struct {
		Items []struct {
			BVID   string `json:"bvid"`
			ViewAt int64  `json:"view_at"`
		} `json:"items"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	database, err := db.OpenHistoryDB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	totalDeleted := 0
	for _, item := range req.Items {
		year := time.Unix(item.ViewAt, 0).Year()
		tableName := db.GetYearTable(year)
		result, err := database.Exec(fmt.Sprintf(
			"DELETE FROM %s WHERE bvid = ? AND view_at = ?", tableName),
			item.BVID, item.ViewAt)
		if err != nil {
			continue
		}
		affected, _ := result.RowsAffected()
		totalDeleted += int(affected)
	}

	c.JSON(http.StatusOK, gin.H{
		"deleted_count": totalDeleted,
		"message":       fmt.Sprintf("Deleted %d records", totalDeleted),
	})
}
