package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"bilibili-history-go/utils"
)

type ExtraDB struct {
	db   *sql.DB
	path string
	mu   sync.RWMutex
}

var (
	likesDB     *ExtraDB
	likesOnce   sync.Once
	watchlaterDB *ExtraDB
	watchlaterOnce sync.Once
	favoritesDB *ExtraDB
	favoritesOnce sync.Once
)

func getExtraDB(dbFileName string) *sql.DB {
	dbPath := utils.GetDatabasePath(dbFileName)

	dir := filepath.Dir(dbPath)
	os.MkdirAll(dir, 0755)

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		utils.LogError("Failed to open database %s: %v", dbFileName, err)
		return nil
	}

	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	return db
}

// ensureWatchLaterSchema creates the watchlater_videos table if it does not exist.
// This keeps the Go backend self-sufficient even when the DB file is fresh.
const watchLaterSchema = `
CREATE TABLE IF NOT EXISTS watchlater_videos (
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
);
CREATE INDEX IF NOT EXISTS idx_wl_bvid ON watchlater_videos(bvid);
CREATE INDEX IF NOT EXISTS idx_wl_add_at ON watchlater_videos(add_at);
CREATE INDEX IF NOT EXISTS idx_wl_owner ON watchlater_videos(owner_name);
CREATE INDEX IF NOT EXISTS idx_wl_tid ON watchlater_videos(tid);
CREATE INDEX IF NOT EXISTS idx_wl_fetch_time ON watchlater_videos(fetch_time);
`

const favoritesFolderSchema = `
CREATE TABLE IF NOT EXISTS favorites_folder (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    media_id INTEGER NOT NULL UNIQUE,
    fid INTEGER DEFAULT 0,
    mid INTEGER DEFAULT 0,
    title TEXT,
    cover TEXT,
    attr INTEGER DEFAULT 0,
    intro TEXT,
    ctime INTEGER DEFAULT 0,
    mtime INTEGER DEFAULT 0,
    state INTEGER DEFAULT 0,
    media_count INTEGER DEFAULT 0,
    fav_state INTEGER DEFAULT 0,
    like_state INTEGER DEFAULT 0,
    folder_type INTEGER DEFAULT 0,
    fetch_time INTEGER NOT NULL
);
`

const favoritesContentSchema = `
CREATE TABLE IF NOT EXISTS favorites_content (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    media_id INTEGER NOT NULL,
    content_id INTEGER DEFAULT 0,
    type INTEGER DEFAULT 0,
    title TEXT,
    cover TEXT,
    bvid TEXT,
    intro TEXT,
    page INTEGER DEFAULT 0,
    duration INTEGER DEFAULT 0,
    upper_mid INTEGER DEFAULT 0,
    attr INTEGER DEFAULT 0,
    ctime INTEGER DEFAULT 0,
    pubtime INTEGER DEFAULT 0,
    fav_time INTEGER DEFAULT 0,
    link TEXT,
    fetch_time INTEGER NOT NULL,
    creator_name TEXT,
    creator_face TEXT,
    bv_id TEXT,
    collect INTEGER DEFAULT 0,
    play INTEGER DEFAULT 0,
    danmaku INTEGER DEFAULT 0,
    play_switch INTEGER DEFAULT 0,
    reply INTEGER DEFAULT 0,
    view_text_1 TEXT,
    first_cid INTEGER DEFAULT 0,
    media_list_link TEXT,
    UNIQUE(media_id, content_id)
);
CREATE INDEX IF NOT EXISTS idx_fc_media_id ON favorites_content(media_id);
CREATE INDEX IF NOT EXISTS idx_fc_bvid ON favorites_content(bvid);
`

const likedVideosSchema = `
CREATE TABLE IF NOT EXISTS liked_videos (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    bvid TEXT NOT NULL UNIQUE,
    aid INTEGER DEFAULT 0,
    title TEXT,
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
);
CREATE INDEX IF NOT EXISTS idx_lv_bvid ON liked_videos(bvid);
CREATE INDEX IF NOT EXISTS idx_lv_fetch_time ON liked_videos(fetch_time);
`

func GetLikesDB() *sql.DB {
	likesOnce.Do(func() {
		db := getExtraDB("bilibili_likes.db")
		if db != nil {
			if _, err := db.Exec(likedVideosSchema); err != nil {
				utils.LogError("Failed to ensure likes schema: %v", err)
			}
		}
		likesDB = &ExtraDB{
			db: db,
		}
	})
	if likesDB == nil {
		return nil
	}
	return likesDB.db
}

func GetWatchLaterDB() *sql.DB {
	watchlaterOnce.Do(func() {
		db := getExtraDB("bilibili_watchlater.db")
		if db != nil {
			if _, err := db.Exec(watchLaterSchema); err != nil {
				utils.LogError("Failed to ensure watchlater schema: %v", err)
			}
		}
		watchlaterDB = &ExtraDB{
			db: db,
		}
	})
	if watchlaterDB == nil {
		return nil
	}
	return watchlaterDB.db
}

func GetFavoritesDB() *sql.DB {
	favoritesOnce.Do(func() {
		db := getExtraDB("bilibili_favorites.db")
		if db != nil {
			if _, err := db.Exec(favoritesFolderSchema); err != nil {
				utils.LogError("Failed to ensure favorites folder schema: %v", err)
			}
			if _, err := db.Exec(favoritesContentSchema); err != nil {
				utils.LogError("Failed to ensure favorites content schema: %v", err)
			}
			// Migration: add folder_type column for existing tables
			var colCount int
			_ = db.QueryRow("SELECT COUNT(*) FROM pragma_table_info('favorites_folder') WHERE name='folder_type'").Scan(&colCount)
			if colCount == 0 {
				_, _ = db.Exec("ALTER TABLE favorites_folder ADD COLUMN folder_type INTEGER DEFAULT 0")
			}
		}
		favoritesDB = &ExtraDB{
			db: db,
		}
	})
	if favoritesDB == nil {
		return nil
	}
	return favoritesDB.db
}

type LikeVideo struct {
	ID          int64  `json:"id"`
	Bvid        string `json:"bvid"`
	Aid         int64  `json:"aid"`
	Title       string `json:"title"`
	Pic         string `json:"pic"`
	Desc        string `json:"desc"`
	Duration    int    `json:"duration"`
	Tid         int    `json:"tid"`
	Tname       string `json:"tname"`
	OwnerName   string `json:"owner_name"`
	OwnerMid    int64  `json:"owner_mid"`
	OwnerFace   string `json:"owner_face"`
	Pubdate     int64  `json:"pubdate"`
	View        int    `json:"view"`
	Danmaku     int    `json:"danmaku"`
	LikeCount   int    `json:"like_count"`
	Link        string `json:"link"`
	FetchTime   int64  `json:"fetch_time"`
	IsSeen      int    `json:"is_seen"`
}

type WatchLaterVideo struct {
	ID          int64  `json:"id"`
	Bvid        string `json:"bvid"`
	Aid         int64  `json:"aid"`
	Title       string `json:"title"`
	Pic         string `json:"pic"`
	Desc        string `json:"desc"`
	Duration    int    `json:"duration"`
	Tid         int    `json:"tid"`
	Tname       string `json:"tname"`
	OwnerName   string `json:"owner_name"`
	OwnerMid    int64  `json:"owner_mid"`
	OwnerFace   string `json:"owner_face"`
	AddAt       int64  `json:"add_at"`
	Pubdate     int64  `json:"pubdate"`
	View        int    `json:"view"`
	Danmaku     int    `json:"danmaku"`
	Link        string `json:"link"`
	FetchTime   int64  `json:"fetch_time"`
}

type FavoriteFolder struct {
	ID          int64  `json:"id"`
	MediaID     int64  `json:"media_id"`
	Fid         int64  `json:"fid"`
	Mid         int64  `json:"mid"`
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
	FolderType  int    `json:"folder_type"` // 0=created, 1=collected
	FetchTime   int64  `json:"fetch_time"`
}

type FavoriteContent struct {
	ID            int64  `json:"id"`
	MediaID       int64  `json:"media_id"`
	ContentID     int64  `json:"content_id"`
	Type          int    `json:"type"`
	Title         string `json:"title"`
	Cover         string `json:"cover"`
	Bvid          string `json:"bvid"`
	Intro         string `json:"intro"`
	Page          int    `json:"page"`
	Duration      int    `json:"duration"`
	UpperMid      int64  `json:"upper_mid"`
	Attr          int    `json:"attr"`
	Ctime         int64  `json:"ctime"`
	Pubtime       int64  `json:"pubtime"`
	FavTime       int64  `json:"fav_time"`
	Link          string `json:"link"`
	FetchTime     int64  `json:"fetch_time"`
	CreatorName   string `json:"creator_name"`
	CreatorFace   string `json:"creator_face"`
	BvID          string `json:"bv_id"`
	Collect       int    `json:"collect"`
	Play          int    `json:"play"`
	Danmaku       int    `json:"danmaku"`
	PlaySwitch    int    `json:"play_switch"`
	Reply         int    `json:"reply"`
	ViewText1     string `json:"view_text_1"`
	FirstCid      int64  `json:"first_cid"`
	MediaListLink string `json:"media_list_link"`
}

func GetLikedVideos(page, size int, sort, order string) ([]map[string]interface{}, int, error) {
	db := GetLikesDB()
	if db == nil {
		return []map[string]interface{}{}, 0, nil
	}

	var total int
	err := db.QueryRow("SELECT COUNT(*) FROM liked_videos").Scan(&total)
	if err != nil {
		return []map[string]interface{}{}, 0, err
	}

	orderBy := "fetch_time"
	sortDir := "DESC"
	if sort != "" {
		switch sort {
		case "pubdate":
			orderBy = "pubdate"
		case "fetch_time":
			orderBy = "fetch_time"
		case "duration":
			orderBy = "duration"
		case "view":
			orderBy = "view"
		}
	}
	if order == "asc" {
		sortDir = "ASC"
	}

	offset := (page - 1) * size
	query := fmt.Sprintf(`
		SELECT * FROM liked_videos
		ORDER BY %s %s
		LIMIT ? OFFSET ?
	`, orderBy, sortDir)

	rows, err := db.Query(query, size, offset)
	if err != nil {
		return []map[string]interface{}{}, 0, err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return []map[string]interface{}{}, 0, err
	}

	var results []map[string]interface{}
	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range columns {
			valuePtrs[i] = &values[i]
		}
		if err := rows.Scan(valuePtrs...); err != nil {
			continue
		}

		result := make(map[string]interface{})
		for i, col := range columns {
			val := values[i]
			if val == nil {
				continue
			}
			switch v := val.(type) {
			case []byte:
				result[col] = string(v)
			default:
				result[col] = v
			}
		}
		results = append(results, result)
	}

	return results, total, nil
}

// SaveFavoriteFolders upserts folders into the local favorites DB, matched by media_id.
func SaveFavoriteFolders(folders []FavoriteFolder) error {
	db := GetFavoritesDB()
	if db == nil {
		return fmt.Errorf("favorites database not available")
	}
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	now := time.Now().Unix()
	stmt, err := tx.Prepare(`INSERT INTO favorites_folder
		(media_id, fid, mid, title, cover, attr, intro, ctime, mtime, state, media_count, fav_state, like_state, folder_type, fetch_time)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(media_id) DO UPDATE SET
			fid=excluded.fid, mid=excluded.mid, title=excluded.title, cover=excluded.cover,
			attr=excluded.attr, intro=excluded.intro, ctime=excluded.ctime, mtime=excluded.mtime,
			state=excluded.state, media_count=excluded.media_count, fav_state=excluded.fav_state,
			like_state=excluded.like_state, folder_type=excluded.folder_type, fetch_time=excluded.fetch_time`)
	if err != nil {
		return fmt.Errorf("prepare stmt: %w", err)
	}
	defer stmt.Close()

	liveIDs := make([]int64, 0, len(folders))
	for _, f := range folders {
		_, err := stmt.Exec(f.MediaID, f.Fid, f.Mid, f.Title, f.Cover, f.Attr, f.Intro,
			f.Ctime, f.Mtime, f.State, f.MediaCount, f.FavState, f.LikeState, f.FolderType, now)
		if err != nil {
			utils.LogError("Failed to upsert favorite folder %d: %v", f.MediaID, err)
			continue
		}
		liveIDs = append(liveIDs, f.MediaID)
	}

	if len(liveIDs) > 0 {
		placeholders := make([]string, len(liveIDs))
		args := make([]interface{}, len(liveIDs))
		for i, id := range liveIDs {
			placeholders[i] = "?"
			args[i] = id
		}
		query := "DELETE FROM favorites_folder WHERE media_id NOT IN (" + strings.Join(placeholders, ",") + ")"
		if _, err := tx.Exec(query, args...); err != nil {
			utils.LogError("Failed to prune stale favorite folders: %v", err)
		}
	} else {
		if _, err := tx.Exec("DELETE FROM favorites_folder"); err != nil {
			utils.LogError("Failed to clear favorite folders: %v", err)
		}
	}

	return tx.Commit()
}

// SaveFavoriteContents upserts contents for a specific media_id into the local favorites DB.
// Existing records are updated in place; new records are inserted. Local records
// not present in the given list are kept (incremental merge, no pruning).
func SaveFavoriteContents(mediaID int64, contents []FavoriteContent) error {
	db := GetFavoritesDB()
	if db == nil {
		return fmt.Errorf("favorites database not available")
	}
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	now := time.Now().Unix()
	stmt, err := tx.Prepare(`INSERT INTO favorites_content
		(media_id, content_id, type, title, cover, bvid, intro, page, duration, upper_mid, attr,
		 ctime, pubtime, fav_time, link, fetch_time, creator_name, creator_face, bv_id,
		 collect, play, danmaku, play_switch, reply, view_text_1, first_cid, media_list_link)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(media_id, content_id) DO UPDATE SET
			type=excluded.type, title=excluded.title, cover=excluded.cover, bvid=excluded.bvid,
			intro=excluded.intro, page=excluded.page, duration=excluded.duration, upper_mid=excluded.upper_mid,
			attr=excluded.attr, ctime=excluded.ctime, pubtime=excluded.pubtime, fav_time=excluded.fav_time,
			link=excluded.link, fetch_time=excluded.fetch_time, creator_name=excluded.creator_name,
			creator_face=excluded.creator_face, bv_id=excluded.bv_id, collect=excluded.collect,
			play=excluded.play, danmaku=excluded.danmaku, play_switch=excluded.play_switch,
			reply=excluded.reply, view_text_1=excluded.view_text_1, first_cid=excluded.first_cid,
			media_list_link=excluded.media_list_link`)
	if err != nil {
		return fmt.Errorf("prepare stmt: %w", err)
	}
	defer stmt.Close()

	for _, c := range contents {
		// 如果新标题是"已失效视频"，尝试保留原标题
		if c.Title == "已失效视频" || c.Title == "" {
			var existingTitle string
			err := db.QueryRow("SELECT title FROM favorites_content WHERE media_id = ? AND content_id = ?",
				c.MediaID, c.ContentID).Scan(&existingTitle)
			if err == nil && existingTitle != "" && existingTitle != "已失效视频" {
				c.Title = existingTitle
			}
		}

		_, err := stmt.Exec(c.MediaID, c.ContentID, c.Type, c.Title, c.Cover, c.Bvid, c.Intro,
			c.Page, c.Duration, c.UpperMid, c.Attr, c.Ctime, c.Pubtime, c.FavTime, c.Link, now,
			c.CreatorName, c.CreatorFace, c.BvID, c.Collect, c.Play, c.Danmaku, c.PlaySwitch,
			c.Reply, c.ViewText1, c.FirstCid, c.MediaListLink)
		if err != nil {
			utils.LogError("Failed to upsert favorite content %d/%d: %v", c.MediaID, c.ContentID, err)
			continue
		}
	}

	return tx.Commit()
}

// SaveLikedVideos upserts liked videos into the local DB, matched by bvid.
func SaveLikedVideos(videos []LikeVideo) error {
	db := GetLikesDB()
	if db == nil {
		return fmt.Errorf("likes database not available")
	}
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	now := time.Now().Unix()
	stmt, err := tx.Prepare(`INSERT INTO liked_videos
		(bvid, aid, title, pic, desc, duration, tid, tname, owner_name, owner_mid, owner_face,
		 pubdate, view, danmaku, like_count, link, fetch_time, is_seen)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(bvid) DO UPDATE SET
			aid=excluded.aid, title=excluded.title, pic=excluded.pic, desc=excluded.desc,
			duration=excluded.duration, tid=excluded.tid, tname=excluded.tname,
			owner_name=excluded.owner_name, owner_mid=excluded.owner_mid, owner_face=excluded.owner_face,
			pubdate=excluded.pubdate, view=excluded.view, danmaku=excluded.danmaku,
			like_count=excluded.like_count, link=excluded.link, fetch_time=excluded.fetch_time`)
	if err != nil {
		return fmt.Errorf("prepare stmt: %w", err)
	}
	defer stmt.Close()

	for _, v := range videos {
		if v.Bvid == "" {
			continue
		}
		link := "https://www.bilibili.com/video/" + v.Bvid
		_, err := stmt.Exec(v.Bvid, v.Aid, v.Title, v.Pic, v.Desc, v.Duration, v.Tid, v.Tname,
			v.OwnerName, v.OwnerMid, v.OwnerFace, v.Pubdate, v.View, v.Danmaku, v.LikeCount,
			link, now, v.IsSeen)
		if err != nil {
			utils.LogError("Failed to upsert liked video %s: %v", v.Bvid, err)
			continue
		}
	}

	return tx.Commit()
}

func GetWatchLaterVideos(page, size int, sort, order string) ([]map[string]interface{}, int, error) {
	db := GetWatchLaterDB()
	if db == nil {
		return []map[string]interface{}{}, 0, nil
	}

	var total int
	err := db.QueryRow("SELECT COUNT(*) FROM watchlater_videos").Scan(&total)
	if err != nil {
		return []map[string]interface{}{}, 0, err
	}

	orderBy := "add_at"
	sortDir := "DESC"
	if sort != "" {
		switch sort {
		case "pubdate":
			orderBy = "pubdate"
		case "add_at":
			orderBy = "add_at"
		case "fetch_time":
			orderBy = "fetch_time"
		case "duration":
			orderBy = "duration"
		case "view":
			orderBy = "view"
		}
	}
	if order == "asc" {
		sortDir = "ASC"
	}

	offset := (page - 1) * size
	query := fmt.Sprintf(`
		SELECT * FROM watchlater_videos
		ORDER BY %s %s
		LIMIT ? OFFSET ?
	`, orderBy, sortDir)

	rows, err := db.Query(query, size, offset)
	if err != nil {
		return []map[string]interface{}{}, 0, err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return []map[string]interface{}{}, 0, err
	}

	var results []map[string]interface{}
	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range columns {
			valuePtrs[i] = &values[i]
		}
		if err := rows.Scan(valuePtrs...); err != nil {
			continue
		}

		result := make(map[string]interface{})
		for i, col := range columns {
			val := values[i]
			if val == nil {
				continue
			}
			switch v := val.(type) {
			case []byte:
				result[col] = string(v)
			default:
				result[col] = v
			}
		}
		results = append(results, result)
	}

	return results, total, nil
}

// SaveWatchLaterVideos performs a differential sync: upserts all remote items,
// then deletes local items that are genuinely absent from the remote list.
// Unlike a blind prune, this only runs after a successful full fetch, so partial
// failures won't accidentally wipe local data.
func SaveWatchLaterVideos(videos []WatchLaterVideo) error {
	db := GetWatchLaterDB()
	if db == nil {
		return fmt.Errorf("watchlater database not available")
	}

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	now := time.Now().Unix()

	// Step 1: upsert all remote items
	stmt, err := tx.Prepare(`INSERT INTO watchlater_videos
		(bvid, aid, title, pic, desc, duration, tid, tname, owner_name, owner_mid, owner_face, add_at, pubdate, view, danmaku, link, fetch_time)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(bvid) DO UPDATE SET
			aid=excluded.aid, title=excluded.title, pic=excluded.pic, desc=excluded.desc,
			duration=excluded.duration, tid=excluded.tid, tname=excluded.tname,
			owner_name=excluded.owner_name, owner_mid=excluded.owner_mid, owner_face=excluded.owner_face,
			add_at=excluded.add_at, pubdate=excluded.pubdate, view=excluded.view,
			danmaku=excluded.danmaku, link=excluded.link, fetch_time=excluded.fetch_time`)
	if err != nil {
		return fmt.Errorf("prepare stmt: %w", err)
	}
	defer stmt.Close()

	remoteBvids := make(map[string]bool, len(videos))
	for _, v := range videos {
		if v.Bvid == "" {
			continue
		}
		remoteBvids[v.Bvid] = true
		_, err := stmt.Exec(v.Bvid, v.Aid, v.Title, v.Pic, v.Desc, v.Duration, v.Tid, v.Tname,
			v.OwnerName, v.OwnerMid, v.OwnerFace, v.AddAt, v.Pubdate, v.View, v.Danmaku, v.Link, now)
		if err != nil {
			utils.LogError("Failed to upsert watchlater %s: %v", v.Bvid, err)
		}
	}

	// Step 2: delete local items not in remote list
	var localBvids []string
	rows, err := tx.Query("SELECT bvid FROM watchlater_videos")
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var bvid string
			if rows.Scan(&bvid) == nil {
				if !remoteBvids[bvid] {
					localBvids = append(localBvids, bvid)
				}
			}
		}
	}
	if len(localBvids) > 0 {
		placeholders := make([]string, len(localBvids))
		args := make([]interface{}, len(localBvids))
		for i, b := range localBvids {
			placeholders[i] = "?"
			args[i] = b
		}
		query := "DELETE FROM watchlater_videos WHERE bvid IN (" + strings.Join(placeholders, ",") + ")"
		if _, err := tx.Exec(query, args...); err != nil {
			utils.LogError("Failed to delete removed watchlater items: %v", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit: %w", err)
	}
	return nil
}

// DeleteWatchLaterVideo removes a single watch later row from the local cache by bvid.
func DeleteWatchLaterVideo(bvid string) error {
	db := GetWatchLaterDB()
	if db == nil {
		return fmt.Errorf("watchlater database not available")
	}
	_, err := db.Exec("DELETE FROM watchlater_videos WHERE bvid = ?", bvid)
	return err
}

// GetWatchLaterVideoByBvid returns a single watch later row by bvid, or nil if not found.
func GetWatchLaterVideoByBvid(bvid string) (*WatchLaterVideo, error) {
	db := GetWatchLaterDB()
	if db == nil {
		return nil, nil
	}
	row := db.QueryRow(`SELECT bvid, aid, title, pic, desc, duration, tid, tname,
		owner_name, owner_mid, owner_face, add_at, pubdate, view, danmaku, link, fetch_time
		FROM watchlater_videos WHERE bvid = ?`, bvid)
	var v WatchLaterVideo
	err := row.Scan(&v.Bvid, &v.Aid, &v.Title, &v.Pic, &v.Desc, &v.Duration, &v.Tid, &v.Tname,
		&v.OwnerName, &v.OwnerMid, &v.OwnerFace, &v.AddAt, &v.Pubdate, &v.View, &v.Danmaku, &v.Link, &v.FetchTime)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &v, nil
}

func GetFavoriteFolders(created bool) ([]map[string]interface{}, int, error) {
	db := GetFavoritesDB()
	if db == nil {
		return []map[string]interface{}{}, 0, nil
	}

	var total int
	var rows *sql.Rows
	var err error

	if created {
		err = db.QueryRow("SELECT COUNT(*) FROM favorites_folder").Scan(&total)
		if err != nil {
			return []map[string]interface{}{}, 0, err
		}
		rows, err = db.Query("SELECT * FROM favorites_folder ORDER BY mtime DESC")
	} else {
		return []map[string]interface{}{}, 0, nil
	}

	if err != nil {
		return []map[string]interface{}{}, 0, err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return []map[string]interface{}{}, 0, err
	}

	var results []map[string]interface{}
	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range columns {
			valuePtrs[i] = &values[i]
		}
		if err := rows.Scan(valuePtrs...); err != nil {
			continue
		}

		result := make(map[string]interface{})
		for i, col := range columns {
			val := values[i]
			if val == nil {
				continue
			}
			switch v := val.(type) {
			case []byte:
				result[col] = string(v)
			default:
				result[col] = v
			}
		}
		results = append(results, result)
	}

	return results, total, nil
}

func GetFavoriteContents(mediaID int64, page, size int) ([]map[string]interface{}, int, error) {
	db := GetFavoritesDB()
	if db == nil {
		return []map[string]interface{}{}, 0, nil
	}

	var total int
	err := db.QueryRow("SELECT COUNT(*) FROM favorites_content WHERE media_id = ?", mediaID).Scan(&total)
	if err != nil {
		return []map[string]interface{}{}, 0, err
	}

	offset := (page - 1) * size
	rows, err := db.Query(`
		SELECT * FROM favorites_content
		WHERE media_id = ?
		ORDER BY fav_time DESC
		LIMIT ? OFFSET ?
	`, mediaID, size, offset)
	if err != nil {
		return []map[string]interface{}{}, 0, err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return []map[string]interface{}{}, 0, err
	}

	var results []map[string]interface{}
	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range columns {
			valuePtrs[i] = &values[i]
		}
		if err := rows.Scan(valuePtrs...); err != nil {
			continue
		}

		result := make(map[string]interface{})
		for i, col := range columns {
			val := values[i]
			if val == nil {
				continue
			}
			switch v := val.(type) {
			case []byte:
				result[col] = string(v)
			default:
				result[col] = v
			}
		}
		results = append(results, result)
	}

	return results, total, nil
}
