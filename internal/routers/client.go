package routers

import (
	"net/http"
	"sort"
	"strconv"
	"strings"

	"bilibili-history-go/internal/database"
	"bilibili-history-go/internal/models"

	"github.com/gin-gonic/gin"
)

// The client channel is what the 满屏 macOS player reads. It is a separate
// group with the same optional token as /watch-progress so a deployment either
// opens both or neither, and so the browser session is never involved.
//
// Every endpoint here is read-only. A client may report a play position
// (POST /watch-progress/report), but it never deletes, renames or re-crawls:
// those decisions stay with the web UI and the scheduler.

const maxClientBatchIDs = 500

func RegisterClientRoutes(r *gin.RouterGroup) {
	group := r.Group("/client", requireClientToken())
	group.POST("/recent", clientRecent)
	group.POST("/history", clientHistory)
	group.POST("/video-status", clientVideoStatus)
	group.GET("/categories", clientCategories)
	group.GET("/summary", clientSummary)
	group.GET("/watchlater", clientWatchLater)
	group.GET("/favorites", clientFavorites)
	group.GET("/favorites/contents", clientFavoriteContents)
}

// clientRecent is the 继续观看 queue: started, not finished, newest first.
func clientRecent(c *gin.Context) {
	var body struct {
		Limit int `json:"limit"`
	}
	// An empty body is a valid "give me the default page".
	if c.Request.ContentLength > 0 {
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, models.ErrorResponse("参数错误: "+err.Error()))
			return
		}
	}
	records, err := database.RecentWatch(body.Limit)
	if err != nil {
		c.JSON(http.StatusOK, models.ErrorResponse(err.Error()))
		return
	}
	c.JSON(http.StatusOK, models.SuccessResponse(gin.H{"records": records}))
}

// clientHistory pages through the permanent archive. The cursor is the last
// view_at of the previous page, so rows arriving at the top during scroll do not
// shift the page boundary.
func clientHistory(c *gin.Context) {
	var body struct {
		Cursor       int64  `json:"cursor"`
		Limit        int    `json:"limit"`
		Keyword      string `json:"keyword"`
		Business     string `json:"business"`
		Year         int    `json:"year"`
		MainCategory string `json:"main_category"`
		SubCategory  string `json:"sub_category"`
		MinDuration  int    `json:"min_duration"`
		MaxDuration  int    `json:"max_duration"`
	}
	if c.Request.ContentLength > 0 {
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, models.ErrorResponse("参数错误: "+err.Error()))
			return
		}
	}
	if body.Cursor < 0 {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("cursor 不能为负"))
		return
	}
	if body.Year != 0 && (body.Year < 1970 || body.Year > 2100) {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("year 超出范围"))
		return
	}
	if body.MinDuration < 0 || body.MaxDuration < 0 {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("时长不能为负"))
		return
	}
	if body.MinDuration > 0 && body.MaxDuration > 0 && body.MinDuration > body.MaxDuration {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("最小时长不能超过最大时长"))
		return
	}
	page, err := database.ArchiveHistory(database.ArchiveHistoryParams{
		Cursor: body.Cursor, Limit: body.Limit, Keyword: body.Keyword,
		Business: strings.TrimSpace(body.Business), Year: body.Year,
		MainCategory: strings.TrimSpace(body.MainCategory), SubCategory: strings.TrimSpace(body.SubCategory),
		MinDuration: body.MinDuration, MaxDuration: body.MaxDuration,
	})
	if err != nil {
		c.JSON(http.StatusOK, models.ErrorResponse(err.Error()))
		return
	}
	c.JSON(http.StatusOK, models.SuccessResponse(page))
}

func clientCategories(c *gin.Context) {
	categories, err := database.GetCategories()
	if err != nil {
		c.JSON(http.StatusOK, models.ErrorResponse(err.Error()))
		return
	}
	c.JSON(http.StatusOK, models.SuccessResponse(gin.H{"categories": categories}))
}

// clientVideoStatus answers from the cached delete markers only. It never calls
// Bilibili: a screenful of covers must not turn into a burst of requests, and a
// wrong "deleted" badge is worse than a missing one.
func clientVideoStatus(c *gin.Context) {
	var body struct {
		Bvids []string `json:"bvids"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("参数错误: "+err.Error()))
		return
	}
	if len(body.Bvids) > maxClientBatchIDs {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("一次最多查询 500 个视频"))
		return
	}
	seen := make(map[string]bool, len(body.Bvids))
	bvids := make([]string, 0, len(body.Bvids))
	for _, bvid := range body.Bvids {
		bvid = strings.TrimSpace(bvid)
		if bvid == "" || seen[bvid] {
			continue
		}
		seen[bvid] = true
		bvids = append(bvids, bvid)
	}

	status := database.GetDeletedStatus(bvids)
	deleted := make([]string, 0, len(status))
	for bvid, gone := range status {
		if gone {
			deleted = append(deleted, bvid)
		}
	}
	// Map iteration order is random; a stable list keeps tests and client diffs sane.
	sort.Strings(deleted)
	c.JSON(http.StatusOK, models.SuccessResponse(gin.H{"deleted": deleted}))
}

// clientSummary is the one-line strip above the archive: a handful of numbers
// whose definitions match the web monthly report.
func clientSummary(c *gin.Context) {
	year, _ := strconv.Atoi(c.DefaultQuery("year", "0"))
	month, _ := strconv.Atoi(c.DefaultQuery("month", "0"))
	if month < 0 || month > 12 {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("month 需要在 0 到 12 之间"))
		return
	}
	if year != 0 && (year < 1970 || year > 2100) {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("year 超出范围"))
		return
	}
	summary, err := database.WatchSummary(year, month)
	if err != nil {
		c.JSON(http.StatusOK, models.ErrorResponse(err.Error()))
		return
	}
	c.JSON(http.StatusOK, models.SuccessResponse(summary))
}

func clientWatchLater(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "40"))
	minDuration, _ := strconv.Atoi(c.DefaultQuery("min_duration", "0"))
	maxDuration, _ := strconv.Atoi(c.DefaultQuery("max_duration", "0"))
	if minDuration < 0 || maxDuration < 0 {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("时长不能为负"))
		return
	}
	if minDuration > 0 && maxDuration > 0 && minDuration > maxDuration {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("最小时长不能超过最大时长"))
		return
	}
	result, err := database.WatchLaterPage(database.WatchLaterParams{
		Page: page, Size: size,
		MainCategory: strings.TrimSpace(c.Query("main_category")),
		SubCategory:  strings.TrimSpace(c.Query("sub_category")),
		MinDuration:  minDuration, MaxDuration: maxDuration,
	})
	if err != nil {
		c.JSON(http.StatusOK, models.ErrorResponse(err.Error()))
		return
	}
	c.JSON(http.StatusOK, models.SuccessResponse(result))
}

func clientFavorites(c *gin.Context) {
	folders, err := database.FavoriteFolders()
	if err != nil {
		c.JSON(http.StatusOK, models.ErrorResponse(err.Error()))
		return
	}
	c.JSON(http.StatusOK, models.SuccessResponse(gin.H{"folders": folders}))
}

func clientFavoriteContents(c *gin.Context) {
	mediaID, err := strconv.ParseInt(c.Query("media_id"), 10, 64)
	if err != nil || mediaID <= 0 {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("需要 media_id"))
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "40"))
	items, total, err := database.FavoriteContents(mediaID, page, size)
	if err != nil {
		c.JSON(http.StatusOK, models.ErrorResponse(err.Error()))
		return
	}
	c.JSON(http.StatusOK, models.SuccessResponse(gin.H{
		"records": items, "total": total, "page": maxInt(page, 1), "size": clampedSize(size),
	}))
}

func maxInt(value, floor int) int {
	if value < floor {
		return floor
	}
	return value
}

// clampedSize echoes the page size the database actually used, so a client that
// asked for 500 knows it got 100.
func clampedSize(size int) int {
	if size <= 0 {
		return 40
	}
	if size > 100 {
		return 100
	}
	return size
}
