package biliapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const (
	BaseURL         = "https://api.bilibili.com"
	HistoryURL      = "https://api.bilibili.com/x/web-interface/history/cursor"
	LoginQrcodeURL  = "https://passport.bilibili.com/x/passport-login/web/qrcode/generate"
	QrcodePollURL   = "https://passport.bilibili.com/x/passport-login/web/qrcode/poll"
	PopularURL      = "https://api.bilibili.com/x/web-interface/popular"
	VideoInfoURL    = "https://api.bilibili.com/x/web-interface/view"
)

type Client struct {
	SESSDATA  string
	Buvid3    string
	UserAgent string
	client    *http.Client
}

type BiliResponse struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

type HistoryEntry struct {
	Title    string `json:"title"`
	LongTitle string `json:"long_title"`
	Cover    string `json:"cover"`
	URI      string `json:"uri"`
	History  HistoryInfo `json:"history"`
	ViewAt   int64  `json:"view_at"`
	Progress int    `json:"progress"`
	Badge    string `json:"badge"`
	ShowTitle string `json:"show_title"`
	Icon    string `json:"icon"`
	Business string `json:"business"`
	Bvid     string `json:"bvid"`
	DTotal   int    `json:"duration"`
}

type HistoryInfo struct {
	Bvid  string `json:"bvid"`
	Page  int    `json:"page"`
	Cid   int    `json:"cid"`
	Part  string `json:"part"`
	Business string `json:"business"`
	Dt    int    `json:"dt"`
}

type HistoryCursorData struct {
	Cursor HistoryCursor `json:"cursor"`
	List   []HistoryEntry `json:"list"`
}

type HistoryCursor struct {
	Max    int64 `json:"max"`
	ViewAt int64 `json:"view_at"`
	Business string `json:"business"`
	Ps     int   `json:"ps"`
}

type VideoInfo struct {
	Bvid     string  `json:"bvid"`
	Aid      int     `json:"aid"`
	Videos   int     `json:"videos"`
	Tid      int     `json:"tid"`
	Tname    string  `json:"tname"`
	Copyright int    `json:"copyright"`
	Pic      string  `json:"pic"`
	Title    string  `json:"title"`
	Pubdate  int64   `json:"pubdate"`
	Ctime    int64   `json:"ctime"`
	Desc     string  `json:"desc"`
	Duration int     `json:"duration"`
	Owner    VideoOwner `json:"owner"`
	Stat     VideoStat  `json:"stat"`
	Rights   VideoRights `json:"rights"`
}

type VideoOwner struct {
	Mid  int    `json:"mid"`
	Name string `json:"name"`
	Face string `json:"face"`
}

type VideoStat struct {
	View     int `json:"view"`
	Danmaku  int `json:"danmaku"`
	Reply    int `json:"reply"`
	Favorite int `json:"favorite"`
	Coin     int `json:"coin"`
	Share    int `json:"share"`
	Like     int `json:"like"`
}

type VideoRights struct {
	Bp          int `json:"bp"`
	Elec        int `json:"elec"`
	Download    int `json:"download"`
	Movie       int `json:"movie"`
	Pay         int `json:"pay"`
	Hd5         int `json:"hd5"`
	NoReprint   int `json:"no_reprint"`
	Autoplay    int `json:"autoplay"`
	UgcPay      int `json:"ugc_pay"`
	IsCooperation int `json:"is_cooperation"`
}

type PopularData struct {
	List   []PopularItem `json:"list"`
	NoMore bool          `json:"no_more"`
}

type PopularItem struct {
	Aid      int        `json:"aid"`
	Videos   int        `json:"videos"`
	Tid      int        `json:"tid"`
	Tname    string     `json:"tname"`
	Copyright int       `json:"copyright"`
	Pic      string     `json:"pic"`
	Title    string     `json:"title"`
	Pubdate  int64      `json:"pubdate"`
	Ctime    int64      `json:"ctime"`
	Desc     string     `json:"desc"`
	Duration int        `json:"duration"`
	Owner    VideoOwner `json:"owner"`
	Stat     VideoStat  `json:"stat"`
	Bvid     string     `json:"bvid"`
}

type QrCodeData struct {
	URL       string `json:"url"`
	QrcodeKey string `json:"qrcode_key"`
}

type QrCodePollData struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	URL     string `json:"url"`
}

func NewClient(sessdata string) *Client {
	return &Client{
		SESSDATA:  sessdata,
		Buvid3:    "random_string",
		UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *Client) getHeaders() map[string]string {
	headers := map[string]string{
		"User-Agent": c.UserAgent,
		"Referer":    "https://www.bilibili.com",
		"Origin":     "https://www.bilibili.com",
		"Accept":     "application/json, text/plain, */*",
	}
	if c.SESSDATA != "" {
		headers["Cookie"] = fmt.Sprintf("SESSDATA=%s; buvid3=%s; b_nut=1234567890; buvid4=random_string", c.SESSDATA, c.Buvid3)
	}
	return headers
}

func (c *Client) Get(urlStr string, params map[string]string) ([]byte, error) {
	u, err := url.Parse(urlStr)
	if err != nil {
		return nil, fmt.Errorf("parse url error: %w", err)
	}

	if params != nil {
		q := u.Query()
		for k, v := range params {
			q.Set(k, v)
		}
		u.RawQuery = q.Encode()
	}

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("create request error: %w", err)
	}

	headers := c.getHeaders()
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request error: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body error: %w", err)
	}

	return body, nil
}

func (c *Client) Post(urlStr string, data interface{}) ([]byte, error) {
	var body io.Reader
	if data != nil {
		jsonData, err := json.Marshal(data)
		if err != nil {
			return nil, fmt.Errorf("marshal data error: %w", err)
		}
		body = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest("POST", urlStr, body)
	if err != nil {
		return nil, fmt.Errorf("create request error: %w", err)
	}

	headers := c.getHeaders()
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request error: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body error: %w", err)
	}

	return respBody, nil
}

func (c *Client) GetHistory(max int64, viewAt int64, ps int) (*HistoryCursorData, error) {
	params := map[string]string{
		"ps":       fmt.Sprintf("%d", ps),
		"max":      "",
		"view_at":  "",
		"business": "",
	}
	if max > 0 {
		params["max"] = fmt.Sprintf("%d", max)
	}
	if viewAt > 0 {
		params["view_at"] = fmt.Sprintf("%d", viewAt)
	}

	body, err := c.Get(HistoryURL, params)
	if err != nil {
		return nil, err
	}

	var resp BiliResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal response error: %w", err)
	}

	if resp.Code != 0 {
		return nil, fmt.Errorf("api error: code=%d, message=%s", resp.Code, resp.Message)
	}

	var data HistoryCursorData
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		return nil, fmt.Errorf("unmarshal data error: %w", err)
	}

	return &data, nil
}

func (c *Client) GetVideoInfo(bvid string) (*VideoInfo, error) {
	params := map[string]string{
		"bvid": bvid,
	}

	body, err := c.Get(VideoInfoURL, params)
	if err != nil {
		return nil, err
	}

	var resp BiliResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal response error: %w", err)
	}

	if resp.Code != 0 {
		return nil, fmt.Errorf("api error: code=%d, message=%s", resp.Code, resp.Message)
	}

	var data VideoInfo
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		return nil, fmt.Errorf("unmarshal data error: %w", err)
	}

	return &data, nil
}

func (c *Client) GetPopular(pn int, ps int) (*PopularData, error) {
	params := map[string]string{
		"pn": fmt.Sprintf("%d", pn),
		"ps": fmt.Sprintf("%d", ps),
	}

	body, err := c.Get(PopularURL, params)
	if err != nil {
		return nil, err
	}

	var resp BiliResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal response error: %w", err)
	}

	if resp.Code != 0 {
		return nil, fmt.Errorf("api error: code=%d, message=%s", resp.Code, resp.Message)
	}

	var data PopularData
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		return nil, fmt.Errorf("unmarshal data error: %w", err)
	}

	return &data, nil
}

func GenerateQrcode() (*QrCodeData, error) {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	req, err := http.NewRequest("GET", LoginQrcodeURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request error: %w", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Referer", "https://passport.bilibili.com/login")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request error: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body error: %w", err)
	}

	var biliResp BiliResponse
	if err := json.Unmarshal(body, &biliResp); err != nil {
		return nil, fmt.Errorf("unmarshal response error: %w", err)
	}

	if biliResp.Code != 0 {
		return nil, fmt.Errorf("api error: code=%d, message=%s", biliResp.Code, biliResp.Message)
	}

	var data QrCodeData
	if err := json.Unmarshal(biliResp.Data, &data); err != nil {
		return nil, fmt.Errorf("unmarshal data error: %w", err)
	}

	return &data, nil
}

func PollQrcode(qrcodeKey string) (*QrCodePollData, []*http.Cookie, error) {
	client := &http.Client{
		Timeout: 30 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	params := url.Values{}
	params.Set("qrcode_key", qrcodeKey)
	urlStr := QrcodePollURL + "?" + params.Encode()

	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("create request error: %w", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Referer", "https://passport.bilibili.com/login")

	resp, err := client.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("request error: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("read body error: %w", err)
	}

	var biliResp BiliResponse
	if err := json.Unmarshal(body, &biliResp); err != nil {
		return nil, nil, fmt.Errorf("unmarshal response error: %w", err)
	}

	var data QrCodePollData
	if len(biliResp.Data) > 0 {
		if err := json.Unmarshal(biliResp.Data, &data); err != nil {
			return nil, nil, fmt.Errorf("unmarshal data error: %w", err)
		}
	} else {
		data.Code = biliResp.Code
		data.Message = biliResp.Message
	}

	cookies := resp.Cookies()

	return &data, cookies, nil
}
