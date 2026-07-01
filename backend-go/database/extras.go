package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"bilibili-history-go/utils"

	_ "github.com/mattn/go-sqlite3"
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

func GetLikesDB() *sql.DB {
	likesOnce.Do(func() {
		likesDB = &ExtraDB{
			db: getExtraDB("bilibili_likes.db"),
		}
	})
	if likesDB == nil {
		return nil
	}
	return likesDB.db
}

func GetWatchLaterDB() *sql.DB {
	watchlaterOnce.Do(func() {
		watchlaterDB = &ExtraDB{
			db: getExtraDB("bilibili_watchlater.db"),
		}
	})
	if watchlaterDB == nil {
		return nil
	}
	return watchlaterDB.db
}

func GetFavoritesDB() *sql.DB {
	favoritesOnce.Do(func() {
		favoritesDB = &ExtraDB{
			db: getExtraDB("bilibili_favorites.db"),
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
