package database

import (
	"database/sql"
	"fmt"
	"strings"

	"bilibili-history-go/internal/models"
	"bilibili-history-go/internal/utils"
)

// BatchGetWatchProgress returns the most recent viewing state for each requested
// bvid. A row in any year table counts as watched; progress == -1 marks a
// completed play and is normalized to the recorded duration.
func BatchGetWatchProgress(bvids []string) (map[string]models.WatchProgress, error) {
	result := make(map[string]models.WatchProgress, len(bvids))
	if len(bvids) == 0 {
		return result, nil
	}

	db := GetSQLiteDB()
	conn := db.GetDB()
	if conn == nil {
		return nil, fmt.Errorf("database not initialized")
	}
	years, err := db.GetAvailableYears()
	if err != nil {
		return nil, err
	}

	placeholders := make([]string, len(bvids))
	args := make([]interface{}, len(bvids))
	for i, bvid := range bvids {
		placeholders[i] = "?"
		args[i] = bvid
	}
	inClause := strings.Join(placeholders, ",")

	for _, year := range years {
		tableName := fmt.Sprintf("bilibili_history_%d", year)
		exists, _ := db.TableExists(tableName)
		if !exists {
			continue
		}
		rows, err := conn.Query(fmt.Sprintf(`
			SELECT bvid, progress, duration, view_at FROM %s
			WHERE bvid IN (%s) AND COALESCE(business, '') NOT IN ('live', 'article', 'article-list')
			ORDER BY view_at DESC
		`, tableName, inClause), args...)
		if err != nil {
			continue
		}
		for rows.Next() {
			var bvid string
			var progress, duration int
			var viewAt int64
			if rows.Scan(&bvid, &progress, &duration, &viewAt) != nil {
				continue
			}
			if _, seen := result[bvid]; seen {
				continue // first row is the newest across all year tables
			}
			finished := progress == -1 || (duration > 0 && progress >= duration)
			if progress < 0 {
				progress = duration
			}
			result[bvid] = models.WatchProgress{
				Bvid: bvid, Watched: true, Progress: progress, Duration: duration,
				Finished: finished, ViewAt: viewAt,
			}
		}
		rows.Close()
	}
	return result, nil
}

// UpsertWatchProgress writes one external playback report into the year table
// matching its timestamp, reusing the same 24 h session merge as Bilibili sync.
func UpsertWatchProgress(record *models.HistoryRecord) error {
	db := GetSQLiteDB()
	conn := db.GetDB()
	if conn == nil {
		return fmt.Errorf("database not initialized")
	}
	year := utils.GetYearFromTimestamp(record.ViewAt)
	if err := db.EnsureTableForYear(year); err != nil {
		return err
	}
	tableName := fmt.Sprintf("bilibili_history_%d", year)

	// Two reports can land in the same second. InsertHistoryRecord only moves a
	// row forward when view_at is strictly newer, so handle an existing row here
	// and allow an equal timestamp to replace the progress.
	var existingID int64
	var existingViewAt int64
	err := conn.QueryRow(fmt.Sprintf(
		"SELECT id, view_at FROM %s WHERE bvid = ? ORDER BY view_at DESC LIMIT 1", tableName),
		record.Bvid).Scan(&existingID, &existingViewAt)
	if err == nil {
		// Device clocks differ by a few minutes. Reject only a clearly stale report, and never
		// rewind view_at even when the progress itself is newer and should still be stored.
		if record.ViewAt < existingViewAt-24*3600 {
			return nil
		}
		_, err = conn.Exec(fmt.Sprintf(
			`UPDATE %s SET view_at = MAX(view_at, ?), progress = ?, current = ?, duration = ?,
			cover = CASE WHEN ? = '' THEN cover ELSE ? END WHERE id = ?`,
			tableName), record.ViewAt, record.Progress, record.Current, record.Duration,
			record.Cover, record.Cover, existingID)
		return err
	}
	if err != sql.ErrNoRows {
		return err
	}
	_, err = InsertHistoryRecord(conn, tableName, record)
	return err
}
