package routers

import (
	"crypto/subtle"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"bilibili-history-go/internal/config"
	"bilibili-history-go/internal/database"
	"bilibili-history-go/internal/models"

	"github.com/gin-gonic/gin"
)

// These endpoints are meant for the 满屏 macOS client: decorate covers with a
// watched marker / progress bar, and write back positions the player reports.
// They stay behind an optional token that is separate from the web UI's own
// session, so enabling it never locks the browser out.

const maxWatchProgressIDs = 500

func RegisterWatchProgressRoutes(r *gin.RouterGroup) {
	group := r.Group("/watch-progress", requireWatchProgressToken())
	group.POST("/ping", pingWatchProgress)
	group.POST("/query", queryWatchProgress)
	group.POST("/report", reportWatchProgress)
}

func requireWatchProgressToken() gin.HandlerFunc {
	return func(c *gin.Context) {
		cfg := config.GetConfig()
		expected := ""
		if cfg != nil {
			expected = strings.TrimSpace(cfg.APIToken)
		}
		if expected == "" {
			c.Next()
			return
		}
		supplied := c.GetHeader("X-API-Token")
		if auth := c.GetHeader("Authorization"); strings.HasPrefix(strings.ToLower(auth), "bearer ") {
			supplied = strings.TrimSpace(auth[len("Bearer "):])
		}
		if subtle.ConstantTimeCompare([]byte(supplied), []byte(expected)) == 1 {
			c.Next()
			return
		}
		c.AbortWithStatusJSON(http.StatusUnauthorized, models.ErrorResponse("访问令牌无效"))
	}
}

func pingWatchProgress(c *gin.Context) {
	c.JSON(http.StatusOK, models.SuccessResponse(gin.H{"ok": true}))
}

func queryWatchProgress(c *gin.Context) {
	var body struct {
		Bvids []string `json:"bvids"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("参数错误: "+err.Error()))
		return
	}
	if len(body.Bvids) > maxWatchProgressIDs {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(fmt.Sprintf("一次最多查询 %d 个视频", maxWatchProgressIDs)))
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

	result, err := database.BatchGetWatchProgress(bvids)
	if err != nil {
		c.JSON(http.StatusOK, models.ErrorResponse(err.Error()))
		return
	}
	c.JSON(http.StatusOK, models.SuccessResponse(result))
}

func reportWatchProgress(c *gin.Context) {
	var body struct {
		Bvid      string `json:"bvid"`
		Aid       int64  `json:"aid"`
		Cid       int64  `json:"cid"`
		Title     string `json:"title"`
		Author    string `json:"author"`
		AuthorMid int64  `json:"author_mid"`
		Cover     string `json:"cover"`
		Duration  int    `json:"duration"`
		Progress  int    `json:"progress"`
		ViewAt    int64  `json:"view_at"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("参数错误: "+err.Error()))
		return
	}
	body.Bvid = strings.TrimSpace(body.Bvid)
	if body.Bvid == "" || body.Progress < 0 {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("需要 bvid 和非负的 progress"))
		return
	}

	now := time.Now()
	dt, _ := strconv.Atoi(now.Format("20060102"))
	// The client sends the timestamp of the viewing session (taken from Bilibili's own
	// history when available), not a wall-clock reading. Ignore nonsense values.
	viewAt := now.Unix()
	if body.ViewAt > 0 && body.ViewAt <= now.Add(24*time.Hour).Unix() {
		viewAt = body.ViewAt
	}
	record := models.HistoryRecord{
		Bvid:       body.Bvid,
		Title:      body.Title,
		LongTitle:  body.Title,
		ShowTitle:  body.Title,
		Part:       body.Title,
		Cover:      body.Cover,
		URI:        fmt.Sprintf("bilibili://video/%d", body.Aid),
		OID:        body.Aid,
		Cid:        body.Cid,
		Page:       1,
		Business:   "video",
		Dt:         dt,
		Videos:     1,
		AuthorName: body.Author,
		AuthorMid:  body.AuthorMid,
		ViewAt:     viewAt,
		Progress:   body.Progress,
		Current:    formatWatchPosition(body.Progress),
		Total:      1,
		Duration:   body.Duration,
	}
	if err := database.UpsertWatchProgress(&record); err != nil {
		c.JSON(http.StatusOK, models.ErrorResponse(err.Error()))
		return
	}
	c.JSON(http.StatusOK, models.SuccessResponse(gin.H{
		"bvid": body.Bvid, "progress": body.Progress, "view_at": record.ViewAt,
	}))
}

func formatWatchPosition(seconds int) string {
	if seconds < 0 {
		seconds = 0
	}
	return fmt.Sprintf("%02d:%02d", seconds/60, seconds%60)
}
