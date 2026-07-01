package handlers

import (
	"net/http"
	"time"

	"bili-history/internal/config"

	"github.com/gin-gonic/gin"
	qrcode "github.com/skip2/go-qrcode"
)

// GenerateQRCode generates a QR code for Bilibili login.
func GenerateQRCode(c *gin.Context) {
	// Call Bilibili API to get QR code URL
	apiURL := "https://passport.bilibili.com/x/passport-login/web/qrcode/generate"

	resp, err := http.Get(apiURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate QR code"})
		return
	}
	defer resp.Body.Close()

	var result struct {
		Code int `json:"code"`
		Data struct {
			URL    string `json:"url"`
			Key    string `json:"qrcode_key"`
			ImgURL string `json:"img_url"`
		} `json:"data"`
	}

	if err := parseJSON(resp.Body, &result); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse response"})
		return
	}

	if result.Code != 0 {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Bilibili API error"})
		return
	}

	// Return format matching Python backend
	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data": gin.H{
			"qrcode_key": result.Data.Key,
			"url":        result.Data.URL,
		},
	})
}

// GenerateQRCodeImage returns the QR code as a PNG image.
func GenerateQRCodeImage(c *gin.Context) {
	// Call Bilibili API to get QR code URL
	apiURL := "https://passport.bilibili.com/x/passport-login/web/qrcode/generate"

	resp, err := http.Get(apiURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate QR code"})
		return
	}
	defer resp.Body.Close()

	var result struct {
		Code int `json:"code"`
		Data struct {
			URL    string `json:"url"`
			Key    string `json:"qrcode_key"`
			ImgURL string `json:"img_url"`
		} `json:"data"`
	}

	if err := parseJSON(resp.Body, &result); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse response"})
		return
	}

	if result.Code != 0 {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Bilibili API error"})
		return
	}

	// Generate QR code PNG
	png, err := qrcode.Encode(result.Data.URL, qrcode.Medium, 256)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to encode QR code"})
		return
	}

	c.Data(http.StatusOK, "image/png", png)
}

// PollLogin polls the QR code login status.
func PollLogin(c *gin.Context) {
	key := c.Query("qrcode_key")
	if key == "" {
		key = c.Query("key") // fallback
	}
	if key == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "qrcode_key is required"})
		return
	}

	apiURL := "https://passport.bilibili.com/x/passport-login/web/qrcode/poll?qrcode_key=" + key

	req, _ := http.NewRequest("GET", apiURL, nil)
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Referer", "https://passport.bilibili.com/")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to poll login status"})
		return
	}
	defer resp.Body.Close()

	var result struct {
		Code int `json:"code"`
		Data struct {
			URL         string `json:"url"`
			RefreshURL  string `json:"refresh_url"`
			Timestamp   int64  `json:"timestamp"`
			Code        int    `json:"code"`
			Message     string `json:"message"`
			SESSDATA    string `json:"SESSDATA"`
			BiliJCT     string `json:"bili_jct"`
			DedeUserID  string `json:"DedeUserID"`
			DedeUserIDCkMd5 string `json:"DedeUserID__ckMd5"`
			SecureCookie string `json:"secure_cookie"`
			CookieInfo  struct {
				Cookies []struct {
					Name  string `json:"name"`
					Value string `json:"value"`
				} `json:"cookies"`
			} `json:"cookie_info"`
		} `json:"data"`
	}

	if err := parseJSON(resp.Body, &result); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse response"})
		return
	}

	switch result.Data.Code {
	case 0: // Success
		cookies := extractCookies(resp, result.Data.URL, result.Data.CookieInfo.Cookies,
			result.Data.SESSDATA, result.Data.BiliJCT, result.Data.DedeUserID, result.Data.DedeUserIDCkMd5)

		if len(cookies) == 0 || cookies["SESSDATA"] == "" {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to extract login cookies"})
			return
		}

		if err := config.SaveCookies(cookies); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save cookies"})
			return
		}
		config.ReloadConfig()

		c.JSON(http.StatusOK, gin.H{
			"status": "success",
			"data": gin.H{
				"code":      0,
				"message":   "登录成功",
				"timestamp": time.Now().Unix(),
			},
		})
	case 86090: // Already scanned, waiting for confirmation
		c.JSON(http.StatusOK, gin.H{
			"status": "success",
			"data": gin.H{
				"code":      86090,
				"message":   "已扫码，请在手机上确认",
				"timestamp": time.Now().Unix(),
			},
		})
	case 86038: // QR code expired
		c.JSON(http.StatusOK, gin.H{
			"status": "success",
			"data": gin.H{
				"code":      86038,
				"message":   "二维码已失效",
				"timestamp": time.Now().Unix(),
			},
		})
	default:
		c.JSON(http.StatusOK, gin.H{
			"status": "success",
			"data": gin.H{
				"code":      result.Data.Code,
				"message":   result.Data.Message,
				"timestamp": time.Now().Unix(),
			},
		})
	}
}

// CheckAndUpdateSESSDATA checks if SESSDATA is valid and sends notification if not.
func CheckAndUpdateSESSDATA(c *gin.Context) {
	cfg, err := config.LoadConfig()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load config"})
		return
	}

	if cfg.SESSDATA == "" || cfg.SESSDATA == "Cookie里的SESSDATA字段值" {
		c.JSON(http.StatusOK, gin.H{
			"valid":  false,
			"message": "SESSDATA not configured",
		})
		return
	}

	// Try to access a protected API to check if SESSDATA is valid
	apiURL := "https://api.bilibili.com/x/web-interface/nav"
	req, _ := http.NewRequest("GET", apiURL, nil)
	req.Header.Set("Cookie", "SESSDATA="+cfg.SESSDATA)
	req.Header.Set("User-Agent", userAgent)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"valid":  false,
			"message": "Failed to check SESSDATA: " + err.Error(),
		})
		return
	}
	defer resp.Body.Close()

	var result struct {
		Code int `json:"code"`
		Data struct {
			IsLogin bool `json:"isLogin"`
		} `json:"data"`
	}

	if err := parseJSON(resp.Body, &result); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"valid":  false,
			"message": "Invalid SESSDATA",
		})
		return
	}

	valid := result.Code == 0 && result.Data.IsLogin
	c.JSON(http.StatusOK, gin.H{
		"valid":     valid,
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

const userAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36"

type cookieKV struct {
	Name  string
	Value string
}

func extractCookies(resp *http.Response, redirectURL string, cookieInfoCookies []struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}, sessdata, biliJCT, dedeUserID, dedeUserIDCkMd5 string) map[string]string {
	cookies := make(map[string]string)
	cookieFields := []string{"SESSDATA", "bili_jct", "DedeUserID", "DedeUserID__ckMd5"}

	for _, c := range resp.Cookies() {
		for _, field := range cookieFields {
			if c.Name == field {
				cookies[field] = c.Value
				break
			}
		}
	}

	for _, v := range resp.Header.Values("Set-Cookie") {
		parts := splitCookieHeader(v)
		for name, value := range parts {
			for _, field := range cookieFields {
				if name == field && cookies[field] == "" {
					cookies[field] = value
					break
				}
			}
		}
	}

	for _, c := range cookieInfoCookies {
		for _, field := range cookieFields {
			if c.Name == field && cookies[field] == "" {
				cookies[field] = c.Value
				break
			}
		}
	}

	if redirectURL != "" && (cookies["SESSDATA"] == "" || cookies["bili_jct"] == "") {
		if idx := indexOf(redirectURL, "?"); idx >= 0 {
			query := redirectURL[idx+1:]
			if hashIdx := indexOf(query, "#"); hashIdx >= 0 {
				query = query[:hashIdx]
			}
			for _, param := range splitBy(query, "&") {
				eqIdx := indexOf(param, "=")
				if eqIdx >= 0 {
					name := param[:eqIdx]
					value := param[eqIdx+1:]
					for _, field := range cookieFields {
						if name == field && cookies[field] == "" {
							cookies[field] = value
							break
						}
					}
				}
			}
		}
	}

	if sessdata != "" && cookies["SESSDATA"] == "" {
		cookies["SESSDATA"] = sessdata
	}
	if biliJCT != "" && cookies["bili_jct"] == "" {
		cookies["bili_jct"] = biliJCT
	}
	if dedeUserID != "" && cookies["DedeUserID"] == "" {
		cookies["DedeUserID"] = dedeUserID
	}
	if dedeUserIDCkMd5 != "" && cookies["DedeUserID__ckMd5"] == "" {
		cookies["DedeUserID__ckMd5"] = dedeUserIDCkMd5
	}

	return cookies
}

func splitCookieHeader(header string) map[string]string {
	result := make(map[string]string)
	firstPart := header
	if idx := indexOf(firstPart, ";"); idx >= 0 {
		firstPart = firstPart[:idx]
	}
	eqIdx := indexOf(firstPart, "=")
	if eqIdx >= 0 {
		name := trimSpaceStr(firstPart[:eqIdx])
		value := firstPart[eqIdx+1:]
		result[name] = value
	}
	return result
}

func splitBy(s, sep string) []string {
	var result []string
	for len(s) > 0 {
		idx := indexOf(s, sep)
		if idx < 0 {
			result = append(result, s)
			break
		}
		result = append(result, s[:idx])
		s = s[idx+len(sep):]
	}
	return result
}

func trimSpaceStr(s string) string {
	start, end := 0, len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t') {
		end--
	}
	return s[start:end]
}
