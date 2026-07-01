package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"bili-history/internal/config"
	"bili-history/internal/db"

	"github.com/gin-gonic/gin"
)

// ListDynamicHosts lists UP主 with dynamics in the database.
func ListDynamicHosts(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "50")
	offsetStr := c.DefaultQuery("offset", "0")
	limit, _ := strconv.Atoi(limitStr)
	offset, _ := strconv.Atoi(offsetStr)

	database, err := db.Open(config.GetOutputPath("database", "bilibili_dynamic.db"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	rows, err := database.Query(`
		SELECT host_mid, COUNT(*) as dynamic_count,
			MAX(publish_ts) as latest_ts
		FROM dynamic_core
		GROUP BY host_mid
		ORDER BY latest_ts DESC
		LIMIT ? OFFSET ?`, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Query error"})
		return
	}
	defer rows.Close()

	type hostInfo struct {
		HostMid       string `json:"host_mid"`
		DynamicCount  int    `json:"dynamic_count"`
		LatestTs      int64  `json:"latest_ts"`
		AuthorName    string `json:"author_name"`
	}
	var hosts []hostInfo
	for rows.Next() {
		var h hostInfo
		rows.Scan(&h.HostMid, &h.DynamicCount, &h.LatestTs)

		// Get author name
		database.QueryRow(`
			SELECT author_name FROM dynamic_core
			WHERE host_mid = ? AND author_name IS NOT NULL AND author_name != ''
			ORDER BY publish_ts DESC LIMIT 1`, h.HostMid).Scan(&h.AuthorName)

		hosts = append(hosts, h)
	}

	c.JSON(http.StatusOK, gin.H{
		"data":   hosts,
		"limit":  limit,
		"offset": offset,
	})
}

// ListDynamicSpace lists dynamics for a specific UP主 from the database.
func ListDynamicSpace(c *gin.Context) {
	hostMid := c.Param("host_mid")
	if hostMid == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "host_mid is required"})
		return
	}

	limitStr := c.DefaultQuery("limit", "20")
	offsetStr := c.DefaultQuery("offset", "0")
	limit, _ := strconv.Atoi(limitStr)
	offset, _ := strconv.Atoi(offsetStr)

	database, err := db.Open(config.GetOutputPath("database", "bilibili_dynamic.db"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	rows, err := database.Query(`
		SELECT c.id_str, c.type, c.publish_ts, c.txt, c.bvid, c.title,
			c.author_name, c.cover,
			COALESCE(s.like_count, 0), COALESCE(s.comment_count, 0),
			COALESCE(s.repost_count, 0), COALESCE(s.view_count, 0)
		FROM dynamic_core c
		LEFT JOIN dynamic_stat s ON c.host_mid = s.host_mid AND c.id_str = s.id_str
		WHERE c.host_mid = ?
		ORDER BY c.publish_ts DESC
		LIMIT ? OFFSET ?`, hostMid, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Query error"})
		return
	}
	defer rows.Close()

	type dynamicItem struct {
		IDStr        string `json:"id_str"`
		Type         string `json:"type"`
		PublishTs    int64  `json:"publish_ts"`
		Txt          string `json:"txt"`
		BVID         string `json:"bvid"`
		Title        string `json:"title"`
		AuthorName   string `json:"author_name"`
		Cover        string `json:"cover"`
		LikeCount    int    `json:"like_count"`
		CommentCount int    `json:"comment_count"`
		RepostCount  int    `json:"repost_count"`
		ViewCount    int    `json:"view_count"`
	}
	var dynamics []dynamicItem
	for rows.Next() {
		var d dynamicItem
		rows.Scan(&d.IDStr, &d.Type, &d.PublishTs, &d.Txt, &d.BVID, &d.Title,
			&d.AuthorName, &d.Cover, &d.LikeCount, &d.CommentCount,
			&d.RepostCount, &d.ViewCount)
		dynamics = append(dynamics, d)
	}

	c.JSON(http.StatusOK, gin.H{
		"data":   dynamics,
		"total":  len(dynamics),
		"limit":  limit,
		"offset": offset,
	})
}

// GetDynamicTypes returns dynamic type descriptions.
func GetDynamicTypes(c *gin.Context) {
	types := map[string]string{
		"DYNAMIC_TYPE_AUTH":    "动态",
		"DYNAMIC_TYPE_AV":      "投稿视频",
		"DYNAMIC_TYPE_PGC":     "番剧/影视",
		"DYNAMIC_TYPE_COURSES": "课程",
		"DYNAMIC_TYPE_WORD":    "文字动态",
		"DYNAMIC_TYPE_DRAW":    "图文动态",
		"DYNAMIC_TYPE_ARTICLE": "专栏文章",
		"DYNAMIC_TYPE_MUSIC":   "音乐",
		"DYNAMIC_TYPE_COMMON_SQUARE":  "通用卡片",
		"DYNAMIC_TYPE_COMMON_VERTICAL": "竖版视频",
		"DYNAMIC_TYPE_LIVE_RCMD":  "直播推荐",
		"DYNAMIC_TYPE_LIVE":     "直播",
		"DYNAMIC_TYPE_MEDIALIST": "收藏夹",
		"DYNAMIC_TYPE_COURSES_SEASON": "课程学期",
		"DYNAMIC_TYPE_TAG":      "话题",
		"DYNAMIC_TYPE_FORWARD":  "转发",
	}
	c.JSON(http.StatusOK, gin.H{"data": types})
}

// GetDynamicDetail returns dynamic detail (placeholder - needs API call).
func GetDynamicDetail(c *gin.Context) {
	dynamicID := c.Param("dynamic_id")
	if dynamicID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "dynamic_id is required"})
		return
	}

	// Query from database
	database, err := db.Open(config.GetOutputPath("database", "bilibili_dynamic.db"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	var txt, bvid, title, cover, authorName string
	var publishTs int64
	err = database.QueryRow(`
		SELECT txt, COALESCE(bvid,''), COALESCE(title,''), COALESCE(cover,''),
			COALESCE(author_name,''), COALESCE(publish_ts,0)
		FROM dynamic_core WHERE id_str = ?`, dynamicID).Scan(
		&txt, &bvid, &title, &cover, &authorName, &publishTs)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Dynamic not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": gin.H{
		"id_str":      dynamicID,
		"txt":         txt,
		"bvid":        bvid,
		"title":       title,
		"cover":       cover,
		"author_name": authorName,
		"publish_ts":  publishTs,
	}})
}

// DeleteDynamicSpace deletes dynamics for a specific UP主.
func DeleteDynamicSpace(c *gin.Context) {
	hostMid := c.Param("host_mid")
	if hostMid == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "host_mid is required"})
		return
	}

	database, err := db.Open(config.GetOutputPath("database", "bilibili_dynamic.db"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	tables := []string{"dynamic_core", "dynamic_author", "dynamic_stat",
		"dynamic_topic", "major_opus_pics", "major_archive_jump_urls"}
	totalDeleted := 0
	for _, table := range tables {
		result, _ := database.Exec(fmt.Sprintf(
			"DELETE FROM %s WHERE host_mid = ?", table), hostMid)
		if result != nil {
			affected, _ := result.RowsAffected()
			totalDeleted += int(affected)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"deleted_count": totalDeleted,
		"message":       fmt.Sprintf("Deleted %d records for host_mid=%s", totalDeleted, hostMid),
	})
}

// Ensure imports used
var _ = strings.TrimSpace
var _ = time.Now
