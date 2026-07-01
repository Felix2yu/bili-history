package services

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"bili-history/internal/config"
	"bili-history/internal/db"
	"bili-history/internal/models"
	"bili-history/internal/services/biliapi"
)

// HistoryFetcher handles fetching and storing bilibili watch history.
type HistoryFetcher struct {
	client *biliapi.Client
	cfg    *config.Config
}

// NewHistoryFetcher creates a new history fetcher.
func NewHistoryFetcher() (*HistoryFetcher, error) {
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, err
	}
	return &HistoryFetcher{
		client: biliapi.NewClient(cfg.SESSDATA, cfg.BiliJCT, cfg.DedeUserID),
		cfg:    cfg,
	}, nil
}

// FetchAndStore fetches history from Bilibili and stores it locally.
func (f *HistoryFetcher) FetchAndStore(onProgress func(page int, count int)) ([]models.HistoryRecord, error) {
	var allEntries []models.HistoryRecord
	page := 0

	for {
		entries, cursor, err := f.client.FetchHistory(20, 0)
		if err != nil {
			return allEntries, fmt.Errorf("fetch history failed: %w", err)
		}
		if len(entries) == 0 {
			break
		}

		page++
		for _, entry := range entries {
			record := models.HistoryRecord{
				BVID:      entry.BVID,
				AID:       0,
				Title:     entry.Title,
				Pic:       entry.Pic,
				Desc:      "",
				Duration:  entry.Duration,
				OwnerName: entry.Owner.Name,
				OwnerMid:  entry.Owner.Mid,
				TagName:   "",
				Tid:       0,
				ViewAt:    entry.ViewAt,
				Progress:  entry.Progress,
				Business:  "",
				View:      0,
				Danmaku:   0,
				Coin:      0,
				Favorite:  0,
				Like:      0,
				Reply:     0,
				Share:     0,
				// Backward-compat aliases
				Thumbnail: entry.Pic,
				OwnerID:   entry.Owner.Mid,
				Cid:       entry.Cid,
			}
			allEntries = append(allEntries, record)
		}

		if onProgress != nil {
			onProgress(page, len(allEntries))
		}

		if cursor == nil || cursor.IsEnd {
			break
		}

		// Fetch next page
		entries2, cursor2, err := f.client.FetchHistory(20, cursor.Earliest)
		if err != nil {
			break
		}
		if len(entries2) == 0 {
			break
		}

		page++
		for _, entry := range entries2 {
			record := models.HistoryRecord{
				BVID:      entry.BVID,
				Title:     entry.Title,
				Pic:       entry.Pic,
				Duration:  entry.Duration,
				OwnerName: entry.Owner.Name,
				OwnerMid:  entry.Owner.Mid,
				ViewAt:    entry.ViewAt,
				Progress:  entry.Progress,
				Thumbnail: entry.Pic,
				OwnerID:   entry.Owner.Mid,
				Cid:       entry.Cid,
			}
			allEntries = append(allEntries, record)
		}

		if onProgress != nil {
			onProgress(page, len(allEntries))
		}

		if cursor2 == nil || cursor2.IsEnd {
			break
		}
	}

	// Save to JSON files
	if err := f.saveHistoryToFiles(allEntries); err != nil {
		return allEntries, fmt.Errorf("save history failed: %w", err)
	}

	// Import into SQLite
	if err := f.importToDatabase(allEntries); err != nil {
		return allEntries, fmt.Errorf("import to db failed: %w", err)
	}

	return allEntries, nil
}

// saveHistoryToFiles saves history entries to JSON files organized by date.
func (f *HistoryFetcher) saveHistoryToFiles(entries []models.HistoryRecord) error {
	basePath := config.GetOutputPath("history_by_date")

	type dayKey struct {
		year  int
		month int
		day   int
	}
	grouped := make(map[dayKey][]models.HistoryRecord)

	for _, entry := range entries {
		t := time.Unix(entry.ViewAt, 0)
		key := dayKey{year: t.Year(), month: int(t.Month()), day: t.Day()}
		grouped[key] = append(grouped[key], entry)
	}

	for key, dayEntries := range grouped {
		dir := filepath.Join(basePath,
			fmt.Sprintf("%d", key.year),
			fmt.Sprintf("%02d", key.month))
		os.MkdirAll(dir, 0755)

		filePath := filepath.Join(dir, fmt.Sprintf("%02d.json", key.day))

		// Read existing data
		var existing []map[string]interface{}
		if data, err := os.ReadFile(filePath); err == nil {
			json.Unmarshal(data, &existing)
		}

		// Deduplicate by bvid+view_at
		existingSet := make(map[string]bool)
		for _, item := range existing {
			bvid, _ := item["bvid"].(string)
			viewAt, _ := item["view_at"].(float64)
			existingSet[fmt.Sprintf("%s_%d", bvid, int64(viewAt))] = true
		}

		for _, entry := range dayEntries {
			key := fmt.Sprintf("%s_%d", entry.BVID, entry.ViewAt)
			if !existingSet[key] {
				existing = append(existing, map[string]interface{}{
					"bvid":       entry.BVID,
					"title":      entry.Title,
					"pic":        entry.Pic,
					"duration":   entry.Duration,
					"owner_name": entry.OwnerName,
					"owner_mid":  entry.OwnerMid,
					"view_at":    entry.ViewAt,
					"progress":   entry.Progress,
				})
			}
		}

		data, _ := json.MarshalIndent(existing, "", "  ")
		if err := os.WriteFile(filePath, data, 0644); err != nil {
			return fmt.Errorf("write file %s: %w", filePath, err)
		}
	}

	return nil
}

// importToDatabase imports history entries into SQLite using the Python-compatible schema.
func (f *HistoryFetcher) importToDatabase(entries []models.HistoryRecord) error {
	database, err := db.OpenHistoryDB()
	if err != nil {
		return err
	}

	// Group entries by year
	yearGroups := make(map[int][]models.HistoryRecord)
	for _, entry := range entries {
		t := time.Unix(entry.ViewAt, 0)
		year := t.Year()
		yearGroups[year] = append(yearGroups[year], entry)
	}

	for year, yearEntries := range yearGroups {
		tableName := db.GetYearTable(year)

		// Create table with EXACT Python schema
		if err := db.EnsureHistoryTable(database, year); err != nil {
			return err
		}

		// Insert/upsert entries using INSERT OR IGNORE to avoid primary key conflicts
		stmt, err := database.Prepare(fmt.Sprintf(`INSERT OR IGNORE INTO %s 
			(bvid, aid, title, "desc", pic, duration, owner_name, owner_mid, tag_name, tid,
			 view_at, progress, business, view, danmaku, coin, favorite, like, reply, share)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, tableName))
		if err != nil {
			return fmt.Errorf("prepare statement: %w", err)
		}
		defer stmt.Close()

		for _, entry := range yearEntries {
			if _, err := stmt.Exec(
				entry.BVID, entry.AID, entry.Title, entry.Desc, entry.Pic,
				entry.Duration, entry.OwnerName, entry.OwnerMid, entry.TagName, entry.Tid,
				entry.ViewAt, entry.Progress, entry.Business, entry.View, entry.Danmaku,
				entry.Coin, entry.Favorite, entry.Like, entry.Reply, entry.Share,
			); err != nil {
				log.Printf("Warning: insert failed for %s: %v", entry.BVID, err)
			}
		}
	}

	return nil
}

// ImportFromJSONFiles imports history from local JSON files into the database.
func ImportFromJSONFiles() error {
	basePath := config.GetOutputPath("history_by_date")
	var allEntries []models.HistoryRecord

	err := filepath.Walk(basePath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || filepath.Ext(path) != ".json" {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		var records []models.HistoryRecord
		if err := json.Unmarshal(data, &records); err != nil {
			// Try generic format (from Python's output)
			var generic []map[string]interface{}
			if err := json.Unmarshal(data, &generic); err != nil {
				return nil
			}
			for _, g := range generic {
				r := parseGenericRecord(g)
				if r.BVID != "" {
					records = append(records, r)
				}
			}
		}
		allEntries = append(allEntries, records...)
		return nil
	})

	if err != nil {
		return err
	}

	database, err := db.OpenHistoryDB()
	if err != nil {
		return err
	}

	// Group by year and insert
	yearGroups := make(map[int][]models.HistoryRecord)
	for _, entry := range allEntries {
		t := time.Unix(entry.ViewAt, 0)
		year := t.Year()
		yearGroups[year] = append(yearGroups[year], entry)
	}

	for year, yearEntries := range yearGroups {
		db.EnsureHistoryTable(database, year)
		db.MigrateIfNeeded(database, year)

		tableName := db.GetYearTable(year)
		stmt, err := database.Prepare(fmt.Sprintf(`INSERT OR IGNORE INTO %s 
			(bvid, aid, title, "desc", pic, duration, owner_name, owner_mid, tag_name, tid,
			 view_at, progress, business, view, danmaku, coin, favorite, like, reply, share)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, tableName))
		if err != nil {
			continue
		}

		for _, entry := range yearEntries {
			stmt.Exec(
				entry.BVID, entry.AID, entry.Title, entry.Desc, entry.Pic,
				entry.Duration, entry.OwnerName, entry.OwnerMid, entry.TagName, entry.Tid,
				entry.ViewAt, entry.Progress, entry.Business, entry.View, entry.Danmaku,
				entry.Coin, entry.Favorite, entry.Like, entry.Reply, entry.Share,
			)
		}
		stmt.Close()
	}

	return nil
}

func parseGenericRecord(g map[string]interface{}) models.HistoryRecord {
	r := models.HistoryRecord{}

	// Handle nested "history" object from Python JSON format
	if history, ok := g["history"].(map[string]interface{}); ok {
		if v, ok := history["bvid"].(string); ok {
			r.BVID = v
		}
	} else if v, ok := g["bvid"].(string); ok {
		r.BVID = v
	}

	if v, ok := g["title"].(string); ok {
		r.Title = v
	} else if history, ok := g["history"].(map[string]interface{}); ok {
		if v, ok := history["title"].(string); ok {
			r.Title = v
		}
	}

	if v, ok := g["pic"].(string); ok {
		r.Pic = v
		r.Thumbnail = v
	}

	if v, ok := g["view_at"].(float64); ok {
		r.ViewAt = int64(v)
	}
	if v, ok := g["duration"].(float64); ok {
		r.Duration = int(v)
	}
	if v, ok := g["progress"].(float64); ok {
		r.Progress = int(v)
	}
	if v, ok := g["owner_name"].(string); ok {
		r.OwnerName = v
	} else if history, ok := g["history"].(map[string]interface{}); ok {
		if owner, ok := history["owner"].(map[string]interface{}); ok {
			if v, ok := owner["name"].(string); ok {
				r.OwnerName = v
			}
		}
	}
	if v, ok := g["owner_mid"].(float64); ok {
		r.OwnerMid = int64(v)
		r.OwnerID = int64(v)
	}

	return r
}

// QueryHistoryAll queries history from the database with cross-year support.
func QueryHistoryAll(page, size int, sortOrder int, tagName, mainCategory, dateRange, business string) ([]map[string]interface{}, int, []int, error) {
	database, err := db.OpenHistoryDB()
	if err != nil {
		return nil, 0, nil, err
	}

	availableYears, err := GetAvailableYears()
	if err != nil {
		return nil, 0, nil, err
	}
	if len(availableYears) == 0 {
		return []map[string]interface{}{}, 0, []int{}, nil
	}

	var queries []string
	var params []interface{}

	var startTimestamp, endTimestamp int64
	if dateRange != "" {
		parts := strings.Split(dateRange, "-")
		if len(parts) == 2 {
			startTime, err1 := time.Parse("20060102", parts[0])
			endTime, err2 := time.Parse("20060102", parts[1])
			if err1 == nil && err2 == nil {
				startTimestamp = startTime.Unix()
				endTimestamp = endTime.Unix() + 86400
			}
		}
	}

	for _, year := range availableYears {
		tableName := db.GetYearTable(year)
		if !tableExists(database, tableName) {
			continue
		}
		query := fmt.Sprintf(`SELECT * FROM %s WHERE 1=1`, tableName)

		if startTimestamp > 0 && endTimestamp > 0 {
			query += " AND view_at >= ? AND view_at < ?"
			params = append(params, startTimestamp, endTimestamp)
		}

		if mainCategory != "" {
			query += " AND main_category = ?"
			params = append(params, mainCategory)
		} else if tagName != "" {
			query += " AND tag_name = ?"
			params = append(params, tagName)
		}

		if business != "" {
			query += " AND business = ?"
			params = append(params, business)
		}

		queries = append(queries, query)
	}

	if len(queries) == 0 {
		return []map[string]interface{}{}, 0, availableYears, nil
	}

	baseQuery := strings.Join(queries, " UNION ALL ")

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM (%s)", baseQuery)
	var total int
	countParams := make([]interface{}, len(params))
	copy(countParams, params)
	database.QueryRow(countQuery, countParams...).Scan(&total)

	orderStr := "DESC"
	if sortOrder == 1 {
		orderStr = "ASC"
	}
	finalQuery := fmt.Sprintf(`SELECT * FROM (%s) ORDER BY view_at %s LIMIT ? OFFSET ?`, baseQuery, orderStr)
	params = append(params, size, (page-1)*size)

	rows, err := database.Query(finalQuery, params...)
	if err != nil {
		return nil, 0, nil, err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, 0, nil, err
	}

	var records []map[string]interface{}
	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range columns {
			valuePtrs[i] = &values[i]
		}
		rows.Scan(valuePtrs...)

		record := make(map[string]interface{})
		for i, col := range columns {
			val := values[i]
			if b, ok := val.([]byte); ok {
				record[col] = string(b)
			} else {
				record[col] = val
			}
		}
		records = append(records, record)
	}

	return records, total, availableYears, nil
}

func tableExists(database *sql.DB, tableName string) bool {
	var name string
	err := database.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name=?", tableName).Scan(&name)
	return err == nil && name == tableName
}

// QueryHistory queries history from a single year table (for backward compatibility).
func QueryHistory(year int, page, pageSize int, keyword string) ([]models.HistoryRecord, int, error) {
	database, err := db.OpenHistoryDB()
	if err != nil {
		return nil, 0, err
	}

	db.EnsureHistoryTable(database, year)
	db.MigrateIfNeeded(database, year)

	tableName := db.GetYearTable(year)
	offset := (page - 1) * pageSize

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s", tableName)
	query := fmt.Sprintf(`SELECT id, bvid, aid, title, "desc", pic, duration, owner_name, owner_mid, 
		tag_name, tid, view_at, progress, business, view, danmaku, coin, favorite, like, reply, share 
		FROM %s`, tableName)

	var args []interface{}
	if keyword != "" {
		countQuery += " WHERE title LIKE ?"
		query += " WHERE title LIKE ?"
		args = append(args, "%"+keyword+"%")
	}

	var total int
	database.QueryRow(countQuery, args...).Scan(&total)

	query += " ORDER BY view_at DESC LIMIT ? OFFSET ?"
	queryArgs := append(args, pageSize, offset)

	rows, err := database.Query(query, queryArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var records []models.HistoryRecord
	for rows.Next() {
		var r models.HistoryRecord
		rows.Scan(&r.ID, &r.BVID, &r.AID, &r.Title, &r.Desc, &r.Pic, &r.Duration,
			&r.OwnerName, &r.OwnerMid, &r.TagName, &r.Tid, &r.ViewAt, &r.Progress,
			&r.Business, &r.View, &r.Danmaku, &r.Coin, &r.Favorite, &r.Like, &r.Reply, &r.Share)
		r.Thumbnail = r.Pic
		r.OwnerID = r.OwnerMid
		records = append(records, r)
	}

	return records, total, nil
}

// GetAvailableYears returns all years that have history data.
func GetAvailableYears() ([]int, error) {
	database, err := db.OpenHistoryDB()
	if err != nil {
		return nil, err
	}

	rows, err := database.Query("SELECT name FROM sqlite_master WHERE type='table' AND name LIKE 'bilibili_history_%' ORDER BY name DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var years []int
	for rows.Next() {
		var tableName string
		rows.Scan(&tableName)
		var year int
		if _, err := fmt.Sscanf(tableName, "bilibili_history_%d", &year); err == nil {
			years = append(years, year)
		}
	}

	if len(years) == 0 {
		years = []int{time.Now().Year()}
	}

	return years, nil
}

// GetDailyCounts returns daily watch counts for a given year.
func GetDailyCounts(year int) ([]models.DailyCount, error) {
	database, err := db.OpenHistoryDB()
	if err != nil {
		return nil, err
	}

	tableName := db.GetYearTable(year)
	query := fmt.Sprintf("SELECT date(view_at, 'unixepoch', 'localtime') as day, COUNT(*) FROM %s GROUP BY day ORDER BY day", tableName)

	rows, err := database.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var counts []models.DailyCount
	for rows.Next() {
		var dc models.DailyCount
		rows.Scan(&dc.Date, &dc.Count)
		counts = append(counts, dc)
	}

	return counts, nil
}

// GetHeatmapData returns heatmap data for a given year.
func GetHeatmapData(year int) (map[string]int, error) {
	database, err := db.OpenHistoryDB()
	if err != nil {
		return nil, err
	}

	tableName := db.GetYearTable(year)
	query := fmt.Sprintf("SELECT date(view_at, 'unixepoch', 'localtime') as day, COUNT(*) FROM %s GROUP BY day", tableName)

	rows, err := database.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	data := make(map[string]int)
	for rows.Next() {
		var day string
		var count int
		rows.Scan(&day, &count)
		data[day] = count
	}

	return data, nil
}

// DeleteHistoryRecord deletes a history record by BVID.
func DeleteHistoryRecord(year int, bvid string) (int64, error) {
	database, err := db.OpenHistoryDB()
	if err != nil {
		return 0, err
	}

	tableName := db.GetYearTable(year)
	result, err := database.Exec(fmt.Sprintf("DELETE FROM %s WHERE bvid = ?", tableName), bvid)
	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}
