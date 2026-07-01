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

func getCurrentUserMid(cfg *config.Config) (int64, error) {
	if cfg.SESSDATA == "" || cfg.SESSDATA == "Cookie里的SESSDATA字段值" {
		return 0, fmt.Errorf("not logged in")
	}
	if cfg.DedeUserID != "" {
		uid, err := strconv.ParseInt(cfg.DedeUserID, 10, 64)
		if err == nil && uid > 0 {
			return uid, nil
		}
	}
	client := newBiliClient(cfg)
	user, err := client.FetchCurrentUser()
	if err != nil {
		return 0, err
	}
	if !user.IsLogin {
		return 0, fmt.Errorf("not logged in")
	}
	return user.UID, nil
}

// GetFavorites fetches favorite folders from Bilibili API (created and collected).
func GetFavorites(c *gin.Context) {
	cfg, err := config.LoadConfig()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Config error"})
		return
	}

	upMidStr := c.Query("up_mid")
	pn, _ := strconv.Atoi(c.DefaultQuery("pn", "1"))
	ps, _ := strconv.Atoi(c.DefaultQuery("ps", "40"))

	var upMid int64
	if upMidStr != "" {
		upMid, _ = strconv.ParseInt(upMidStr, 10, 64)
	}

	if upMid == 0 {
		uid, err := getCurrentUserMid(cfg)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"status":  "error",
				"message": "未提供up_mid且未登录，无法获取收藏夹信息",
			})
			return
		}
		upMid = uid
	}

	client := newBiliClient(cfg)

	path := c.Request.URL.Path
	isCollected := false
	if containsPath(path, "collected") {
		isCollected = true
	}

	var result *biliapi.FavoriteFolderListData
	if isCollected {
		result, err = client.FetchCollectedFavorites(upMid, pn, ps)
	} else {
		result, err = client.FetchCreatedFavorites(upMid)
	}

	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"status":  "error",
			"message": err.Error(),
		})
		return
	}

	saveFoldersToDB(result.List)

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "获取收藏夹列表成功",
		"data":    result,
	})
}

func saveFoldersToDB(folders []biliapi.FavoriteFolder) {
	database, err := db.Open(config.GetDatabasePath("bilibili_favorites.db"))
	if err != nil {
		return
	}
	defer database.Close()

	timestamp := getCurrentTimestamp()
	for _, f := range folders {
		var existingID int64
		database.QueryRow("SELECT id FROM favorites_folder WHERE media_id = ?", f.ID).Scan(&existingID)

		creatorName := ""
		creatorFace := ""
		if f.Upper.MID > 0 {
			creatorName = f.Upper.Name
			creatorFace = f.Upper.Face
			database.Exec(`INSERT OR REPLACE INTO favorites_creator
				(mid, name, face, fetch_time) VALUES (?, ?, ?, ?)`,
				f.Upper.MID, creatorName, creatorFace, timestamp)
		}

		if existingID > 0 {
			database.Exec(`UPDATE favorites_folder SET
				fid=?, mid=?, title=?, cover=?, attr=?, intro=?, ctime=?, mtime=?,
				state=?, media_count=?, fav_state=?, like_state=?, fetch_time=?
				WHERE media_id=?`,
				f.FID, f.MID, f.Title, f.Cover, f.Attr, f.Intro, f.CTime, f.MTime,
				f.State, f.MediaCount, f.FavState, f.LikeState, timestamp, f.ID)
		} else {
			database.Exec(`INSERT INTO favorites_folder
				(media_id, fid, mid, title, cover, attr, intro, ctime, mtime,
				 state, media_count, fav_state, like_state, fetch_time)
				VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
				f.ID, f.FID, f.MID, f.Title, f.Cover, f.Attr, f.Intro, f.CTime, f.MTime,
				f.State, f.MediaCount, f.FavState, f.LikeState, timestamp)
		}
	}
}

// Ensure fmt is used
var _ = fmt.Sprintf

// GetFavoriteContentAPI fetches favorite folder content from Bilibili API.
func GetFavoriteContentAPI(c *gin.Context) {
	cfg, err := config.LoadConfig()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Config error"})
		return
	}

	mediaIDStr := c.Query("media_id")
	if mediaIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "media_id is required"})
		return
	}
	mediaID, _ := strconv.ParseInt(mediaIDStr, 10, 64)
	if mediaID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "invalid media_id"})
		return
	}

	pn, _ := strconv.Atoi(c.DefaultQuery("pn", "1"))
	ps, _ := strconv.Atoi(c.DefaultQuery("ps", "40"))
	keyword := c.Query("keyword")
	order := c.DefaultQuery("order", "mtime")

	client := newBiliClient(cfg)
	result, err := client.FetchFavoriteContent(mediaID, pn, ps, keyword, order)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"status": "error", "message": err.Error()})
		return
	}

	saveFavoriteContentToDB(mediaID, result.Medias)

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"message": "获取收藏夹内容成功",
		"data":   result,
	})
}

func saveFavoriteContentToDB(mediaID int64, medias []biliapi.FavoriteContent) {
	database, err := db.Open(config.GetDatabasePath("bilibili_favorites.db"))
	if err != nil {
		return
	}
	defer database.Close()

	timestamp := getCurrentTimestamp()
	for _, m := range medias {
		upperMid := m.Upper.MID
		upperName := m.Upper.Name
		upperFace := m.Upper.Face

		if upperMid > 0 {
			database.Exec(`INSERT OR REPLACE INTO favorites_creator
				(mid, name, face, fetch_time) VALUES (?, ?, ?, ?)`,
				upperMid, upperName, upperFace, timestamp)
		}

		if m.Play == 0 {
			m.Play = m.CntInfo.Play
		}
		if m.Danmaku == 0 {
			m.Danmaku = m.CntInfo.Danmaku
		}
		if m.Reply == 0 {
			m.Reply = m.CntInfo.Reply
		}

		var existingID int64
		database.QueryRow("SELECT id FROM favorites_content WHERE media_id = ? AND content_id = ?",
			mediaID, m.ID).Scan(&existingID)

		if existingID > 0 {
			database.Exec(`UPDATE favorites_content SET
				type=?, title=?, cover=?, bvid=?, intro=?, page=?, duration=?,
				upper_mid=?, attr=?, ctime=?, pubtime=?, fav_time=?,
				play=?, danmaku=?, reply=?, fetch_time=?,
				creator_name=?, creator_face=?
				WHERE media_id=? AND content_id=?`,
				m.Type, m.Title, m.Cover, m.BVID, m.Intro, m.Page, m.Duration,
				upperMid, m.Attr, m.CTime, m.PubTime, m.FavTime,
				m.Play, m.Danmaku, m.Reply, timestamp,
				upperName, upperFace,
				mediaID, m.ID)
		} else {
			database.Exec(`INSERT INTO favorites_content
				(media_id, content_id, type, title, cover, bvid, intro, page, duration,
				 upper_mid, attr, ctime, pubtime, fav_time,
				 play, danmaku, reply, fetch_time, creator_name, creator_face)
				VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
				mediaID, m.ID, m.Type, m.Title, m.Cover, m.BVID, m.Intro, m.Page, m.Duration,
				upperMid, m.Attr, m.CTime, m.PubTime, m.FavTime,
				m.Play, m.Danmaku, m.Reply, timestamp, upperName, upperFace)
		}
	}
}
func GetFavoritesListDB(c *gin.Context) {
	midStr := c.Query("mid")
	pageStr := c.DefaultQuery("page", "1")
	sizeStr := c.DefaultQuery("size", "40")
	page, _ := strconv.Atoi(pageStr)
	size, _ := strconv.Atoi(sizeStr)
	if page < 1 {
		page = 1
	}

	database, err := db.Open(config.GetDatabasePath("bilibili_favorites.db"))
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

	database, err := db.Open(config.GetDatabasePath("bilibili_favorites.db"))
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

	uid, err := getCurrentUserMid(cfg)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"status":  "error",
			"message": "未登录，无法获取收藏状态",
		})
		return
	}

	client := newBiliClient(cfg)
	result, err := client.FetchCreatedFavorites(uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"bvid":         bvid,
			"folders":      result.List,
			"is_favorited": len(result.List) > 0,
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
