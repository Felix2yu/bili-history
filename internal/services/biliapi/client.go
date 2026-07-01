package biliapi

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

const (
	userAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36"
	referer   = "https://www.bilibili.com/"
)

// Client is an HTTP client for Bilibili APIs.
type Client struct {
	httpClient *http.Client
	sessdata   string
	biliJCT    string
	dedeUserID string
}

// NewClient creates a new Bilibili API client.
func NewClient(sessdata, biliJCT, dedeUserID string) *Client {
	return &Client{
		httpClient: &http.Client{Timeout: 15 * time.Second},
		sessdata:   sessdata,
		biliJCT:    biliJCT,
		dedeUserID: dedeUserID,
	}
}

// SetCookies updates the authentication cookies.
func (c *Client) SetCookies(sessdata, biliJCT, dedeUserID string) {
	c.sessdata = sessdata
	c.biliJCT = biliJCT
	c.dedeUserID = dedeUserID
}

func (c *Client) newRequest(url string) (*http.Request, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Referer", referer)
	if c.sessdata != "" {
		req.Header.Set("Cookie", "SESSDATA="+c.sessdata)
	}
	return req, nil
}

func (c *Client) doRequest(url string) (map[string]interface{}, error) {
	req, err := c.newRequest(url)
	if err != nil {
		return nil, err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c *Client) doRequestRaw(url string) ([]byte, error) {
	req, err := c.newRequest(url)
	if err != nil {
		return nil, err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

// HistoryEntry represents a single entry from the history API.
type HistoryEntry struct {
	BVID      string `json:"bvid"`
	Title     string `json:"title"`
	Pic       string `json:"pic"`
	Desc      string `json:"desc"`
	Owner     struct {
		Mid  int64  `json:"mid"`
		Name string `json:"name"`
	} `json:"owner"`
	Duration int   `json:"duration"`
	ViewAt   int64 `json:"view_at"`
	Progress int   `json:"progress"`
	Repeat   int   `json:"repeat"`
	Cid      int64 `json:"cid"`
}

// HistoryResponse is the API response for history.
type HistoryResponse struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

// CursorData holds cursor info for paginated history.
type CursorData struct {
	Earliest int64 `json:"earliest"`
	IsEnd    bool  `json:"is_end"`
}

// FetchHistory fetches watch history from Bilibili.
// max: maximum number of entries per request (0 = default 20)
// viewAt: cursor timestamp, 0 = latest
func (c *Client) FetchHistory(max int, viewAt int64) ([]HistoryEntry, *CursorData, error) {
	url := "https://api.bilibili.com/x/web-interface/history/cursor"
	if max > 0 {
		url += "?max=" + strconv.Itoa(max)
	}
	if viewAt > 0 {
		url += "&view_at=" + strconv.FormatInt(viewAt, 10)
	}

	data, err := c.doRequest(url)
	if err != nil {
		return nil, nil, err
	}

	code, _ := data["code"].(float64)
	if code != 0 {
		msg, _ := data["message"].(string)
		return nil, nil, fmt.Errorf("API error code %v: %s", code, msg)
	}

	respData, _ := data["data"].(map[string]interface{})
	if respData == nil {
		return nil, nil, fmt.Errorf("empty data")
	}

	// Parse list
	listJSON, _ := json.Marshal(respData["list"])
	var entries []HistoryEntry
	if listJSON != nil {
		json.Unmarshal(listJSON, &entries)
	}

	// Parse cursor
	cursor := &CursorData{}
	cursorJSON, _ := json.Marshal(respData["cursor"])
	if cursorJSON != nil {
		json.Unmarshal(cursorJSON, cursor)
	}

	return entries, cursor, nil
}

// FetchAllHistory fetches all history by paginating through the cursor.
func (c *Client) FetchAllHistory(onPage func(entries []HistoryEntry, pageNum int)) ([]HistoryEntry, error) {
	var allEntries []HistoryEntry
	page := 0

	for {
		entries, cursor, err := c.FetchHistory(20, 0)
		if err != nil {
			return allEntries, err
		}
		if len(entries) == 0 {
			break
		}

		page++
		allEntries = append(allEntries, entries...)
		if onPage != nil {
			onPage(entries, page)
		}

		if cursor == nil || cursor.IsEnd {
			break
		}

		// Use earliest as cursor for next page
		entries2, cursor2, err := c.FetchHistory(20, cursor.Earliest)
		if err != nil {
			break
		}
		if len(entries2) == 0 {
			break
		}
		page++
		allEntries = append(allEntries, entries2...)
		if onPage != nil {
			onPage(entries2, page)
		}

		if cursor2 == nil || cursor2.IsEnd {
			break
		}
	}

	return allEntries, nil
}

// FetchHistoryByDate fetches history filtered by date range.
func (c *Client) FetchHistoryByDate(startTime, endTime int64) ([]HistoryEntry, error) {
	var result []HistoryEntry
	var viewAt int64

	for {
		entries, cursor, err := c.FetchHistory(20, viewAt)
		if err != nil {
			return result, err
		}
		if len(entries) == 0 {
			break
		}

		for _, entry := range entries {
			if entry.ViewAt >= startTime && entry.ViewAt <= endTime {
				result = append(result, entry)
			}
			// Stop if we've gone past our time range
			if entry.ViewAt < startTime {
				return result, nil
			}
		}

		if cursor == nil || cursor.IsEnd {
			break
		}
		viewAt = cursor.Earliest
	}

	return result, nil
}

// VideoInfo represents detailed video information.
type VideoInfo struct {
	BVID    string `json:"bvid"`
	Title   string `json:"title"`
	Desc    string `json:"desc"`
	Dur     int    `json:"duration"`
	PubDate int64  `json:"pubdate"`
	Stat    struct {
		View    int `json:"view"`
		Danmaku int `json:"danmaku"`
		Reply   int `json:"reply"`
		Fav     int `json:"favorite"`
		Coin    int `json:"coin"`
		Share   int `json:"share"`
		Like    int `json:"like"`
	} `json:"stat"`
	Owner struct {
		Mid  int64  `json:"mid"`
		Name string `json:"name"`
		Face string `json:"face"`
	} `json:"owner"`
	Tid  int    `json:"tid"`
	TName string `json:"tname"`
	Pic   string `json:"pic"`
}

// FetchVideoInfo fetches detailed video information by BVID.
func (c *Client) FetchVideoInfo(bvid string) (*VideoInfo, error) {
	url := fmt.Sprintf("https://api.bilibili.com/x/web-interface/view?bvid=%s", bvid)
	data, err := c.doRequest(url)
	if err != nil {
		return nil, err
	}

	code, _ := data["code"].(float64)
	if code != 0 {
		msg, _ := data["message"].(string)
		return nil, fmt.Errorf("API error: %s", msg)
	}

	dataJSON, _ := json.Marshal(data["data"])
	var info VideoInfo
	if err := json.Unmarshal(dataJSON, &info); err != nil {
		return nil, err
	}

	return &info, nil
}

// BatchFetchVideoInfo fetches video details for multiple BVIDs concurrently.
func (c *Client) BatchFetchVideoInfo(bvids []string, maxConcurrency int) ([]VideoInfo, error) {
	if maxConcurrency <= 0 {
		maxConcurrency = 10
	}

	type result struct {
		info *VideoInfo
		err  error
	}

	results := make(chan result, len(bvids))
	sem := make(chan struct{}, maxConcurrency)

	for _, bvid := range bvids {
		sem <- struct{}{}
		go func(b string) {
			defer func() { <-sem }()
			info, err := c.FetchVideoInfo(b)
			results <- result{info, err}
		}(bvid)
	}

	// Wait for all goroutines to finish
	for i := 0; i < maxConcurrency; i++ {
		sem <- struct{}{}
	}
	close(results)

	var infos []VideoInfo
	var lastErr error
	for r := range results {
		if r.err != nil {
			lastErr = r.err
			continue
		}
		if r.info != nil {
			infos = append(infos, *r.info)
		}
	}

	if len(infos) == 0 && lastErr != nil {
		return nil, lastErr
	}
	return infos, nil
}
