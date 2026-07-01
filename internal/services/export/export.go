package export

import (
	"fmt"
	"os"
	"time"

	"bili-history/internal/config"
	"bili-history/internal/db"
)

// ExportToExcel exports history data to an Excel file.
func ExportToExcel(year int) (string, error) {
	database, err := db.OpenHistoryDB()
	if err != nil {
		return "", err
	}

	tableName := db.GetYearTable(year)
	rows, err := database.Query(fmt.Sprintf(`
		SELECT bvid, title, owner_name, duration, view_at, progress
		FROM %s ORDER BY view_at DESC`, tableName))
	if err != nil {
		return "", err
	}
	defer rows.Close()

	// Generate CSV as a simple export format
	outputPath := config.GetOutputPath(fmt.Sprintf("export_history_%d.csv", year))
	file, err := os.Create(outputPath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// Write BOM for Excel UTF-8 compatibility
	file.Write([]byte{0xEF, 0xBB, 0xBF})

	// Write header
	file.WriteString("BVID,Title,Author,Duration(s),WatchedAt,Progress(s)\n")

	count := 0
	for rows.Next() {
		var bvid, title, ownerName string
		var duration, viewAt, progress int
		rows.Scan(&bvid, &title, &ownerName, &duration, &viewAt, &progress)

		watchedAt := time.Unix(int64(viewAt), 0).Format("2006-01-02 15:04:05")
		line := fmt.Sprintf("%s,\"%s\",\"%s\",%d,%s,%d\n",
			bvid, escapeCSV(title), escapeCSV(ownerName), duration, watchedAt, progress)
		file.WriteString(line)
		count++
	}

	return outputPath, nil
}

func escapeCSV(s string) string {
	// Simple CSV escaping
	result := ""
	for _, c := range s {
		if c == '"' {
			result += "\"\""
		} else {
			result += string(c)
		}
	}
	return result
}
