package wbi

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

// MixinKey encoding table (same as Python version).
var mixinKeyEncTab = [64]int{
	46, 47, 18, 2, 53, 8, 23, 32, 15, 50, 10, 31, 58, 3, 45, 35,
	27, 43, 5, 49, 33, 9, 42, 19, 29, 28, 14, 39, 12, 38, 41, 13,
	37, 48, 7, 16, 24, 55, 40, 61, 26, 17, 0, 1, 60, 51, 30, 4,
	22, 25, 54, 21, 56, 59, 6, 63, 57, 62, 11, 36, 20, 34, 44, 52,
}

type cachedKeys struct {
	imgKey    string
	subKey    string
	lastFetch time.Time
}

var (
	cache    cachedKeys
	cacheMu  sync.RWMutex
	cacheTTL = 1 * time.Hour
)

// GetMixinKey derives the mixin key from imgKey + subKey.
func GetMixinKey(orig string) string {
	var b strings.Builder
	for _, i := range mixinKeyEncTab {
		if i < len(orig) {
			b.WriteByte(orig[i])
		}
	}
	s := b.String()
	if len(s) > 32 {
		s = s[:32]
	}
	return s
}

// FetchWBIKeys fetches the latest WBI signing keys from Bilibili API.
func FetchWBIKeys(sessdata string) (imgKey, subKey string, err error) {
	cacheMu.RLock()
	if cache.imgKey != "" && cache.subKey != "" && time.Since(cache.lastFetch) < cacheTTL {
		imgKey, subKey = cache.imgKey, cache.subKey
		cacheMu.RUnlock()
		return
	}
	cacheMu.RUnlock()

	cacheMu.Lock()
	defer cacheMu.Unlock()

	// Double-check
	if cache.imgKey != "" && cache.subKey != "" && time.Since(cache.lastFetch) < cacheTTL {
		return cache.imgKey, cache.subKey, nil
	}

	req, err := http.NewRequest("GET", "https://api.bilibili.com/x/web-interface/nav", nil)
	if err != nil {
		return "", "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")
	req.Header.Set("Referer", "https://www.bilibili.com/")
	if sessdata != "" {
		req.Header.Set("Cookie", "SESSDATA="+sessdata)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("failed to fetch nav info: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", fmt.Errorf("failed to read response: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", "", fmt.Errorf("failed to parse response: %w", err)
	}

	code, _ := result["code"].(float64)
	if code != 0 {
		msg, _ := result["message"].(string)
		return "", "", fmt.Errorf("API error: %s", msg)
	}

	data, _ := result["data"].(map[string]interface{})
	wbiImg, _ := data["wbi_img"].(map[string]interface{})
	imgURL, _ := wbiImg["img_url"].(string)
	subURL, _ := wbiImg["sub_url"].(string)

	if imgURL == "" || subURL == "" {
		return "", "", fmt.Errorf("wbi_img or sub_url is empty")
	}

	imgParts := strings.Split(imgURL, "/")
	subParts := strings.Split(subURL, "/")

	imgKey = strings.Split(imgParts[len(imgParts)-1], ".")[0]
	subKey = strings.Split(subParts[len(subParts)-1], ".")[0]

	cache = cachedKeys{
		imgKey:    imgKey,
		subKey:    subKey,
		lastFetch: time.Now(),
	}

	return imgKey, subKey, nil
}

// SignParams signs the given parameters using WBI.
func SignParams(params map[string]string, imgKey, subKey string) map[string]string {
	mixinKey := GetMixinKey(imgKey + subKey)

	params["wts"] = strconv.FormatInt(time.Now().Unix(), 10)

	// Sort keys
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Filter values and build query
	var parts []string
	for _, k := range keys {
		v := params[k]
		v = strings.NewReplacer("!", "", "'", "", "(", "", ")", "", "*", "").Replace(v)
		parts = append(parts, url.QueryEscape(k)+"="+url.QueryEscape(v))
	}
	query := strings.Join(parts, "&")

	// Calculate w_rid
	hash := md5.Sum([]byte(query + mixinKey))
	wRid := fmt.Sprintf("%x", hash)

	result := make(map[string]string, len(params)+1)
	for k, v := range params {
		result[k] = v
	}
	result["w_rid"] = wRid

	return result
}
