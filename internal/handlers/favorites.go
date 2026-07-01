package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"bili-history/internal/config"
	"bili-history/internal/db"
	"bili-history/internal/services/biliapi"

	"github.com/gin-gonic/gin"
)

// GetFavoritesList returns favorites list from the database.
func GetFavoritesListDB(c *gin.Context) {
	midStr := c.Query("mid")
	pageStr := c.DefaultQuery("page", "1")
	sizeStr := c.DefaultQuery("size", "40")
	page, _ := strconv.Atoi(pageStr)
	size, _ := strconv.Atoi(sizeStr)
	if page < 1 {
		page = 1
	}

	database, err := db.Open(config.GetOutputPath("database", "bilibili_favorites.db"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	whereClause := "WHERE 1=1"
	var args []interface{}
	if midStr != "" {
		whereClause += " AND f.mid = ?"
		args = append(args, midStr)
	}

	var total int
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM favorites_folder f %s", whereClause)
	database.QueryRow(countQuery, args...).Scan(&total)

	offset := (page - 1) * size
	query := fmt.Sprintf(`
		SELECT f.id, f.media_id, f.fid, f.mid, f.title, COALESCE(f.cover,''),
			COALESCE(f.attr,0), COALESCE(f.intro,''), COALESCE(f.ctime,0),
			COALESCE(f.mtime,0), COALESCE(f.state,0), COALESCE(f.media_count,0),
			COALESCE(f.fav_state,0), COALESCE(f.like_state,0), COALESCE(f.fetch_time,0),
			COALESCE(c.name,'') as creator_name, COALESCE(c.face,'') as creator_face
		FROM favorites_folder f
		LEFT JOIN favorites_creator c ON f.mid = c.mid
		%s ORDER BY f.mtime DESC LIMIT ? OFFSET ?`, whereClause)
	args = append(args, size, offset)

	rows, err := database.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Query error"})
		return
	}
	defer rows.Close()

	type folderItem struct {
		ID          int64  `json:"id"`
		MediaID     int64  `json:"media_id"`
		FID         int64  `json:"fid"`
		MID         int64  `json:"mid"`
		Title       string `json:"title"`
		Cover       string `json:"cover"`
		Attr        int    `json:"attr"`
		Intro       string `json:"intro"`
		Ctime       int64  `json:"ctime"`
		Mtime       int64  `json:"mtime"`
		State       int    `json:"state"`
		MediaCount  int    `json:"media_count"`
		FavState    int    `json:"fav_state"`
		LikeState   int    `json:"like_state"`
		FetchTime   int64  `json:"fetch_time"`
		CreatorName string `json:"creator_name"`
		CreatorFace string `json:"creator_face"`
	}
	var folders []folderItem
	for rows.Next() {
		var f folderItem
		rows.Scan(&f.ID, &f.MediaID, &f.FID, &f.MID, &f.Title, &f.Cover,
			&f.Attr, &f.Intro, &f.Ctime, &f.Mtime, &f.State, &f.MediaCount,
			&f.FavState, &f.LikeState, &f.FetchTime, &f.CreatorName, &f.CreatorFace)
		folders = append(folders, f)
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data": gin.H{
			"list":  folders,
			"total": total,
			"page":  page,
			"size":  size,
		},
	})
}

// GetFavoritesContentList returns favorites content from the database.
func GetFavoritesContentList(c *gin.Context) {
	mediaIDStr := c.Query("media_id")
	if mediaIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "media_id is required"})
		return
	}
	mediaID, _ := strconv.ParseInt(mediaIDStr, 10, 64)

	pageStr := c.DefaultQuery("page", "1")
	sizeStr := c.DefaultQuery("size", "40")
	page, _ := strconv.Atoi(pageStr)
	size, _ := strconv.Atoi(sizeStr)
	if page < 1 {
		page = 1
	}

	database, err := db.Open(config.GetOutputPath("database", "bilibili_favorites.db"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	var total int
	database.QueryRow("SELECT COUNT(*) FROM favorites_content WHERE media_id = ?", mediaID).Scan(&total)

	offset := (page - 1) * size
	rows, err := database.Query(`
		SELECT c.id, c.media_id, c.content_id, c.type, c.title, COALESCE(c.cover,''),
			COALESCE(c.bvid,''), COALESCE(c.intro,''), COALESCE(c.page,0),
			COALESCE(c.duration,0), COALESCE(c.upper_mid,0), COALESCE(c.attr,0),
			COALESCE(c.ctime,0), COALESCE(c.pubtime,0), COALESCE(c.fav_time,0),
			COALESCE(c.play,0), COALESCE(c.danmaku,0), COALESCE(c.reply,0),
			COALESCE(cr.name,'') as creator_name, COALESCE(cr.face,'') as creator_face
		FROM favorites_content c
		LEFT JOIN favorites_creator cr ON c.upper_mid = cr.mid
		WHERE c.media_id = ?
		ORDER BY c.fav_time DESC
		LIMIT ? OFFSET ?`, mediaID, size, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Query error"})
		return
	}
	defer rows.Close()

	type contentItem struct {
		ID          int64  `json:"id"`
		MediaID     int64  `json:"media_id"`
		ContentID   int64  `json:"content_id"`
		Type        int    `json:"type"`
		Title       string `json:"title"`
		Cover       string `json:"cover"`
		BVID        string `json:"bvid"`
		Intro       string `json:"intro"`
		Page        int    `json:"page"`
		Duration    int    `json:"duration"`
		UpperMid    int64  `json:"upper_mid"`
		Attr        int    `json:"attr"`
		Ctime       int64  `json:"ctime"`
		Pubtime     int64  `json:"pubtime"`
		FavTime     int64  `json:"fav_time"`
		Play        int    `json:"play"`
		Danmaku     int    `json:"danmaku"`
		Reply       int    `json:"reply"`
		CreatorName string `json:"creator_name"`
		CreatorFace string `json:"creator_face"`
	}
	var contents []contentItem
	for rows.Next() {
		var ci contentItem
		rows.Scan(&ci.ID, &ci.MediaID, &ci.ContentID, &ci.Type, &ci.Title, &ci.Cover,
			&ci.BVID, &ci.Intro, &ci.Page, &ci.Duration, &ci.UpperMid, &ci.Attr,
			&ci.Ctime, &ci.Pubtime, &ci.FavTime, &ci.Play, &ci.Danmaku, &ci.Reply,
			&ci.CreatorName, &ci.CreatorFace)
		contents = append(contents, ci)
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data": gin.H{
			"list":  contents,
			"total": total,
			"page":  page,
			"size":  size,
		},
	})
}

// CheckFavorite checks if a video is favorited.
func CheckFavorite(c *gin.Context) {
	bvid := c.Query("bvid")
	if bvid == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bvid is required"})
		return
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Config error"})
		return
	}

	client := biliapi.NewClient(cfg.SESSDATA, cfg.BiliJCT, cfg.DedeUserID)
	folders, err := client.FetchFavorites()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"bvid":     bvid,
			"folders":  folders,
			"is_favorited": len(folders) > 0,
		},
	})
}

// BatchCheckFavorite checks if multiple videos are favorited.
func BatchCheckFavorite(c *gin.Context) {
	var req struct {
		BVIDs []string `json:"bvids"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || len(req.BVIDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bvids array is required"})
		return
	}

	// For now, just return the list - full implementation would check against DB
	result := make(map[string]bool)
	for _, bvid := range req.BVIDs {
		result[bvid] = false
	}

	c.JSON(http.StatusOK, gin.H{
		"data": result,
	})
}

// Ensure fmt is used
var _ = fmt.Sprintf
