package handlers

import (
	"net/http"
	"time"

	"bili-history/internal/config"
	"bili-history/internal/models"

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

	// Generate QR code PNG
	png, err := qrcode.Encode(result.Data.URL, qrcode.Medium, 256)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to encode QR code"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"url":     result.Data.URL,
		"key":     result.Data.Key,
		"img_url": result.Data.ImgURL,
		"qr_png":  png,
	})
}

// PollLogin polls the QR code login status.
func PollLogin(c *gin.Context) {
	key := c.Query("key")
	if key == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "key is required"})
		return
	}

	apiURL := "https://passport.bilibili.com/x/passport-login/web/qrcode/poll?qrcode_key=" + key

	resp, err := http.Get(apiURL)
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
		} `json:"data"`
	}

	if err := parseJSON(resp.Body, &result); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse response"})
		return
	}

	switch result.Data.Code {
	case 0: // Success
		cookies := map[string]string{
			"SESSDATA":        result.Data.SESSDATA,
			"bili_jct":        result.Data.BiliJCT,
			"DedeUserID":      result.Data.DedeUserID,
			"DedeUserID__ckMd5": result.Data.DedeUserIDCkMd5,
		}
		// Save to config
		if err := config.SaveCookies(cookies); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save cookies"})
			return
		}
		// Reload config
		config.ReloadConfig()

		c.JSON(http.StatusOK, models.LoginPollResponse{
			Status:  "success",
			Message: "Login successful",
			Cookies: cookies,
		})
	case 86090: // Already scanned, waiting for confirmation
		c.JSON(http.StatusOK, models.LoginPollResponse{
			Status:  "pending",
			Message: "Already scanned, waiting for confirmation",
		})
	case 86038: // QR code expired
		c.JSON(http.StatusOK, models.LoginPollResponse{
			Status:  "expired",
			Message: "QR code expired",
		})
	default:
		c.JSON(http.StatusOK, models.LoginPollResponse{
			Status:  "pending",
			Message: result.Data.Message,
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
