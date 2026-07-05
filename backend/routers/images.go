package routers

import (
	"crypto/md5"
	"encoding/hex"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"bilibili-history-go/config"
	"bilibili-history-go/models"
	"bilibili-history-go/services"
	"bilibili-history-go/utils"

	"github.com/gin-gonic/gin"
)

func RegisterImageRoutes(r *gin.RouterGroup) {
	images := r.Group("/images")
	{
		images.POST("/start", startImageDownload)
		images.POST("/stop", stopImageDownload)
		images.GET("/status", getImageDownloadStatus)
		images.POST("/clear", clearImages)
		images.GET("/local/:image_type/:file_hash", getLocalImage)
		images.GET("/proxy", proxyImage)
	}
}

func startImageDownload(c *gin.Context) {
	yearStr := c.Query("year")
	useSessdata := c.Query("use_sessdata") != "false"

	var year *int
	if yearStr != "" {
		y, err := strconv.Atoi(yearStr)
		if err == nil && y > 2000 && y < 2100 {
			year = &y
		}
	}

	go services.StartFullImageDownload(year, useSessdata)

	message := "开始下载所有年份的图片"
	if year != nil {
		message = "开始下载" + strconv.Itoa(*year) + "年的图片"
	}

	c.JSON(http.StatusOK, models.SuccessResponse(gin.H{
		"message": message,
	}))
}

func stopImageDownload(c *gin.Context) {
	services.StopImageDownload()
	c.JSON(http.StatusOK, models.SuccessResponse(gin.H{
		"message": "已停止图片下载",
	}))
}

func getImageDownloadStatus(c *gin.Context) {
	status := services.GetImageDownloadStatus()

	// 使用 config 中的 output_folder 或默认路径
	cfg := config.GetConfig()
	outputFolder := "output"
	if cfg != nil && cfg.OutputFolder != "" {
		outputFolder = cfg.OutputFolder
	}

	// 尝试多个可能的路径
	possiblePaths := []string{
		filepath.Join("/app", outputFolder),
		outputFolder,
		utils.GetOutputPath(),
	}

	var coversDir, avatarsDir, dynamicDir string
	for _, basePath := range possiblePaths {
		testCovers := filepath.Join(basePath, "images", "covers")
		if _, err := os.Stat(testCovers); err == nil {
			coversDir = testCovers
			avatarsDir = filepath.Join(basePath, "images", "avatars")
			dynamicDir = filepath.Join(basePath, "dynamic")
			break
		}
	}

	coversTotal, coversDownloaded := countLocalImages(coversDir)
	avatarsTotal, avatarsDownloaded := countLocalImages(avatarsDir)
	dynamicTotal, dynamicDownloaded := countAllDynamicImages(dynamicDir)

	c.JSON(http.StatusOK, models.SuccessResponse(gin.H{
		"is_downloading": status.IsRunning,
		"last_update":    status.StartTime,
		"covers": gin.H{
			"total":         coversTotal,
			"downloaded":    coversDownloaded,
			"failed":        0,
			"total_planned": status.TotalImages / 2,
			"failed_urls":   []interface{}{},
		},
		"avatars": gin.H{
			"total":         avatarsTotal + dynamicTotal,
			"downloaded":    avatarsDownloaded + dynamicDownloaded,
			"failed":        0,
			"total_planned": status.TotalImages / 2,
			"failed_urls":   []interface{}{},
		},
	}))
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func countLocalImages(dir string) (total int, downloaded int) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return 0, 0
	}

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		// 统计所有图片文件（jpg, png, webp, gif）
		if !info.IsDir() {
			ext := strings.ToLower(filepath.Ext(path))
			if ext == ".jpg" || ext == ".jpeg" || ext == ".png" || ext == ".webp" || ext == ".gif" {
				total++
				downloaded++
			}
		}
		return nil
	})
	if err != nil {
		return 0, 0
	}
	return total, downloaded
}

// countAllDynamicImages 统计所有动态目录下的图片
func countAllDynamicImages(dynamicDir string) (total int, downloaded int) {
	if _, err := os.Stat(dynamicDir); os.IsNotExist(err) {
		return 0, 0
	}

	// 遍历 dynamic 目录下的所有子目录
	entries, err := os.ReadDir(dynamicDir)
	if err != nil {
		return 0, 0
	}

	for _, entry := range entries {
		if entry.IsDir() {
			imagesDir := filepath.Join(dynamicDir, entry.Name(), "images")
			t, d := countLocalImages(imagesDir)
			total += t
			downloaded += d
		}
	}
	return total, downloaded
}

func clearImages(c *gin.Context) {
	cleared := services.ClearAllImages()
	if cleared {
		c.JSON(http.StatusOK, models.SuccessResponse(gin.H{
			"message": "已清空所有图片和下载状态",
			"cleared_paths": []string{
				"output/images/covers",
				"output/images/avatars",
				"output/images/proxy",
			},
		}))
	} else {
		c.JSON(http.StatusOK, gin.H{
			"status":  "error",
			"message": "清空图片失败",
		})
	}
}

func getLocalImage(c *gin.Context) {
	imageType := c.Param("image_type")
	fileHash := c.Param("file_hash")

	if imageType != "covers" && imageType != "avatars" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的图片类型"})
		return
	}

	imgPath := findLocalImage(imageType, fileHash)
	if imgPath == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "图片不存在"})
		return
	}

	c.File(imgPath)
}

func findLocalImage(imageType, fileHash string) string {
	outputPath := utils.GetOutputPath("images")
	typePath := filepath.Join(outputPath, imageType)

	if len(fileHash) >= 2 {
		subDir := fileHash[:2]

		years := []string{}
		if entries, err := os.ReadDir(typePath); err == nil {
			for _, entry := range entries {
				if entry.IsDir() {
					name := entry.Name()
					if len(name) == 4 {
						if _, err := strconv.Atoi(name); err == nil {
							years = append(years, name)
						}
					}
				}
			}
		}

		for i := len(years) - 1; i >= 0; i-- {
			yearPath := filepath.Join(typePath, years[i], subDir)
			if path := findImageByHash(yearPath, fileHash); path != "" {
				return path
			}
		}

		rootSubDir := filepath.Join(typePath, subDir)
		if path := findImageByHash(rootSubDir, fileHash); path != "" {
			return path
		}
	}

	return ""
}

func findImageByHash(dir, fileHash string) string {
	exts := []string{".jpg", ".jpeg", ".png", ".webp", ".gif"}
	for _, ext := range exts {
		imgPath := filepath.Join(dir, fileHash+ext)
		if _, err := os.Stat(imgPath); err == nil {
			return imgPath
		}
	}
	return ""
}

func proxyImage(c *gin.Context) {
	url := c.Query("url")
	if url == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "URL不能为空"})
		return
	}

	if strings.HasPrefix(url, "http://") {
		url = "https://" + strings.TrimPrefix(url, "http://")
	}

	hash := md5.Sum([]byte(url))
	fileHash := hex.EncodeToString(hash[:])
	ext := filepath.Ext(url)
	if ext == "" || len(ext) > 5 {
		ext = ".jpg"
	}

	outputPath := utils.GetOutputPath("images")
	proxyDir := filepath.Join(outputPath, "proxy", fileHash[:2])
	os.MkdirAll(proxyDir, 0755)
	savePath := filepath.Join(proxyDir, fileHash+ext)

	if _, err := os.Stat(savePath); err == nil {
		c.File(savePath)
		return
	}

	localPath, err := services.DownloadImage(url, "proxy/"+fileHash[:2], fileHash+ext)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "下载失败: " + err.Error()})
		return
	}

	c.File(localPath)
}
