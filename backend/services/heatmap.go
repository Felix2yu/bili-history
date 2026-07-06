package services

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"time"

	"bilibili-history-go/config"
	"bilibili-history-go/database"
	"bilibili-history-go/utils"
)

func getHeatmapOutputDir() string {
	cfg, _ := config.LoadConfig()
	if cfg != nil && cfg.Heatmap.OutputDir != "" {
		return cfg.Heatmap.OutputDir
	}
	if cfg != nil && cfg.OutputFolder != "" {
		return filepath.Join(cfg.OutputFolder, "heatmap")
	}
	return "output/heatmap"
}

func getHeatmapColors() []config.ColorPiece {
	cfg, _ := config.LoadConfig()
	if cfg != nil && len(cfg.Heatmap.Colors) > 0 {
		return cfg.Heatmap.Colors
	}
	return []config.ColorPiece{
		{Min: 1, Max: 10, Color: "#FFECF1"},
		{Min: 11, Max: 50, Color: "#FFB3CA"},
		{Min: 51, Max: 100, Color: "#FF8CB0"},
		{Min: 101, Max: 200, Color: "#FF6699"},
		{Min: 201, Max: 9999, Color: "#E84B85"},
	}
}

func colorFromHex(hex string) color.RGBA {
	var r, g, b uint8
	if len(hex) == 7 && hex[0] == '#' {
		fmt.Sscanf(hex[1:], "%02x%02x%02x", &r, &g, &b)
	}
	return color.RGBA{R: r, G: g, B: b, A: 255}
}

func getCountColor(count int, colors []config.ColorPiece) color.RGBA {
	for _, c := range colors {
		if count >= c.Min && count <= c.Max {
			return colorFromHex(c.Color)
		}
	}
	return color.RGBA{R: 255, G: 255, B: 255, A: 255}
}

func isLeapYear(year int) bool {
	return year%4 == 0 && (year%100 != 0 || year%400 == 0)
}

func daysInYear(year int) int {
	if isLeapYear(year) {
		return 366
	}
	return 365
}

// GenerateHeatmapPNG generates a calendar heatmap PNG for the given year.
func GenerateHeatmapPNG(year int) (string, error) {
	data, err := database.GenerateHeatmapData(year)
	if err != nil {
		return "", err
	}

	cellSize := 20
	cellGap := 4
	headerHeight := 40
	leftPadding := 50
	titleHeight := 50

	// Calculate grid dimensions: 7 rows (Sun-Sat), ~53 columns
	totalDays := daysInYear(year)
	cols := (totalDays + 6) / 7
	imgWidth := leftPadding + cols*(cellSize+cellGap) + cellGap
	imgHeight := titleHeight + headerHeight + 7*(cellSize+cellGap) + cellGap + 30

	img := image.NewRGBA(image.Rect(0, 0, imgWidth, imgHeight))

	// Background: white
	for y := 0; y < imgHeight; y++ {
		for x := 0; x < imgWidth; x++ {
			img.Set(x, y, color.RGBA{R: 255, G: 255, B: 255, A: 255})
		}
	}

	colors := getHeatmapColors()
	emptyColor := color.RGBA{R: 230, G: 230, B: 230, A: 255}

	// Draw title
	drawText(img, leftPadding, 30, fmt.Sprintf("Bilibili %d 年每日观看热力图", year), true)

	// Draw day labels (Mon, Wed, Fri)
	days := []string{"日", "一", "二", "三", "四", "五", "六"}
	for i, day := range days {
		y := titleHeight + headerHeight + i*(cellSize+cellGap) + cellSize/2 + 4
		if i%2 == 1 {
			drawText(img, 10, y, day, false)
		}
	}

	// Draw month labels
	months := []string{"1月", "2月", "3月", "4月", "5月", "6月",
		"7月", "8月", "9月", "10月", "11月", "12月"}
	yearStart := time.Date(year, 1, 1, 0, 0, 0, 0, time.Local)
	for m := 1; m <= 12; m++ {
		monthStart := time.Date(year, time.Month(m), 1, 0, 0, 0, 0, time.Local)
		dayOfYear := monthStart.YearDay() - 1
		week := dayOfYear / 7
		x := leftPadding + week*(cellSize+cellGap)
		drawText(img, x, titleHeight + headerHeight - 8, months[m-1], false)
	}

	// Draw cells
	for dayOfYear := 0; dayOfYear < totalDays; dayOfYear++ {
		date := yearStart.AddDate(0, 0, dayOfYear)
		week := dayOfYear / 7
		dayOfWeek := int(date.Weekday())

		x := leftPadding + week*(cellSize+cellGap)
		y := titleHeight + headerHeight + dayOfWeek*(cellSize+cellGap)

		dateStr := date.Format("2006-01-02")
		count := data.Data[dateStr]

		c := emptyColor
		if count > 0 {
			c = getCountColor(count, colors)
		}

		// Draw cell with rounded corners (approximate with rectangles)
		for dy := 0; dy < cellSize; dy++ {
			for dx := 0; dx < cellSize; dx++ {
				img.Set(x+dx, y+dy, c)
			}
		}
	}

	// Draw summary at bottom
	summaryY := titleHeight + headerHeight + 7*(cellSize+cellGap) + cellGap + 15
	drawText(img, leftPadding, summaryY, fmt.Sprintf("共 %d 天有观看记录，总计 %d 个视频", len(data.Data), data.Total), false)

	// Save to file
	outputDir := getHeatmapOutputDir()
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return "", fmt.Errorf("创建输出目录失败: %w", err)
	}

	filename := fmt.Sprintf("heatmap_%d.png", year)
	outputPath := filepath.Join(outputDir, filename)

	f, err := os.Create(outputPath)
	if err != nil {
		return "", fmt.Errorf("创建文件失败: %w", err)
	}
	defer f.Close()

	if err := png.Encode(f, img); err != nil {
		return "", fmt.Errorf("编码PNG失败: %w", err)
	}

	utils.LogSuccess("热力图已生成: %s", outputPath)
	return outputPath, nil
}

func drawText(img *image.RGBA, x, y int, text string, isBold bool) {
	// Simple pixel font rendering for basic Chinese characters
	// Each character is rendered as a small bitmap
	for i, ch := range text {
		_ = isBold
		offsetX := x + i*14

		// Render character using simple dot pattern
		renderChar(img, offsetX, y-10, ch)
	}
}

func renderChar(img *image.RGBA, x, y int, ch rune) {
	// Simple bitmap font for common characters
	// Using a minimal approach: draw character as filled blocks
	fontColor := color.RGBA{R: 51, G: 51, B: 51, A: 255}

	// For ASCII chars, use simple 8x12 bitmap
	if ch >= 32 && ch < 127 {
		// Simple ASCII rendering - draw a small rectangle per char
		for dy := 0; dy < 12; dy++ {
			for dx := 0; dx < 8; dx++ {
				// Simple pattern based on character
				if isCharPixel(ch, dx, dy) {
					img.Set(x+dx, y+dy, fontColor)
				}
			}
		}
		return
	}

	// For CJK characters, use a larger block
	for dy := 0; dy < 14; dy++ {
		for dx := 0; dx < 14; dx++ {
			if isCJKPixel(ch, dx, dy) {
				img.Set(x+dx, y+dy, fontColor)
			}
		}
	}
}

func isCharPixel(ch rune, x, y int) bool {
	// Simplified bitmap for digits and basic ASCII
	if ch >= '0' && ch <= '9' {
		return digitBitmap[ch-'0'][y*8+x]
	}
	// Default: simple block for other chars
	return x > 0 && x < 7 && y > 1 && y < 10
}

var digitBitmap = [10][96]bool{
	// 0
	{false, false, true, true, true, true, false, false,
		false, true, false, false, false, false, true, false,
		false, true, false, false, false, false, true, false,
		false, true, false, false, false, false, true, false,
		false, true, false, false, false, false, true, false,
		false, true, false, false, false, false, true, false,
		false, true, false, false, false, false, true, false,
		false, true, false, false, false, false, true, false,
		false, false, true, true, true, true, false, false,
		false, false, false, false, false, false, false, false,
		false, false, false, false, false, false, false, false,
		false, false, false, false, false, false, false, false},
	// 1
	{false, false, false, false, true, false, false, false,
		false, false, false, true, true, false, false, false,
		false, false, true, false, true, false, false, false,
		false, false, false, false, true, false, false, false,
		false, false, false, false, true, false, false, false,
		false, false, false, false, true, false, false, false,
		false, false, false, false, true, false, false, false,
		false, false, false, false, true, false, false, false,
		false, false, true, true, true, true, true, false,
		false, false, false, false, false, false, false, false,
		false, false, false, false, false, false, false, false,
		false, false, false, false, false, false, false, false},
	// 2
	{false, false, true, true, true, true, false, false,
		false, true, false, false, false, false, true, false,
		false, false, false, false, false, false, true, false,
		false, false, false, false, false, false, true, false,
		false, false, true, true, true, true, false, false,
		false, true, false, false, false, false, false, false,
		false, true, false, false, false, false, false, false,
		false, true, false, false, false, false, false, false,
		false, true, true, true, true, true, true, false,
		false, false, false, false, false, false, false, false,
		false, false, false, false, false, false, false, false,
		false, false, false, false, false, false, false, false},
	// 3
	{false, false, true, true, true, true, false, false,
		false, true, false, false, false, false, true, false,
		false, false, false, false, false, false, true, false,
		false, false, false, true, true, true, false, false,
		false, false, false, false, false, false, true, false,
		false, false, false, false, false, false, true, false,
		false, false, false, false, false, false, true, false,
		false, true, false, false, false, false, true, false,
		false, false, true, true, true, true, false, false,
		false, false, false, false, false, false, false, false,
		false, false, false, false, false, false, false, false,
		false, false, false, false, false, false, false, false},
	// 4
	{false, false, false, false, false, true, false, false,
		false, false, false, false, true, true, false, false,
		false, false, false, true, false, true, false, false,
		false, false, true, false, false, true, false, false,
		false, true, false, false, false, true, false, false,
		false, true, true, true, true, true, true, false,
		false, false, false, false, false, true, false, false,
		false, false, false, false, false, true, false, false,
		false, false, false, false, false, true, false, false,
		false, false, false, false, false, false, false, false,
		false, false, false, false, false, false, false, false,
		false, false, false, false, false, false, false, false},
	// 5
	{false, true, true, true, true, true, true, false,
		false, true, false, false, false, false, false, false,
		false, true, false, false, false, false, false, false,
		false, true, true, true, true, true, false, false,
		false, false, false, false, false, false, true, false,
		false, false, false, false, false, false, true, false,
		false, false, false, false, false, false, true, false,
		false, true, false, false, false, false, true, false,
		false, false, true, true, true, true, false, false,
		false, false, false, false, false, false, false, false,
		false, false, false, false, false, false, false, false,
		false, false, false, false, false, false, false, false},
	// 6
	{false, false, true, true, true, true, false, false,
		false, true, false, false, false, false, true, false,
		false, true, false, false, false, false, false, false,
		false, true, true, true, true, true, false, false,
		false, true, false, false, false, false, true, false,
		false, true, false, false, false, false, true, false,
		false, true, false, false, false, false, true, false,
		false, true, false, false, false, false, true, false,
		false, false, true, true, true, true, false, false,
		false, false, false, false, false, false, false, false,
		false, false, false, false, false, false, false, false,
		false, false, false, false, false, false, false, false},
	// 7
	{false, true, true, true, true, true, true, false,
		false, false, false, false, false, false, true, false,
		false, false, false, false, false, false, true, false,
		false, false, false, false, false, true, false, false,
		false, false, false, false, true, false, false, false,
		false, false, false, false, true, false, false, false,
		false, false, false, false, true, false, false, false,
		false, false, false, false, true, false, false, false,
		false, false, false, false, true, false, false, false,
		false, false, false, false, false, false, false, false,
		false, false, false, false, false, false, false, false,
		false, false, false, false, false, false, false, false},
	// 8
	{false, false, true, true, true, true, false, false,
		false, true, false, false, false, false, true, false,
		false, true, false, false, false, false, true, false,
		false, false, true, true, true, true, false, false,
		false, true, false, false, false, false, true, false,
		false, true, false, false, false, false, true, false,
		false, true, false, false, false, false, true, false,
		false, true, false, false, false, false, true, false,
		false, false, true, true, true, true, false, false,
		false, false, false, false, false, false, false, false,
		false, false, false, false, false, false, false, false,
		false, false, false, false, false, false, false, false},
	// 9
	{false, false, true, true, true, true, false, false,
		false, true, false, false, false, false, true, false,
		false, true, false, false, false, false, true, false,
		false, true, false, false, false, false, true, false,
		false, false, true, true, true, true, true, false,
		false, false, false, false, false, false, true, false,
		false, false, false, false, false, false, true, false,
		false, true, false, false, false, false, true, false,
		false, false, true, true, true, true, false, false,
		false, false, false, false, false, false, false, false,
		false, false, false, false, false, false, false, false,
		false, false, false, false, false, false, false, false},
}

// CJK bitmap patterns for characters used in heatmap labels
// Each pattern is a 10x10 grid (10 rows x 10 cols)
var cjkPatterns = map[rune][100]bool{
	// 月 (month)
	'月': {
		false, false, true, true, true, true, true, true, false, false,
		false, false, true, false, false, false, false, true, false, false,
		false, false, true, false, false, false, false, true, false, false,
		false, false, true, true, true, true, true, true, false, false,
		false, false, true, false, false, false, false, true, false, false,
		false, false, true, false, false, false, false, true, false, false,
		false, false, true, false, false, false, false, true, false, false,
		false, false, true, false, false, false, false, true, false, false,
		false, false, true, false, false, false, false, true, false, false,
		false, false, true, true, true, true, true, true, false, false},
	// 日 (day)
	'日': {
		false, true, true, true, true, true, true, true, true, false,
		false, true, false, false, false, false, false, false, true, false,
		false, true, false, false, false, false, false, false, true, false,
		false, true, false, false, false, false, false, false, true, false,
		false, true, true, true, true, true, true, true, true, false,
		false, true, false, false, false, false, false, false, true, false,
		false, true, false, false, false, false, false, false, true, false,
		false, true, false, false, false, false, false, false, true, false,
		false, true, false, false, false, false, false, false, true, false,
		false, true, true, true, true, true, true, true, true, false},
	// 年 (year)
	'年': {
		false, false, false, false, true, false, false, false, false, false,
		false, false, true, true, true, true, true, false, false, false,
		false, false, false, true, false, false, false, false, false, false,
		false, true, true, true, true, true, true, true, false, false,
		false, false, false, true, false, false, false, false, false, false,
		false, false, false, true, false, false, false, false, false, false,
		false, false, false, true, false, false, false, false, false, false,
		false, false, false, true, false, false, false, false, false, false,
		false, false, false, true, false, false, false, false, false, false,
		false, false, false, true, false, false, false, false, false, false},
	// 一 (one)
	'一': {
		false, false, false, false, false, false, false, false, false, false,
		false, false, false, false, false, false, false, false, false, false,
		false, false, false, false, false, false, false, false, false, false,
		false, false, false, false, false, false, false, false, false, false,
		false, true, true, true, true, true, true, true, true, false,
		false, false, false, false, false, false, false, false, false, false,
		false, false, false, false, false, false, false, false, false, false,
		false, false, false, false, false, false, false, false, false, false,
		false, false, false, false, false, false, false, false, false, false,
		false, false, false, false, false, false, false, false, false, false},
	// 二 (two)
	'二': {
		false, false, false, false, false, false, false, false, false, false,
		false, false, false, false, false, false, false, false, false, false,
		false, false, false, false, false, false, false, false, false, false,
		false, false, true, true, true, true, true, true, false, false,
		false, false, false, false, false, false, false, false, false, false,
		false, true, true, true, true, true, true, true, true, false,
		false, false, false, false, false, false, false, false, false, false,
		false, false, false, false, false, false, false, false, false, false,
		false, false, false, false, false, false, false, false, false, false,
		false, false, false, false, false, false, false, false, false, false},
	// 三 (three)
	'三': {
		false, false, false, false, false, false, false, false, false, false,
		false, false, false, false, false, false, false, false, false, false,
		false, true, true, true, true, true, true, true, false, false,
		false, false, false, false, false, false, false, false, false, false,
		false, false, true, true, true, true, true, false, false, false,
		false, false, false, false, false, false, false, false, false, false,
		false, true, true, true, true, true, true, true, true, false,
		false, false, false, false, false, false, false, false, false, false,
		false, false, false, false, false, false, false, false, false, false,
		false, false, false, false, false, false, false, false, false, false},
	// 四 (four)
	'四': {
		false, false, true, true, true, true, true, true, false, false,
		false, false, true, false, false, false, false, true, false, false,
		false, false, true, false, false, false, false, true, false, false,
		false, false, true, false, true, false, false, true, false, false,
		false, false, true, false, false, true, false, true, false, false,
		false, false, true, false, false, false, true, true, false, false,
		false, false, true, false, false, false, false, true, false, false,
		false, false, true, false, false, false, false, true, false, false,
		false, false, false, true, true, true, true, false, false, false,
		false, false, false, false, false, false, false, false, false, false},
	// 五 (five)
	'五': {
		false, true, true, true, true, true, true, true, false, false,
		false, true, false, false, false, false, false, false, false, false,
		false, true, false, false, false, false, false, false, false, false,
		false, true, true, true, true, true, true, false, false, false,
		false, false, false, false, false, false, false, true, false, false,
		false, false, false, false, false, false, false, true, false, false,
		false, false, false, false, false, false, false, true, false, false,
		false, true, false, false, false, false, false, true, false, false,
		false, false, true, true, true, true, true, false, false, false,
		false, false, false, false, false, false, false, false, false, false},
	// 六 (six)
	'六': {
		false, false, false, false, true, false, false, false, false, false,
		false, false, false, false, false, false, false, false, false, false,
		false, false, true, true, true, true, true, false, false, false,
		false, false, false, true, false, false, false, false, false, false,
		false, false, false, true, false, false, false, false, false, false,
		false, false, true, false, true, false, false, false, false, false,
		false, false, true, false, false, true, false, false, false, false,
		false, true, true, false, false, false, true, true, false, false,
		false, true, false, false, false, false, false, true, false, false,
		false, false, true, true, true, true, true, false, false, false},
	// 共 (total)
	'共': {
		false, false, false, false, true, false, false, false, false, false,
		false, true, true, true, true, true, true, true, false, false,
		false, false, false, true, false, false, false, false, false, false,
		false, false, false, true, false, false, false, false, false, false,
		false, true, true, true, true, true, true, true, true, false,
		false, false, false, true, false, false, false, false, false, false,
		false, false, true, false, false, true, false, false, false, false,
		false, true, false, false, false, false, true, false, false, false,
		true, false, false, false, false, false, false, false, true, false,
		false, false, false, false, false, false, false, false, false, false},
	// 有 (have)
	'有': {
		false, false, false, true, false, false, false, false, false, false,
		false, false, true, true, true, true, true, false, false, false,
		false, false, false, true, false, false, false, false, false, false,
		false, true, true, true, true, true, true, true, false, false,
		false, false, true, false, false, false, false, true, false, false,
		false, false, true, false, false, false, false, true, false, false,
		false, false, true, false, false, false, false, true, false, false,
		false, false, true, false, false, false, false, true, false, false,
		false, false, false, true, true, true, true, false, false, false,
		false, false, false, false, false, false, false, false, false, false},
	// 观 (watch)
	'观': {
		false, false, true, false, false, false, false, true, false, false,
		false, false, true, false, false, false, false, true, false, false,
		false, false, true, true, true, true, false, true, false, false,
		false, false, true, false, false, true, false, true, false, false,
		false, false, false, false, false, false, true, false, false, false,
		false, false, false, false, false, false, true, false, false, false,
		false, false, false, false, false, true, false, true, false, false,
		false, false, false, false, true, false, false, false, true, false,
		false, false, false, true, false, false, false, false, false, true,
		false, false, false, false, false, false, false, false, false, false},
	// 看 (watch/look)
	'看': {
		false, false, false, true, true, true, false, false, false, false,
		false, false, false, false, true, false, false, false, false, false,
		false, true, true, true, true, true, true, true, false, false,
		false, false, true, false, false, false, false, true, false, false,
		false, false, true, false, false, false, false, true, false, false,
		false, false, true, true, true, true, true, true, false, false,
		false, false, true, false, false, false, false, true, false, false,
		false, false, true, false, false, false, false, true, false, false,
		false, false, true, false, false, false, false, true, false, false,
		false, false, false, true, true, true, true, false, false, false},
	// 记 (record)
	'记': {
		false, false, false, false, false, false, true, false, false, false,
		false, false, true, true, true, false, false, true, false, false,
		false, false, false, false, false, false, false, true, false, false,
		false, false, false, false, false, true, true, true, true, false,
		false, false, false, false, false, false, false, true, false, false,
		false, false, false, false, false, false, false, true, false, false,
		false, false, false, false, false, false, false, true, false, false,
		false, false, false, false, false, false, false, true, false, false,
		false, false, false, false, false, true, true, false, false, false,
		false, false, false, false, false, false, false, false, false, false},
	// 录 (record)
	'录': {
		false, false, true, true, true, true, true, false, false, false,
		false, false, false, true, false, false, false, false, false, false,
		false, false, true, true, true, true, true, false, false, false,
		false, false, false, true, false, false, false, false, false, false,
		false, true, true, true, true, true, true, true, false, false,
		false, false, false, true, false, false, false, false, false, false,
		false, false, true, false, true, false, true, false, false, false,
		false, true, false, false, false, false, false, true, false, false,
		true, false, false, false, false, false, false, false, true, false,
		false, false, false, false, false, false, false, false, false, false},
	// 总 (total)
	'总': {
		false, false, false, false, true, false, false, false, false, false,
		false, false, false, true, false, true, false, false, false, false,
		false, true, true, true, true, true, true, true, false, false,
		false, false, true, false, false, false, false, true, false, false,
		false, false, false, true, true, true, true, false, false, false,
		false, false, false, false, true, false, false, false, false, false,
		false, false, false, true, false, true, false, false, false, false,
		false, false, true, false, false, false, true, false, false, false,
		false, true, false, false, false, false, false, true, false, false,
		false, false, true, true, false, true, true, false, false, false},
	// 计 (count)
	'计': {
		false, false, false, false, false, false, true, false, false, false,
		false, false, true, true, true, false, false, true, false, false,
		false, false, false, false, false, false, false, true, false, false,
		false, false, false, false, false, false, false, true, false, false,
		false, false, false, false, false, false, false, true, false, false,
		false, false, false, false, false, false, false, true, false, false,
		false, false, false, false, false, false, false, true, false, false,
		false, false, false, false, false, false, false, true, false, false,
		false, false, false, false, false, false, false, true, false, false,
		false, false, false, false, false, false, false, true, false, false},
	// 个 (unit)
	'个': {
		false, false, false, false, true, false, false, false, false, false,
		false, false, false, true, false, true, false, false, false, false,
		false, false, true, false, false, false, true, false, false, false,
		false, true, false, false, false, false, false, true, false, false,
		true, false, false, false, false, false, false, false, true, false,
		false, false, false, false, false, false, false, false, false, false,
		false, false, false, false, true, false, false, false, false, false,
		false, false, false, false, true, false, false, false, false, false,
		false, false, false, false, true, false, false, false, false, false,
		false, false, false, false, true, false, false, false, false, false},
	// 视 (video)
	'视': {
		false, false, true, false, false, false, false, true, false, false,
		false, false, true, false, false, false, false, true, false, false,
		false, false, true, true, true, true, false, true, false, false,
		false, false, true, false, false, true, false, true, false, false,
		false, false, false, false, false, false, true, false, false, false,
		false, false, false, false, false, false, true, false, false, false,
		false, false, false, false, false, true, false, true, false, false,
		false, false, false, false, true, false, false, false, true, false,
		false, false, false, true, false, false, false, false, false, true,
		false, false, false, false, false, false, false, false, false, false},
	// 频 (video/frequent)
	'频': {
		false, false, true, true, true, false, false, true, false, false,
		false, false, false, true, false, false, false, true, false, false,
		false, false, true, true, true, true, false, true, false, false,
		false, false, true, false, false, true, false, true, false, false,
		false, false, false, false, false, false, true, false, false, false,
		false, false, false, false, false, false, true, false, false, false,
		false, false, false, false, false, true, false, true, false, false,
		false, false, false, false, true, false, false, false, true, false,
		false, false, false, true, false, false, false, false, false, true,
		false, false, false, false, false, false, false, false, false, false},
}

func isCJKPixel(ch rune, x, y int) bool {
	if pattern, ok := cjkPatterns[ch]; ok {
		if x >= 0 && x < 10 && y >= 0 && y < 10 {
			return pattern[y*10+x]
		}
	}
	// Fallback: draw a solid block for unknown CJK characters
	if x >= 2 && x < 8 && y >= 2 && y < 8 {
		return true
	}
	return false
}
