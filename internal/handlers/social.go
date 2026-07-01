package handlers

import (
	"database/sql"
	"net/http"
	"strconv"

	"bili-history/internal/config"
	"bili-history/internal/db"
	"bili-history/internal/services/biliapi"

	"github.com/gin-gonic/gin"
)

func ensureLikeTable(database *sql.DB) {
	database.Exec(`CREATE TABLE IF NOT EXISTS liked_videos (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		bvid TEXT NOT NULL UNIQUE,
		aid INTEGER,
		title TEXT NOT NULL,
		pic TEXT,
		desc TEXT,
		duration INTEGER DEFAULT 0,
		tid INTEGER DEFAULT 0,
		tname TEXT,
		owner_name TEXT,
		owner_mid INTEGER DEFAULT 0,
		owner_face TEXT,
		pubdate INTEGER DEFAULT 0,
		view INTEGER DEFAULT 0,
		danmaku INTEGER DEFAULT 0,
		like_count INTEGER DEFAULT 0,
		link TEXT,
		fetch_time INTEGER NOT NULL,
		is_seen INTEGER DEFAULT 0
	)`)
	database.Exec(`CREATE INDEX IF NOT EXISTS idx_liked_bvid ON liked_videos(bvid)`)
	database.Exec(`CREATE INDEX IF NOT EXISTS idx_liked_pubdate ON liked_videos(pubdate)`)
}

func ensureWatchLaterTable(database *sql.DB) {
	database.Exec(`CREATE TABLE IF NOT EXISTS watchlater_videos (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		bvid TEXT NOT NULL UNIQUE,
		aid INTEGER,
		title TEXT NOT NULL,
		pic TEXT,
		desc TEXT,
		duration INTEGER DEFAULT 0,
		tid INTEGER DEFAULT 0,
		tname TEXT,
		owner_name TEXT,
		owner_mid INTEGER DEFAULT 0,
		owner_face TEXT,
		add_at INTEGER DEFAULT 0,
		pubdate INTEGER DEFAULT 0,
		view INTEGER DEFAULT 0,
		danmaku INTEGER DEFAULT 0,
		link TEXT,
		fetch_time INTEGER NOT NULL
	)`)
	database.Exec(`CREATE INDEX IF NOT EXISTS idx_wl_bvid ON watchlater_videos(bvid)`)
	database.Exec(`CREATE INDEX IF NOT EXISTS idx_wl_add_at ON watchlater_videos(add_at)`)
}

func GetLikes(c *gin.Context) {
	cfg, err := config.LoadConfig()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Config error"})
		return
	}

	if cfg.SESSDATA == "" || cfg.SESSDATA == "Cookie里的SESSDATA字段值" {
		c.JSON(http.StatusOK, gin.H{"status": "error", "message": "未登录，无法获取点赞列表"})
		return
	}

	path := c.Request.URL.Path
	isLocal := false
	if containsPath(path, "local") {
		isLocal = true
	}

	if isLocal {
		getLikesLocal(c)
		return
	}

	client := newBiliClient(cfg)

	currentPn := 1
	ps := 50
	maxPages := 50
	var allItems []biliapi.LikeVideo
	fetchedCount := 0

	database, err := db.Open(config.GetDatabasePath("bilibili_likes.db"))
	if err == nil {
		defer database.Close()
		ensureLikeTable(database)
		now := getCurrentTimestamp()
		database.Exec("UPDATE liked_videos SET is_seen = 0")

		for currentPn <= maxPages {
			result, err := client.FetchLikes(currentPn, ps)
			if err != nil {
				if currentPn == 1 {
					c.JSON(http.StatusOK, gin.H{"status": "error", "message": err.Error()})
					return
				}
				break
			}

			if len(result.List) == 0 {
				break
			}

			stopEarly := false
			for _, v := range result.List {
				if v.BVID == "" {
					continue
				}

				var isSeen int
				database.QueryRow("SELECT is_seen FROM liked_videos WHERE bvid = ?", v.BVID).Scan(&isSeen)
				if isSeen == 1 {
					stopEarly = true
					break
				}

				var existingID int64
				database.QueryRow("SELECT id FROM liked_videos WHERE bvid = ?", v.BVID).Scan(&existingID)

				if existingID > 0 {
					database.Exec(`UPDATE liked_videos SET
						aid=?, title=?, pic=?, "desc"=?, duration=?, tid=?, tname=?,
						owner_name=?, owner_mid=?, owner_face=?, pubdate=?,
						view=?, danmaku=?, like_count=?, link=?, fetch_time=?, is_seen=1
						WHERE bvid=?`,
						v.AID, v.Title, v.Pic, v.Desc, v.Duration, v.TID, v.TName,
						v.OwnerName, v.OwnerMid, v.OwnerFace, v.PubDate,
						v.View, v.Danmaku, v.LikeCount, v.Link, now, v.BVID)
				} else {
					database.Exec(`INSERT INTO liked_videos
						(bvid, aid, title, pic, "desc", duration, tid, tname,
						 owner_name, owner_mid, owner_face, pubdate,
						 view, danmaku, like_count, link, fetch_time, is_seen)
						VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 1)`,
						v.BVID, v.AID, v.Title, v.Pic, v.Desc, v.Duration, v.TID, v.TName,
						v.OwnerName, v.OwnerMid, v.OwnerFace, v.PubDate,
						v.View, v.Danmaku, v.LikeCount, v.Link, now)
				}

				allItems = append(allItems, v)
				fetchedCount++
			}

			if stopEarly {
				break
			}
			if len(result.List) < ps {
				break
			}
			currentPn++
		}

		var total int
		database.QueryRow("SELECT COUNT(*) FROM liked_videos").Scan(&total)

		c.JSON(http.StatusOK, gin.H{
			"status": "success",
			"data": gin.H{
				"list":    allItems,
				"total":   total,
				"fetched": fetchedCount,
				"pages":   currentPn,
			},
		})
		return
	}

	result, err := client.FetchLikes(1, 50)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"status": "error", "message": err.Error()})
		return
	}

	saveLikesToDB(result.List)

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data": gin.H{
			"list":  result.List,
			"total": len(result.List),
		},
	})
}

func saveLikesToDB(items []biliapi.LikeVideo) {
	database, err := db.Open(config.GetDatabasePath("bilibili_likes.db"))
	if err != nil {
		return
	}
	defer database.Close()

	ensureLikeTable(database)
	now := getCurrentTimestamp()

	for _, v := range items {
		var existingID int64
		database.QueryRow("SELECT id FROM liked_videos WHERE bvid = ?", v.BVID).Scan(&existingID)

		if existingID > 0 {
			database.Exec(`UPDATE liked_videos SET
				aid=?, title=?, pic=?, "desc"=?, duration=?, tid=?, tname=?,
				owner_name=?, owner_mid=?, owner_face=?, pubdate=?,
				view=?, danmaku=?, like_count=?, link=?, fetch_time=?, is_seen=1
				WHERE bvid=?`,
				v.AID, v.Title, v.Pic, v.Desc, v.Duration, v.TID, v.TName,
				v.OwnerName, v.OwnerMid, v.OwnerFace, v.PubDate,
				v.View, v.Danmaku, v.LikeCount, v.Link, now, v.BVID)
		} else {
			database.Exec(`INSERT INTO liked_videos
				(bvid, aid, title, pic, "desc", duration, tid, tname,
				 owner_name, owner_mid, owner_face, pubdate,
				 view, danmaku, like_count, link, fetch_time, is_seen)
				VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 1)`,
				v.BVID, v.AID, v.Title, v.Pic, v.Desc, v.Duration, v.TID, v.TName,
				v.OwnerName, v.OwnerMid, v.OwnerFace, v.PubDate,
				v.View, v.Danmaku, v.LikeCount, v.Link, now)
		}
	}
}

func getLikesLocal(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "50"))
	sort := c.DefaultQuery("sort", "pubdate")
	order := c.DefaultQuery("order", "desc")

	if page < 1 {
		page = 1
	}
	if size < 1 || size > 500 {
		size = 50
	}

	allowedSorts := map[string]bool{
		"pubdate": true, "fetch_time": true, "duration": true,
		"owner_name": true, "view": true, "like_count": true,
	}
	if !allowedSorts[sort] {
		sort = "pubdate"
	}
	if order != "asc" && order != "desc" {
		order = "desc"
	}

	database, err := db.Open(config.GetDatabasePath("bilibili_likes.db"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Database error"})
		return
	}
	defer database.Close()

	ensureLikeTable(database)

	var total int
	database.QueryRow("SELECT COUNT(*) FROM liked_videos").Scan(&total)

	offset := (page - 1) * size
	query := `SELECT bvid, aid, title, pic, "desc", duration, tid, tname,
		owner_name, owner_mid, owner_face, pubdate, view, danmaku, like_count, link, fetch_time
		FROM liked_videos ORDER BY "` + sort + `" ` + order + ` LIMIT ? OFFSET ?`

	rows, err := database.Query(query, size, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Query error"})
		return
	}
	defer rows.Close()

	type likeItem struct {
		BVID      string `json:"bvid"`
		AID       int64  `json:"aid"`
		Title     string `json:"title"`
		Pic       string `json:"pic"`
		Desc      string `json:"desc"`
		Duration  int    `json:"duration"`
		TID       int    `json:"tid"`
		TName     string `json:"tname"`
		OwnerName string `json:"owner_name"`
		OwnerMid  int64  `json:"owner_mid"`
		OwnerFace string `json:"owner_face"`
		PubDate   int64  `json:"pubdate"`
		View      int    `json:"view"`
		Danmaku   int    `json:"danmaku"`
		LikeCount int    `json:"like_count"`
		Link      string `json:"link"`
		FetchTime int64  `json:"fetch_time"`
	}

	var items []likeItem
	for rows.Next() {
		var item likeItem
		rows.Scan(&item.BVID, &item.AID, &item.Title, &item.Pic, &item.Desc,
			&item.Duration, &item.TID, &item.TName, &item.OwnerName, &item.OwnerMid,
			&item.OwnerFace, &item.PubDate, &item.View, &item.Danmaku, &item.LikeCount,
			&item.Link, &item.FetchTime)
		items = append(items, item)
	}

	hasMore := offset+size < total

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data": gin.H{
			"list":     items,
			"total":    total,
			"page":     page,
			"size":     size,
			"has_more": hasMore,
		},
	})
}

func GetWatchLater(c *gin.Context) {
	cfg, err := config.LoadConfig()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Config error"})
		return
	}

	if cfg.SESSDATA == "" || cfg.SESSDATA == "Cookie里的SESSDATA字段值" {
		c.JSON(http.StatusOK, gin.H{"status": "error", "message": "未登录，无法获取稍后再看列表"})
		return
	}

	path := c.Request.URL.Path
	isLocal := false
	if containsPath(path, "local") {
		isLocal = true
	}

	if isLocal {
		getWatchLaterLocal(c)
		return
	}

	client := newBiliClient(cfg)
	result, err := client.FetchWatchLater()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"status": "error", "message": err.Error()})
		return
	}

	saveWatchLaterToDB(result.List)

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data": gin.H{
			"list":  result.List,
			"total": len(result.List),
		},
	})
}

func saveWatchLaterToDB(items []biliapi.WatchLater) {
	database, err := db.Open(config.GetDatabasePath("bilibili_watchlater.db"))
	if err != nil {
		return
	}
	defer database.Close()

	ensureWatchLaterTable(database)
	now := getCurrentTimestamp()

	for _, v := range items {
		var existingID int64
		database.QueryRow("SELECT id FROM watchlater_videos WHERE bvid = ?", v.BVID).Scan(&existingID)

		if existingID > 0 {
			database.Exec(`UPDATE watchlater_videos SET
				aid=?, title=?, pic=?, "desc"=?, duration=?, tid=?, tname=?,
				owner_name=?, owner_mid=?, owner_face=?, add_at=?, pubdate=?,
				view=?, danmaku=?, link=?, fetch_time=?
				WHERE bvid=?`,
				v.AID, v.Title, v.Pic, v.Desc, v.Duration, v.TID, v.TName,
				v.OwnerName, v.OwnerMid, v.OwnerFace, v.AddAt, v.PubDate,
				v.View, v.Danmaku, v.Link, now, v.BVID)
		} else {
			database.Exec(`INSERT INTO watchlater_videos
				(bvid, aid, title, pic, "desc", duration, tid, tname,
				 owner_name, owner_mid, owner_face, add_at, pubdate,
				 view, danmaku, link, fetch_time)
				VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
				v.BVID, v.AID, v.Title, v.Pic, v.Desc, v.Duration, v.TID, v.TName,
				v.OwnerName, v.OwnerMid, v.OwnerFace, v.AddAt, v.PubDate,
				v.View, v.Danmaku, v.Link, now)
		}
	}
}

func getWatchLaterLocal(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "50"))

	if page < 1 {
		page = 1
	}
	if size < 1 || size > 500 {
		size = 50
	}

	database, err := db.Open(config.GetDatabasePath("bilibili_watchlater.db"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Database error"})
		return
	}
	defer database.Close()

	ensureWatchLaterTable(database)

	var total int
	database.QueryRow("SELECT COUNT(*) FROM watchlater_videos").Scan(&total)

	offset := (page - 1) * size
	rows, err := database.Query(`SELECT bvid, aid, title, pic, "desc", duration, tid, tname,
		owner_name, owner_mid, owner_face, add_at, pubdate, view, danmaku, link, fetch_time
		FROM watchlater_videos ORDER BY add_at DESC LIMIT ? OFFSET ?`, size, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Query error"})
		return
	}
	defer rows.Close()

	type wlItem struct {
		BVID      string `json:"bvid"`
		AID       int64  `json:"aid"`
		Title     string `json:"title"`
		Pic       string `json:"pic"`
		Desc      string `json:"desc"`
		Duration  int    `json:"duration"`
		TID       int    `json:"tid"`
		TName     string `json:"tname"`
		OwnerName string `json:"owner_name"`
		OwnerMid  int64  `json:"owner_mid"`
		OwnerFace string `json:"owner_face"`
		AddAt     int64  `json:"add_at"`
		PubDate   int64  `json:"pubdate"`
		View      int    `json:"view"`
		Danmaku   int    `json:"danmaku"`
		Link      string `json:"link"`
		FetchTime int64  `json:"fetch_time"`
	}

	var items []wlItem
	for rows.Next() {
		var item wlItem
		rows.Scan(&item.BVID, &item.AID, &item.Title, &item.Pic, &item.Desc,
			&item.Duration, &item.TID, &item.TName, &item.OwnerName, &item.OwnerMid,
			&item.OwnerFace, &item.AddAt, &item.PubDate, &item.View, &item.Danmaku,
			&item.Link, &item.FetchTime)
		items = append(items, item)
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data": gin.H{
			"list":  items,
			"total": total,
			"page":  page,
			"size":  size,
		},
	})
}
