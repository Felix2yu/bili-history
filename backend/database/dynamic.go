package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"bilibili-history-go/utils"
)

type DynamicHost struct {
	HostMid       string `json:"host_mid"`
	UpName        string `json:"up_name"`
	FacePath      string `json:"face_path"`
	ItemCount     int    `json:"item_count"`
	CoreCount     int    `json:"core_count"`
	LastPublishTS int64  `json:"last_publish_ts"`
	LastFetchTime int64  `json:"last_fetch_time"`
	LastDynamicID string `json:"last_dynamic_id"`
}

type DynamicItem struct {
	ID              string   `json:"id_str"`
	Type            string   `json:"type"`
	HostMid         string   `json:"host_mid"`
	AuthorName      string   `json:"author_name"`
	AuthorFace      string   `json:"author_face"`
	Txt             string   `json:"txt"`
	OpusTitle       string   `json:"opus_title"`
	OpusSummaryText string   `json:"opus_summary_text"`
	Bvid            string   `json:"bvid"`
	Title           string   `json:"title"`
	Desc            string   `json:"desc"`
	Cover           string   `json:"cover"`
	PublishTS       int64    `json:"publish_ts"`
	MediaLocals     []string `json:"media_locals"`
	LiveMediaLocals []string `json:"live_media_locals"`
	RawJSON         string   `json:"-"`
	FetchTime       int64    `json:"-"`
}

var (
	dynamicDB   *ExtraDB
	dynamicOnce sync.Once
)

const dynamicSchema = `
CREATE TABLE IF NOT EXISTS dynamic_hosts (
    host_mid TEXT PRIMARY KEY,
    up_name TEXT DEFAULT '',
    face_path TEXT DEFAULT '',
    item_count INTEGER DEFAULT 0,
    core_count INTEGER DEFAULT 0,
    last_publish_ts INTEGER DEFAULT 0,
    last_fetch_time INTEGER DEFAULT 0,
    last_dynamic_id TEXT DEFAULT ''
);
CREATE TABLE IF NOT EXISTS dynamics (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    id_str TEXT NOT NULL UNIQUE,
    type TEXT DEFAULT '',
    host_mid TEXT NOT NULL,
    author_name TEXT DEFAULT '',
    author_face TEXT DEFAULT '',
    txt TEXT DEFAULT '',
    opus_title TEXT DEFAULT '',
    opus_summary_text TEXT DEFAULT '',
    bvid TEXT DEFAULT '',
    title TEXT DEFAULT '',
    desc TEXT DEFAULT '',
    cover TEXT DEFAULT '',
    publish_ts INTEGER DEFAULT 0,
    media_locals TEXT DEFAULT '[]',
    live_media_locals TEXT DEFAULT '[]',
    raw_json TEXT DEFAULT '',
    fetch_time INTEGER DEFAULT 0
);
CREATE INDEX IF NOT EXISTS idx_dynamics_host_mid ON dynamics(host_mid);
CREATE INDEX IF NOT EXISTS idx_dynamics_publish_ts ON dynamics(publish_ts);
CREATE INDEX IF NOT EXISTS idx_dynamics_bvid ON dynamics(bvid);
`

func GetDynamicDB() *sql.DB {
	dynamicOnce.Do(func() {
		db := getExtraDB("bilibili_dynamic.db")
		if db != nil {
			if _, err := db.Exec(dynamicSchema); err != nil {
				utils.LogError("Failed to ensure dynamic schema: %v", err)
			}
			// 迁移：添加 author_face 列
			db.Exec("ALTER TABLE dynamics ADD COLUMN author_face TEXT DEFAULT ''")
			// 迁移：添加 last_dynamic_id 列
			db.Exec("ALTER TABLE dynamic_hosts ADD COLUMN last_dynamic_id TEXT DEFAULT ''")
		}
		dynamicDB = &ExtraDB{db: db}
	})
	if dynamicDB == nil {
		return nil
	}
	return dynamicDB.db
}

func GetDynamicHosts(limit, offset int) ([]DynamicHost, error) {
	db := GetDynamicDB()
	if db == nil {
		return nil, fmt.Errorf("dynamic database not initialized")
	}

	if limit <= 0 || limit > 200 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	rows, err := db.Query(`
		SELECT host_mid, up_name, face_path, item_count, core_count, last_publish_ts, last_fetch_time
		FROM dynamic_hosts
		ORDER BY last_fetch_time DESC
		LIMIT ? OFFSET ?
	`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var hosts []DynamicHost
	for rows.Next() {
		var h DynamicHost
		if err := rows.Scan(&h.HostMid, &h.UpName, &h.FacePath, &h.ItemCount, &h.CoreCount, &h.LastPublishTS, &h.LastFetchTime); err != nil {
			continue
		}
		hosts = append(hosts, h)
	}
	return hosts, nil
}

func GetDynamicSpace(hostMid string, limit, offset int) (int, []DynamicItem, error) {
	db := GetDynamicDB()
	if db == nil {
		return 0, nil, fmt.Errorf("dynamic database not initialized")
	}

	if limit <= 0 || limit > 200 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	var total int
	err := db.QueryRow("SELECT COUNT(*) FROM dynamics WHERE host_mid = ?", hostMid).Scan(&total)
	if err != nil {
		return 0, nil, err
	}

	rows, err := db.Query(`
		SELECT id_str, type, host_mid, author_name, author_face, txt, opus_title, opus_summary_text,
		       bvid, title, desc, cover, publish_ts, media_locals, live_media_locals
		FROM dynamics
		WHERE host_mid = ?
		ORDER BY publish_ts DESC
		LIMIT ? OFFSET ?
	`, hostMid, limit, offset)
	if err != nil {
		return 0, nil, err
	}
	defer rows.Close()

	var items []DynamicItem
	for rows.Next() {
		var item DynamicItem
		var mediaLocalsJSON, liveMediaLocalsJSON string
		if err := rows.Scan(&item.ID, &item.Type, &item.HostMid, &item.AuthorName, &item.AuthorFace, &item.Txt,
			&item.OpusTitle, &item.OpusSummaryText, &item.Bvid, &item.Title, &item.Desc,
			&item.Cover, &item.PublishTS, &mediaLocalsJSON, &liveMediaLocalsJSON); err != nil {
			continue
		}
		json.Unmarshal([]byte(mediaLocalsJSON), &item.MediaLocals)
		json.Unmarshal([]byte(liveMediaLocalsJSON), &item.LiveMediaLocals)
		items = append(items, item)
	}

	return total, items, nil
}

func SaveDynamicHost(host DynamicHost) error {
	db := GetDynamicDB()
	if db == nil {
		return fmt.Errorf("dynamic database not initialized")
	}

	_, err := db.Exec(`
		INSERT INTO dynamic_hosts (host_mid, up_name, face_path, item_count, core_count, last_publish_ts, last_fetch_time, last_dynamic_id)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(host_mid) DO UPDATE SET
			up_name = excluded.up_name,
			face_path = CASE WHEN excluded.face_path != '' THEN excluded.face_path ELSE dynamic_hosts.face_path END,
			item_count = excluded.item_count,
			core_count = excluded.core_count,
			last_publish_ts = excluded.last_publish_ts,
			last_fetch_time = excluded.last_fetch_time,
			last_dynamic_id = CASE WHEN excluded.last_dynamic_id != '' THEN excluded.last_dynamic_id ELSE dynamic_hosts.last_dynamic_id END
	`, host.HostMid, host.UpName, host.FacePath, host.ItemCount, host.CoreCount, host.LastPublishTS, host.LastFetchTime, host.LastDynamicID)
	return err
}

// GetLatestDynamicID 获取指定用户最新的动态ID
func GetLatestDynamicID(hostMid string) string {
	db := GetDynamicDB()
	if db == nil {
		return ""
	}

	var idStr string
	err := db.QueryRow("SELECT id_str FROM dynamics WHERE host_mid = ? ORDER BY publish_ts DESC LIMIT 1", hostMid).Scan(&idStr)
	if err != nil {
		return ""
	}
	return idStr
}

// IsDynamicExists 检查动态是否已存在
func IsDynamicExists(idStr string) bool {
	db := GetDynamicDB()
	if db == nil {
		return false
	}

	var exists bool
	db.QueryRow("SELECT EXISTS(SELECT 1 FROM dynamics WHERE id_str = ?)", idStr).Scan(&exists)
	return exists
}

// GetDynamicCount 获取指定用户的动态总数
func GetDynamicCount(hostMid string) (int, error) {
	db := GetDynamicDB()
	if db == nil {
		return 0, fmt.Errorf("dynamic database not initialized")
	}

	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM dynamics WHERE host_mid = ?", hostMid).Scan(&count)
	return count, err
}

// GetDynamicByID 根据ID获取动态详情
func GetDynamicByID(idStr string) (*DynamicItem, error) {
	db := GetDynamicDB()
	if db == nil {
		return nil, fmt.Errorf("dynamic database not initialized")
	}

	var item DynamicItem
	var mediaLocalsJSON, liveMediaLocalsJSON string
	err := db.QueryRow(`
		SELECT id_str, type, host_mid, author_name, author_face, txt, opus_title, opus_summary_text,
		       bvid, title, desc, cover, publish_ts, media_locals, live_media_locals
		FROM dynamics WHERE id_str = ?
	`, idStr).Scan(&item.ID, &item.Type, &item.HostMid, &item.AuthorName, &item.AuthorFace, &item.Txt,
		&item.OpusTitle, &item.OpusSummaryText, &item.Bvid, &item.Title, &item.Desc,
		&item.Cover, &item.PublishTS, &mediaLocalsJSON, &liveMediaLocalsJSON)
	if err != nil {
		return nil, err
	}
	json.Unmarshal([]byte(mediaLocalsJSON), &item.MediaLocals)
	json.Unmarshal([]byte(liveMediaLocalsJSON), &item.LiveMediaLocals)
	return &item, nil
}

// DeleteDynamic 删除指定动态
func DeleteDynamic(idStr string) error {
	db := GetDynamicDB()
	if db == nil {
		return fmt.Errorf("dynamic database not initialized")
	}

	_, err := db.Exec("DELETE FROM dynamics WHERE id_str = ?", idStr)
	return err
}

// UpdateDynamicHostStats 更新UP主统计信息
func UpdateDynamicHostStats(hostMid string) error {
	db := GetDynamicDB()
	if db == nil {
		return fmt.Errorf("dynamic database not initialized")
	}

	var totalCount int
	db.QueryRow("SELECT COUNT(*) FROM dynamics WHERE host_mid = ?", hostMid).Scan(&totalCount)

	var maxPublishTS int64
	db.QueryRow("SELECT COALESCE(MAX(publish_ts), 0) FROM dynamics WHERE host_mid = ?", hostMid).Scan(&maxPublishTS)

	coreCount := 0
	db.QueryRow("SELECT COUNT(*) FROM dynamics WHERE host_mid = ? AND (type = 'DYNAMIC_TYPE_AV' OR type = 'DYNAMIC_TYPE_DRAW')", hostMid).Scan(&coreCount)

	_, err := db.Exec(`
		UPDATE dynamic_hosts SET item_count = ?, core_count = ?, last_publish_ts = ?
		WHERE host_mid = ?
	`, totalCount, coreCount, maxPublishTS, hostMid)
	return err
}

func SaveDynamics(hostMid string, items []DynamicItem) (int, error) {
	db := GetDynamicDB()
	if db == nil {
		return 0, fmt.Errorf("dynamic database not initialized")
	}

	now := time.Now().Unix()
	inserted := 0

	for _, item := range items {
		mediaJSON, _ := json.Marshal(item.MediaLocals)
		liveMediaJSON, _ := json.Marshal(item.LiveMediaLocals)

		result, err := db.Exec(`
			INSERT INTO dynamics (id_str, type, host_mid, author_name, author_face, txt, opus_title, opus_summary_text,
			                      bvid, title, desc, cover, publish_ts, media_locals, live_media_locals, raw_json, fetch_time)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
			ON CONFLICT(id_str) DO UPDATE SET
				txt = excluded.txt,
				author_face = CASE WHEN excluded.author_face != '' THEN excluded.author_face ELSE dynamics.author_face END,
				media_locals = excluded.media_locals,
				live_media_locals = excluded.live_media_locals
		`, item.ID, item.Type, hostMid, item.AuthorName, item.AuthorFace, item.Txt, item.OpusTitle, item.OpusSummaryText,
			item.Bvid, item.Title, item.Desc, item.Cover, item.PublishTS,
			string(mediaJSON), string(liveMediaJSON), item.RawJSON, now)
		if err != nil {
			utils.LogWarning("保存动态失败 %s: %v", item.ID, err)
			continue
		}
		aff, _ := result.RowsAffected()
		if aff > 0 {
			inserted++
		}
	}

	// Update host stats
	if len(items) > 0 {
		var maxPublishTS int64
		coreCount := 0
		for _, item := range items {
			if item.PublishTS > maxPublishTS {
				maxPublishTS = item.PublishTS
			}
			if item.Type == "DYNAMIC_TYPE_AV" || item.Type == "DYNAMIC_TYPE_DRAW" {
				coreCount++
			}
		}
		var totalCount int
		db.QueryRow("SELECT COUNT(*) FROM dynamics WHERE host_mid = ?", hostMid).Scan(&totalCount)

		var upName, facePath string
		db.QueryRow("SELECT up_name, face_path FROM dynamic_hosts WHERE host_mid = ?", hostMid).Scan(&upName, &facePath)
		if len(items) > 0 && items[0].AuthorName != "" {
			upName = items[0].AuthorName
		}

		SaveDynamicHost(DynamicHost{
			HostMid:       hostMid,
			UpName:        upName,
			FacePath:      facePath,
			ItemCount:     totalCount,
			CoreCount:     coreCount,
			LastPublishTS: maxPublishTS,
			LastFetchTime: now,
		})
	}

	return inserted, nil
}

func DeleteDynamicSpace(hostMid string) error {
	db := GetDynamicDB()
	if db == nil {
		return fmt.Errorf("dynamic database not initialized")
	}

	_, err := db.Exec("DELETE FROM dynamics WHERE host_mid = ?", hostMid)
	if err != nil {
		return err
	}

	_, err = db.Exec("DELETE FROM dynamic_hosts WHERE host_mid = ?", hostMid)
	return err
}
