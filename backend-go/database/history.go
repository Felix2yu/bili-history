package database

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"bilibili-history-go/models"
	"bilibili-history-go/utils"
)

type HistoryQueryParams struct {
	Page           int
	Size           int
	SortOrder      int
	TagName        string
	MainCategory   string
	DateRange      string
	Business       string
	UseLocalImages bool
	UseSessdata    bool
}

type HistorySearchParams struct {
	Page           int
	Size           int
	SortOrder      int
	Search         string
	SearchType     string
	UseLocalImages bool
	UseSessdata    bool
}

func GetHistoryPage(params HistoryQueryParams) (*models.PagedResponse, []int, error) {
	db := GetSQLiteDB()
	conn := db.GetDB()
	if conn == nil {
		return nil, nil, fmt.Errorf("database not initialized")
	}

	availableYears, err := db.GetAvailableYears()
	if err != nil {
		return nil, nil, err
	}

	var queries []string
	var paramsList []interface{}

	var startTimestamp, endTimestamp int64
	if params.DateRange != "" {
		parts := strings.Split(params.DateRange, "-")
		if len(parts) == 2 {
			startTime, err := time.Parse("20060102", parts[0])
			if err == nil {
				startTimestamp = startTime.Unix()
			}
			endTime, err := time.Parse("20060102", parts[1])
			if err == nil {
				endTimestamp = endTime.Unix() + 86400
			}
		}
	}

	for _, year := range availableYears {
		tableName := fmt.Sprintf("bilibili_history_%d", year)
		exists, _ := db.TableExists(tableName)
		if !exists {
			continue
		}

		query := fmt.Sprintf("SELECT * FROM %s WHERE 1=1", tableName)

		if startTimestamp > 0 && endTimestamp > 0 {
			query += " AND view_at >= ? AND view_at < ?"
			paramsList = append(paramsList, startTimestamp, endTimestamp)
		}

		if params.MainCategory != "" {
			query += " AND main_category = ?"
			paramsList = append(paramsList, params.MainCategory)
		} else if params.TagName != "" {
			query += " AND tag_name = ?"
			paramsList = append(paramsList, params.TagName)
		}

		if params.Business != "" {
			query += " AND business = ?"
			paramsList = append(paramsList, params.Business)
		}

		queries = append(queries, query)
	}

	if len(queries) == 0 {
		return &models.PagedResponse{
			Records: []models.HistoryRecord{},
			Total:   0,
			Size:    params.Size,
			Current: params.Page,
		}, availableYears, nil
	}

	baseQuery := strings.Join(queries, " UNION ALL ")

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM (%s)", baseQuery)
	var total int64
	err = conn.QueryRow(countQuery, paramsList...).Scan(&total)
	if err != nil {
		return nil, nil, err
	}

	sortDir := "DESC"
	if params.SortOrder == 1 {
		sortDir = "ASC"
	}

	finalQuery := fmt.Sprintf(`
		SELECT * FROM (%s)
		ORDER BY view_at %s
		LIMIT ? OFFSET ?
	`, baseQuery, sortDir)

	queryParams := make([]interface{}, len(paramsList))
	copy(queryParams, paramsList)
	queryParams = append(queryParams, params.Size, (params.Page-1)*params.Size)

	rows, err := conn.Query(finalQuery, queryParams...)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	records, err := scanHistoryRecords(rows)
	if err != nil {
		return nil, nil, err
	}

	processedRecords := processRecords(records, params.UseLocalImages, params.UseSessdata)

	return &models.PagedResponse{
		Records: processedRecords,
		Total:   total,
		Size:    params.Size,
		Current: params.Page,
	}, availableYears, nil
}

func SearchHistory(params HistorySearchParams) (*models.PagedResponse, []int, error) {
	db := GetSQLiteDB()
	conn := db.GetDB()
	if conn == nil {
		return nil, nil, fmt.Errorf("database not initialized")
	}

	availableYears, err := db.GetAvailableYears()
	if err != nil {
		return nil, nil, err
	}

	var subQueries []string
	var baseParams []interface{}

	fieldMap := map[string]string{
		"title":  "title",
		"author": "author_name",
		"tag":    "tag_name",
		"remark": "remark",
	}

	var whereClause string
	var searchParams []interface{}

	if params.Search != "" {
		searchKeyword := strings.TrimSpace(params.Search)
		if searchKeyword != "" {
			var fieldConditions []string

			if params.SearchType == "all" || params.SearchType == "" {
				for _, field := range fieldMap {
					fieldConditions = append(fieldConditions, fmt.Sprintf("%s LIKE ?", field))
					searchParams = append(searchParams, "%"+searchKeyword+"%")
				}
			} else {
				field, ok := fieldMap[params.SearchType]
				if ok {
					fieldConditions = append(fieldConditions, fmt.Sprintf("%s LIKE ?", field))
					searchParams = append(searchParams, "%"+searchKeyword+"%")
				}
			}

			if len(fieldConditions) > 0 {
				whereClause = "WHERE (" + strings.Join(fieldConditions, " OR ") + ")"
			}
		}
	}

	for _, year := range availableYears {
		tableName := fmt.Sprintf("bilibili_history_%d", year)
		exists, _ := db.TableExists(tableName)
		if !exists {
			continue
		}

		subQuery := fmt.Sprintf("SELECT * FROM %s %s", tableName, whereClause)
		subQueries = append(subQueries, subQuery)
		baseParams = append(baseParams, searchParams...)
	}

	if len(subQueries) == 0 {
		return &models.PagedResponse{
			Records: []models.HistoryRecord{},
			Total:   0,
			Size:    params.Size,
			Current: params.Page,
		}, availableYears, nil
	}

	baseQuery := strings.Join(subQueries, " UNION ALL ")

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM (%s)", baseQuery)
	var total int64
	err = conn.QueryRow(countQuery, baseParams...).Scan(&total)
	if err != nil {
		return nil, nil, err
	}

	sortDir := "DESC"
	if params.SortOrder == 1 {
		sortDir = "ASC"
	}

	query := fmt.Sprintf(`
		SELECT * FROM (%s)
		ORDER BY view_at %s
		LIMIT ? OFFSET ?
	`, baseQuery, sortDir)

	queryParams := make([]interface{}, len(baseParams))
	copy(queryParams, baseParams)
	queryParams = append(queryParams, params.Size, (params.Page-1)*params.Size)

	rows, err := conn.Query(query, queryParams...)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	records, err := scanHistoryRecords(rows)
	if err != nil {
		return nil, nil, err
	}

	processedRecords := processRecords(records, params.UseLocalImages, params.UseSessdata)

	return &models.PagedResponse{
		Records: processedRecords,
		Total:   total,
		Size:    params.Size,
		Current: params.Page,
	}, availableYears, nil
}

func scanHistoryRecords(rows *sql.Rows) ([]models.HistoryRecord, error) {
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	var records []models.HistoryRecord

	for rows.Next() {
		record := models.HistoryRecord{}
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))

		for i := range columns {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			continue
		}

		for i, col := range columns {
			val := values[i]
			if val == nil {
				continue
			}

			switch col {
			case "id":
				record.ID = toInt64(val)
			case "title":
				record.Title = toString(val)
			case "long_title":
				record.LongTitle = toString(val)
			case "cover":
				record.Cover = toString(val)
			case "covers":
				record.Covers = toString(val)
			case "uri":
				record.URI = toString(val)
			case "oid":
				record.OID = toInt64(val)
			case "epid":
				record.Epid = toInt64(val)
			case "bvid":
				record.Bvid = toString(val)
			case "page":
				record.Page = toInt(val)
			case "cid":
				record.Cid = toInt64(val)
			case "part":
				record.Part = toString(val)
			case "business":
				record.Business = toString(val)
			case "dt":
				record.Dt = toInt(val)
			case "videos":
				record.Videos = toInt(val)
			case "author_name":
				record.AuthorName = toString(val)
			case "author_face":
				record.AuthorFace = toString(val)
			case "author_mid":
				record.AuthorMid = toInt64(val)
			case "view_at":
				record.ViewAt = toInt64(val)
			case "progress":
				record.Progress = toInt(val)
			case "badge":
				record.Badge = toString(val)
			case "show_title":
				record.ShowTitle = toString(val)
			case "duration":
				record.Duration = toInt(val)
			case "current":
				record.Current = toString(val)
			case "total":
				record.Total = toInt(val)
			case "new_desc":
				record.NewDesc = toString(val)
			case "is_finish":
				record.IsFinish = toInt(val)
			case "is_fav":
				record.IsFav = toInt(val)
			case "kid":
				record.Kid = toInt64(val)
			case "tag_name":
				record.TagName = toString(val)
			case "live_status":
				record.LiveStatus = toInt(val)
			case "main_category":
				record.MainCategory = toString(val)
			case "remark":
				record.Remark = toString(val)
			case "remark_time":
				record.RemarkTime = toInt64(val)
			}
		}

		records = append(records, record)
	}

	return records, nil
}

func toInt64(val interface{}) int64 {
	switch v := val.(type) {
	case int64:
		return v
	case int:
		return int64(v)
	case float64:
		return int64(v)
	default:
		return 0
	}
}

func toInt(val interface{}) int {
	switch v := val.(type) {
	case int:
		return v
	case int64:
		return int(v)
	case float64:
		return int(v)
	default:
		return 0
	}
}

func toString(val interface{}) string {
	switch v := val.(type) {
	case string:
		return v
	case []byte:
		return string(v)
	default:
		return ""
	}
}

func processRecords(records []models.HistoryRecord, useLocalImages, useSessdata bool) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(records))
	for _, record := range records {
		result = append(result, processRecord(record, useLocalImages, useSessdata))
	}
	return result
}

func processRecord(record models.HistoryRecord, useLocalImages, useSessdata bool) map[string]interface{} {
	result := map[string]interface{}{
		"id":            record.ID,
		"title":         record.Title,
		"long_title":    record.LongTitle,
		"cover":         processImageURL(record.Cover, "covers", useLocalImages, useSessdata),
		"author_face":   processImageURL(record.AuthorFace, "avatars", useLocalImages, useSessdata),
		"uri":           record.URI,
		"oid":           record.OID,
		"epid":          record.Epid,
		"bvid":          record.Bvid,
		"page":          record.Page,
		"cid":           record.Cid,
		"part":          record.Part,
		"business":      record.Business,
		"dt":            record.Dt,
		"videos":        record.Videos,
		"author_name":   record.AuthorName,
		"author_mid":    record.AuthorMid,
		"view_at":       record.ViewAt,
		"progress":      record.Progress,
		"badge":         record.Badge,
		"show_title":    record.ShowTitle,
		"duration":      record.Duration,
		"current":       record.Current,
		"total":         record.Total,
		"new_desc":      record.NewDesc,
		"is_finish":     record.IsFinish,
		"is_fav":        record.IsFav,
		"kid":           record.Kid,
		"tag_name":      record.TagName,
		"live_status":   record.LiveStatus,
		"main_category": record.MainCategory,
		"remark":        record.Remark,
		"remark_time":   record.RemarkTime,
		"original_url":  buildOriginalURL(record),
		"view_time":     utils.FormatDateTime(utils.TimestampToTime(record.ViewAt)),
	}

	if record.Covers != "" {
		var coversList []string
		coversList = append(coversList, record.Covers)
		result["covers"] = coversList
	} else {
		result["covers"] = []string{}
	}

	return result
}

func processImageURL(url, imageType string, useLocal, useSessdata bool) string {
	if !useLocal || url == "" {
		if !useSessdata && strings.Contains(url, "?") {
			return strings.Split(url, "?")[0]
		}
		return url
	}
	return url
}

func buildOriginalURL(record models.HistoryRecord) string {
	uri := strings.TrimSpace(record.URI)
	if uri != "" {
		return uri
	}

	business := strings.ToLower(strings.TrimSpace(record.Business))
	bvid := strings.TrimSpace(record.Bvid)
	page := record.Page
	oid := record.OID
	epid := record.Epid

	switch business {
	case "archive":
		if bvid != "" {
			url := fmt.Sprintf("https://www.bilibili.com/video/%s", bvid)
			if page > 1 {
				url = fmt.Sprintf("%s?p=%d", url, page)
			}
			return url
		}
	case "pgc":
		if epid > 0 {
			return fmt.Sprintf("https://www.bilibili.com/bangumi/play/ep%d", epid)
		}
		if oid > 0 {
			return fmt.Sprintf("https://www.bilibili.com/bangumi/play/ss%d", oid)
		}
	case "live":
		if oid > 0 {
			return fmt.Sprintf("https://live.bilibili.com/%d", oid)
		}
	case "article", "article-list":
		if oid > 0 {
			return fmt.Sprintf("https://www.bilibili.com/read/cv%d", oid)
		}
	}

	if bvid != "" {
		return fmt.Sprintf("https://www.bilibili.com/video/%s", bvid)
	}

	return ""
}

func UpdateRemark(bvid string, viewAt int64, remark string) (map[string]interface{}, error) {
	db := GetSQLiteDB()
	conn := db.GetDB()
	if conn == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	year := utils.GetYearFromTimestamp(viewAt)
	tableName := fmt.Sprintf("bilibili_history_%d", year)

	exists, _ := db.TableExists(tableName)
	if !exists {
		return nil, fmt.Errorf("未找到 %d 年的历史记录数据", year)
	}

	currentTime := utils.NowUnix()

	result, err := conn.Exec(fmt.Sprintf(`
		UPDATE %s
		SET remark = ?, remark_time = ?
		WHERE bvid = ? AND view_at = ?
	`, tableName), remark, currentTime, bvid, viewAt)

	if err != nil {
		return nil, err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return nil, fmt.Errorf("未找到指定的视频记录")
	}

	return map[string]interface{}{
		"bvid":        bvid,
		"view_at":     viewAt,
		"remark":      remark,
		"remark_time": currentTime,
	}, nil
}

func GetVideoByCID(cid int64, useLocalImages, useSessdata bool) (map[string]interface{}, error) {
	db := GetSQLiteDB()
	conn := db.GetDB()
	if conn == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	availableYears, err := db.GetAvailableYears()
	if err != nil {
		return nil, err
	}

	var queries []string
	for _, year := range availableYears {
		tableName := fmt.Sprintf("bilibili_history_%d", year)
		exists, _ := db.TableExists(tableName)
		if !exists {
			continue
		}
		queries = append(queries, fmt.Sprintf(`
			SELECT
				id, title, long_title, cover, covers, uri, oid, epid, bvid, page,
				cid, part, business, dt, videos, author_name, author_face, author_mid,
				view_at, progress, badge, show_title, duration, current, total,
				new_desc, is_finish, is_fav, kid, tag_name, live_status, main_category,
				remark, remark_time
			FROM %s
			WHERE cid = %d
		`, tableName, cid))
	}

	if len(queries) == 0 {
		return nil, fmt.Errorf("未找到任何历史记录数据")
	}

	unionQuery := strings.Join(queries, " UNION ALL ") + " LIMIT 1"

	row := conn.QueryRow(unionQuery)

	record := models.HistoryRecord{}
	err = row.Scan(
		&record.ID, &record.Title, &record.LongTitle, &record.Cover, &record.Covers,
		&record.URI, &record.OID, &record.Epid, &record.Bvid, &record.Page,
		&record.Cid, &record.Part, &record.Business, &record.Dt, &record.Videos,
		&record.AuthorName, &record.AuthorFace, &record.AuthorMid, &record.ViewAt,
		&record.Progress, &record.Badge, &record.ShowTitle, &record.Duration,
		&record.Current, &record.Total, &record.NewDesc, &record.IsFinish,
		&record.IsFav, &record.Kid, &record.TagName, &record.LiveStatus,
		&record.MainCategory, &record.Remark, &record.RemarkTime,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("未找到CID为%d的视频记录", cid)
	}

	if err != nil {
		return nil, err
	}

	return processRecord(record, useLocalImages, useSessdata), nil
}
