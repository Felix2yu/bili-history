package routers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"bilibili-history-go/config"
	"bilibili-history-go/models"
	"bilibili-history-go/utils"

	"github.com/gin-gonic/gin"
)

func RegisterLoginRoutes(r *gin.RouterGroup) {
	login := r.Group("/login")
	{
		login.GET("/qrcode/generate", generateQRCode)
		login.GET("/qrcode/image", getQRCodeImage)
		login.GET("/qrcode/poll", pollScanStatus)
		login.POST("/logout", logout)
		login.GET("/check", checkLoginStatus)
		login.GET("/check-and-notify", checkAndNotify)
	}
}

const userAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36"

type QRCodeGenerateResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		URL       string `json:"url"`
		QRCodeKey string `json:"qrcode_key"`
	} `json:"data"`
}

func generateQRCode(c *gin.Context) {
	utils.LogInfo("开始生成二维码...")

	req, err := http.NewRequest("GET", "https://passport.bilibili.com/x/passport-login/web/qrcode/generate", nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse("网络请求失败: "+err.Error()))
		return
	}

	req.Header.Set("User-Agent", userAgent)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse("网络请求失败: "+err.Error()))
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse("读取响应失败: "+err.Error()))
		return
	}

	var result QRCodeGenerateResponse
	if err := json.Unmarshal(body, &result); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse("解析响应失败: "+err.Error()))
		return
	}

	if result.Code != 0 {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(fmt.Sprintf("B站API返回错误: %s", result.Message)))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(map[string]string{
		"qrcode_key": result.Data.QRCodeKey,
		"url":        result.Data.URL,
	}))
}

func getQRCodeImage(c *gin.Context) {
	c.JSON(http.StatusNotFound, models.ErrorResponse("二维码图片接口暂未实现"))
}

type QRCodePollResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Code          int    `json:"code"`
		Message       string `json:"message"`
		URL           string `json:"url"`
		RefreshToken  string `json:"refresh_token"`
		Timestamp     int64  `json:"timestamp"`
		CookieInfo    struct {
			Cookies []struct {
				Name  string `json:"name"`
				Value string `json:"value"`
			} `json:"cookies"`
		} `json:"cookie_info"`
	} `json:"data"`
}

func pollScanStatus(c *gin.Context) {
	qrcodeKey := c.Query("qrcode_key")
	if qrcodeKey == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("缺少必要的qrcode_key参数"))
		return
	}

	utils.LogInfo("开始轮询扫码状态，qrcode_key: %s", qrcodeKey)

	req, err := http.NewRequest("GET", "https://passport.bilibili.com/x/passport-login/web/qrcode/poll", nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse("网络请求失败: "+err.Error()))
		return
	}

	q := req.URL.Query()
	q.Add("qrcode_key", qrcodeKey)
	req.URL.RawQuery = q.Encode()
	req.Header.Set("User-Agent", userAgent)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse("网络请求失败: "+err.Error()))
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse("读取响应失败: "+err.Error()))
		return
	}

	var result QRCodePollResponse
	if err := json.Unmarshal(body, &result); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse("解析响应失败: "+err.Error()))
		return
	}

	if result.Code != 0 {
		c.JSON(http.StatusOK, gin.H{
			"status": "error",
			"data": map[string]interface{}{
				"code":      result.Code,
				"message":   result.Message,
				"timestamp": time.Now().Unix(),
			},
		})
		return
	}

	if result.Data.Code == 0 {
		cookies := make(map[string]string)

		for _, cookie := range result.Data.CookieInfo.Cookies {
			cookies[cookie.Name] = cookie.Value
		}

		if result.Data.URL != "" {
			parseURLCookies(result.Data.URL, cookies)
		}

		saveCookies(cookies)
		utils.LogInfo("登录成功，cookies已保存")
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data": map[string]interface{}{
			"code":      result.Data.Code,
			"message":   result.Data.Message,
			"timestamp": time.Now().Unix(),
		},
	})
}

func parseURLCookies(url string, cookies map[string]string) {
	if !contains(url, "?") {
		return
	}

	query := url[len(url)-len(url)+containsIndex(url, "?")+1:]
	for _, param := range split(query, "&") {
		if contains(param, "=") {
			parts := split(param, "=")
			name := parts[0]
			value := parts[1]
			if name == "DedeUserID" || name == "DedeUserID__ckMd5" || name == "SESSDATA" || name == "bili_jct" {
				cookies[name] = value
			}
		}
	}
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func containsIndex(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

func split(s, sep string) []string {
	var result []string
	for len(s) > 0 {
		idx := containsIndex(s, sep)
		if idx == -1 {
			result = append(result, s)
			break
		}
		result = append(result, s[:idx])
		s = s[idx+len(sep):]
	}
	return result
}

func saveCookies(cookies map[string]string) {
	cfg, err := config.LoadConfig()
	if err != nil {
		utils.LogError("加载配置失败: %v", err)
		return
	}

	if sessdata, ok := cookies["SESSDATA"]; ok {
		cfg.SESSDATA = sessdata
	}
	if biliJct, ok := cookies["bili_jct"]; ok {
		cfg.BiliJct = biliJct
	}
	if dedeUserID, ok := cookies["DedeUserID"]; ok {
		cfg.DedeUserID = dedeUserID
	}
	if dedeUserIDCkMd5, ok := cookies["DedeUserID__ckMd5"]; ok {
		cfg.DedeUserIDCkMd5 = dedeUserIDCkMd5
	}

	if err := config.SaveConfig(cfg); err != nil {
		utils.LogError("保存配置失败: %v", err)
	}
}

func logout(c *gin.Context) {
	cfg, err := config.LoadConfig()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse("加载配置失败"))
		return
	}

	cfg.SESSDATA = ""
	config.SaveConfig(cfg)

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "已成功退出登录",
	})
}

type NavResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	TTL     int         `json:"ttl"`
	Data    interface{} `json:"data"`
}

func checkLoginStatus(c *gin.Context) {
	cfg, _ := config.LoadConfig()
	if cfg == nil || cfg.SESSDATA == "" {
		c.JSON(http.StatusOK, gin.H{
			"code":    -101,
			"message": "未登录",
			"ttl":     1,
			"data":    nil,
		})
		return
	}

	req, err := http.NewRequest("GET", "https://api.bilibili.com/x/web-interface/nav", nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse("请求失败: "+err.Error()))
		return
	}

	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Cookie", fmt.Sprintf("SESSDATA=%s", cfg.SESSDATA))

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse("网络请求失败: "+err.Error()))
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse("读取响应失败: "+err.Error()))
		return
	}

	var result interface{}
	json.Unmarshal(body, &result)

	c.JSON(http.StatusOK, result)
}

func checkAndNotify(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "检查功能暂未完全实现",
		"data": map[string]interface{}{
			"valid":   false,
			"notified": false,
		},
	})
}
