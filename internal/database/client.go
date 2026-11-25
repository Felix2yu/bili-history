package database

import (
	"database/sql"
	"fmt"
	"sort"
	"strings"
	"time"

	"bilibili-history-go/internal/models"
)

// This file backs the external client channel (/api/client/*). The queries read
// the same per-year history tables and mirror databases the web UI uses; nothing
// here writes, so a client can never corrupt the archive.

// clientVideoColumns is the projection every client list uses, in scan order.
const clientVideoColumns = `bvid, COALESCE(NULLIF(title, ''), long_title, ''), COALESCE(cover, ''),
	COALESCE(author_name, ''), COALESCE(author_mid, 0), COALESCE(duration, 0),
	COALESCE(progress, 0), view_at, COALESCE(business, '')`

// excludedBusinesses never carry a playable position: a live room has no fixed
// duration and an article has no player, so a 继续观看 queue must skip them.
var excludedBusinesses = []string{"live", "article", "article-list"}

func excludedBusinessSQL() string {
	quoted := make([]string, len(excludedBusinesses))
	for i, value := range excludedBusinesses {
		quoted[i] = "'" + value + "'"
	}
	return "COALESCE(business, '') NOT IN (" + strings.Join(quoted, ",") + ")"
}

// scanClientVideos reads rows produced by clientVideoColumns.
func scanClientVideos(rows *sql.Rows) ([]models.ClientVideo, error) {
	defer rows.Close()
	var result []models.ClientVideo
	for rows.Next() {
		var item models.ClientVideo
		if err := rows.Scan(&item.Bvid, &item.Title, &item.Cover, &item.Author, &item.AuthorMid,
			&item.Duration, &item.Progress, &item.ViewAt, &item.Business); err != nil {
			continue
		}
		if item.Bvid == "" {
			continue
		}
		// A finished play is stored as -1 by Bilibili's own sync; normalize it so
		// the client only ever sees a position inside the video.
		item.Finished = item.Progress < 0 || (item.Duration > 0 && item.Progress >= item.Duration)
		if item.Progress < 0 {
			item.Progress = item.Duration
		}
		if item.Progress < 0 {
			item.Progress = 0
		}
		result = append(result, item)
	}
	return result, nil
}

// yearTables returns the existing history tables, newest year first.
func yearTables() ([]string, error) {
	db := GetSQLiteDB()
	if db == nil || db.GetDB() == nil {
		return nil, fmt.Errorf("database not initialized")
	}
	years, err := db.GetAvailableYears()
	if err != nil {
		return nil, err
	}
	// GetAvailableYears orders by table name DESC, which is only correct while
	// every year has four digits; sort explicitly so a five-digit year cannot
	// slip in between.
	sort.Sort(sort.Reverse(sort.IntSlice(years)))
	tables := make([]string, 0, len(years))
	for _, year := range years {
		name := fmt.Sprintf("bilibili_history_%d", year)
		if exists, _ := db.TableExists(name); exists {
			tables = append(tables, name)
		}
	}
	return tables, nil
}

// RecentWatch returns the videos the user started but has not finished, newest
// first. One entry per bvid: bili-history writes a row per viewing session, and
// the queue should offer the video once, at its latest position.
func RecentWatch(limit int) ([]models.ClientVideo, error) {
	if limit <= 0 {
		limit = 30
	}
	if limit > 100 {
		limit = 100
	}
	tables, err := yearTables()
	if err != nil {
		return nil, err
	}

	seen := make(map[string]bool, limit)
	var merged []models.ClientVideo
	for _, table := range tables {
		// Read more than needed per year table: after de-duplication across years
		// and after dropping finished rows, the newest `limit` entries are almost
		// always inside this window, and a too-small window would under-fill.
		query := fmt.Sprintf(`SELECT %s FROM %s
			WHERE progress > 0 AND bvid != '' AND %s
			  AND NOT (progress = -1 OR (duration > 0 AND progress >= duration))
			ORDER BY view_at DESC LIMIT ?`, clientVideoColumns, table, excludedBusinessSQL())
		rows, err := GetSQLiteDB().GetDB().Query(query, limit*3)
		if err != nil {
			continue
		}
		items, err := scanClientVideos(rows)
		if err != nil {
			continue
		}
		for _, item := range items {
			if seen[item.Bvid] || item.Finished {
				continue
			}
			seen[item.Bvid] = true
			merged = append(merged, item)
		}
	}
	sort.SliceStable(merged, func(i, j int) bool { return merged[i].ViewAt > merged[j].ViewAt })
	if len(merged) > limit {
		merged = merged[:limit]
	}
	if merged == nil {
		merged = []models.ClientVideo{}
	}
	return merged, nil
}

// ArchiveHistoryParams describes one cursor page of the archive.
type ArchiveHistoryParams struct {
	// Cursor is the view_at to continue after; 0 starts at the newest row.
	Cursor  int64
	Limit   int
	Keyword string
	// Business limits the list to one kind ("video", "live", …); empty means all.
	Business string
	// Year limits the query to one history table; 0 scans every year.
	Year int
}

// ArchiveHistory returns one page of the permanent history. The cursor is
// view_at rather than an offset: the archive grows at the top every hour, so an
// offset page would repeat or skip rows between requests.
func ArchiveHistory(params ArchiveHistoryParams) (*models.ClientHistoryPage, error) {
	if params.Limit <= 0 {
		params.Limit = 40
	}
	if params.Limit > 100 {
		params.Limit = 100
	}

	var tables []string
	if params.Year > 0 {
		name := fmt.Sprintf("bilibili_history_%d", params.Year)
		if exists, _ := GetSQLiteDB().TableExists(name); !exists {
			return &models.ClientHistoryPage{Records: []models.ClientVideo{}}, nil
		}
		tables = []string{name}
	} else {
		var err error
		if tables, err = yearTables(); err != nil {
			return nil, err
		}
	}

	conn := GetSQLiteDB().GetDB()
	keyword := strings.TrimSpace(params.Keyword)
	var merged []models.ClientVideo
	for _, table := range tables {
		where := []string{"bvid != ''"}
		args := []interface{}{}
		if params.Cursor > 0 {
			where = append(where, "view_at < ?")
			args = append(args, params.Cursor)
		}
		if keyword != "" {
			where = append(where, "(title LIKE ? OR long_title LIKE ? OR author_name LIKE ?)")
			like := "%" + keyword + "%"
			args = append(args, like, like, like)
		}
		if params.Business != "" {
			where = append(where, "COALESCE(business, '') = ?")
			args = append(args, params.Business)
		}
		query := fmt.Sprintf("SELECT %s FROM %s WHERE %s ORDER BY view_at DESC LIMIT ?",
			clientVideoColumns, table, strings.Join(where, " AND "))
		args = append(args, params.Limit)
		rows, err := conn.Query(query, args...)
		if err != nil {
			continue
		}
		items, err := scanClientVideos(rows)
		if err != nil {
			continue
		}
		merged = append(merged, items...)
	}

	sort.SliceStable(merged, func(i, j int) bool { return merged[i].ViewAt > merged[j].ViewAt })
	page := &models.ClientHistoryPage{Records: []models.ClientVideo{}}
	if len(merged) == 0 {
		return page, nil
	}
	// Many viewing sessions of the same video are legitimate history rows; the
	// archive lists them as they happened, so only the page length is trimmed.
	if len(merged) > params.Limit {
		merged = merged[:params.Limit]
	}
	page.Records = merged
	if len(merged) == params.Limit {
		page.Next = merged[len(merged)-1].ViewAt
	}
	return page, nil
}

// WatchSummary summarizes one month (or a whole year when month is 0). The
// definitions match computeSummary in report.go: 深夜 is 22:00–06:00 and the
// duration is the videos' own length, so a summary and the web report agree.
func WatchSummary(year, month int) (*models.ClientSummary, error) {
	if year <= 0 {
		year = time.Now().Year()
	}
	result := &models.ClientSummary{Year: year, Month: month}
	name := fmt.Sprintf("bilibili_history_%d", year)
	if exists, _ := GetSQLiteDB().TableExists(name); !exists {
		return result, nil
	}

	start := time.Date(year, time.January, 1, 0, 0, 0, 0, time.Local)
	end := time.Date(year+1, time.January, 1, 0, 0, 0, 0, time.Local)
	if month >= 1 && month <= 12 {
		start = time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.Local)
		end = start.AddDate(0, 1, 0)
	}

	var finished, lateNight sql.NullInt64
	err := GetSQLiteDB().GetDB().QueryRow(fmt.Sprintf(`SELECT
			COUNT(*),
			COALESCE(SUM(COALESCE(duration, 0)), 0),
			COUNT(DISTINCT author_mid),
			COUNT(DISTINCT dt),
			COALESCE(SUM(CASE WHEN is_finish = 1 OR progress = -1 OR (duration > 0 AND progress >= duration) THEN 1 ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN CAST(strftime('%%H', view_at, 'unixepoch', 'localtime') AS INTEGER) >= 22
				OR CAST(strftime('%%H', view_at, 'unixepoch', 'localtime') AS INTEGER) < 6 THEN 1 ELSE 0 END), 0)
		FROM %s WHERE view_at >= ? AND view_at < ? AND bvid != ''`, name),
		start.Unix(), end.Unix()).Scan(&result.Videos, &result.Seconds, &result.UniqueAuthors,
		&result.UniqueDays, &finished, &lateNight)
	if err != nil {
		return result, err
	}
	result.Finished = int(finished.Int64)
	if result.Videos > 0 {
		result.FinishedRate = float64(result.Finished) / float64(result.Videos)
		result.LateNightRatio = float64(lateNight.Int64) / float64(result.Videos)
	}
	return result, nil
}

// WatchLaterPage reads the archived 稍后再看 mirror. A deployment that never
// synced it has no database file, which is an empty list rather than an error.
func WatchLaterPage(page, size int) ([]models.ClientWatchLaterItem, int, error) {
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 40
	}
	if size > 100 {
		size = 100
	}
	conn := GetWatchLaterDB()
	items := []models.ClientWatchLaterItem{}
	if conn == nil {
		return items, 0, nil
	}
	var total int
	if err := conn.QueryRow("SELECT COUNT(*) FROM watchlater_videos").Scan(&total); err != nil {
		return items, 0, err
	}
	rows, err := conn.Query(`SELECT bvid, COALESCE(title, ''), COALESCE(pic, ''), COALESCE(owner_name, ''),
			COALESCE(duration, 0), COALESCE(add_at, 0), COALESCE(view, 0)
		FROM watchlater_videos ORDER BY add_at DESC LIMIT ? OFFSET ?`, size, (page-1)*size)
	if err != nil {
		return items, 0, err
	}
	defer rows.Close()
	for rows.Next() {
		var item models.ClientWatchLaterItem
		if rows.Scan(&item.Bvid, &item.Title, &item.Cover, &item.Author, &item.Duration, &item.AddAt, &item.View) == nil {
			items = append(items, item)
		}
	}
	return items, total, nil
}

// FavoriteFolders reads the archived 收藏夹 list.
func FavoriteFolders() ([]models.ClientFavoriteFolder, error) {
	folders := []models.ClientFavoriteFolder{}
	conn := GetFavoritesDB()
	if conn == nil {
		return folders, nil
	}
	rows, err := conn.Query(`SELECT media_id, COALESCE(title, ''), COALESCE(cover, ''),
			COALESCE(media_count, 0), COALESCE(mtime, 0)
		FROM favorites_folder ORDER BY mtime DESC`)
	if err != nil {
		return folders, err
	}
	defer rows.Close()
	for rows.Next() {
		var folder models.ClientFavoriteFolder
		if rows.Scan(&folder.MediaID, &folder.Title, &folder.Cover, &folder.MediaCount, &folder.Mtime) == nil {
			folders = append(folders, folder)
		}
	}
	return folders, nil
}

// FavoriteContents reads one page of an archived 收藏夹. Rows whose bvid was
// only ever stored in bv_id (older syncs) still get an id, so the client can
// look them up like any other video.
func FavoriteContents(mediaID int64, page, size int) ([]models.ClientFavoriteItem, int, error) {
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 40
	}
	if size > 100 {
		size = 100
	}
	items := []models.ClientFavoriteItem{}
	conn := GetFavoritesDB()
	if conn == nil {
		return items, 0, nil
	}
	var total int
	if err := conn.QueryRow("SELECT COUNT(*) FROM favorites_content WHERE media_id = ?", mediaID).Scan(&total); err != nil {
		return items, 0, err
	}
	rows, err := conn.Query(`SELECT COALESCE(NULLIF(bvid, ''), COALESCE(bv_id, '')), COALESCE(title, ''),
			COALESCE(cover, ''), COALESCE(creator_name, ''), COALESCE(upper_mid, 0),
			COALESCE(duration, 0), COALESCE(fav_time, 0)
		FROM favorites_content WHERE media_id = ? ORDER BY fav_time DESC LIMIT ? OFFSET ?`,
		mediaID, size, (page-1)*size)
	if err != nil {
		return items, 0, err
	}
	defer rows.Close()
	for rows.Next() {
		var item models.ClientFavoriteItem
		if rows.Scan(&item.Bvid, &item.Title, &item.Cover, &item.Author, &item.UpperMid,
			&item.Duration, &item.FavTime) == nil {
			items = append(items, item)
		}
	}
	return items, total, nil
}
