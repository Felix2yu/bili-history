package biliapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	HistoryURL              = "https://api.bilibili.com/x/web-interface/history/cursor"
	VideoInfoURL            = "https://api.bilibili.com/x/web-interface/view"
	WatchLaterURL           = "https://api.bilibili.com/x/v2/history/toview"
	WatchLaterDelURL        = "https://api.bilibili.com/x/v2/history/toview/del"
	DynamicSpaceURL         = "https://api.bilibili.com/x/polymer/web-dynamic/v1/feed/space"
	FavoriteFolderListURL   = "https://api.bilibili.com/x/v3/fav/folder/created/list-all"
	FavoriteResourceListURL = "https://api.bilibili.com/x/v3/fav/resource/list"
	FavoriteDealURL         = "https://api.bilibili.com/x/v3/fav/resource/deal"
	LikedVideoURL           = "https://api.bilibili.com/x/v2/liked/web/video"
	LikeURL                 = "https://api.bilibili.com/x/web-interface/archive/like"
)

type Client struct {
	SESSDATA   string
	BiliJct    string
	DedeUserID string
	Buvid3     string
	UserAgent  string
	client     *http.Client
}

type BiliResponse struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

type HistoryEntry struct {
	Title      string `json:"title"`
	LongTitle  string `json:"long_title"`
	Cover      string `json:"cover"`
	URI        string `json:"uri"`
	History    HistoryInfo `json:"history"`
	ViewAt     int64  `json:"view_at"`
	Progress   int    `json:"progress"`
	Badge      string `json:"badge"`
	ShowTitle  string `json:"show_title"`
	Icon       string `json:"icon"`
	Business   string `json:"business"`
	Bvid       string `json:"bvid"`
	DTotal     int    `json:"duration"`
	AuthorName string `json:"author_name"`
	AuthorFace string `json:"author_face"`
	AuthorMid  int64  `json:"author_mid"`
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

// NewClientWithConfig constructs a client with full credentials (SESSDATA + bili_jct + DedeUserID)
// from the application config, required for write operations like removing watch later items.
func NewClientWithConfig(sessdata, biliJct, dedeUserID string) *Client {
	c := NewClient(sessdata)
	c.BiliJct = biliJct
	c.DedeUserID = dedeUserID
	return c
}

func (c *Client) getHeaders() map[string]string {
	headers := map[string]string{
		"User-Agent": c.UserAgent,
		"Referer":    "https://www.bilibili.com",
		"Origin":     "https://www.bilibili.com",
		"Accept":     "application/json, text/plain, */*",
	}
	if c.SESSDATA != "" {
		cookies := []string{
			fmt.Sprintf("SESSDATA=%s", c.SESSDATA),
			fmt.Sprintf("buvid3=%s", c.Buvid3),
			"b_nut=1234567890",
			"buvid4=random_string",
		}
		if c.BiliJct != "" {
			cookies = append(cookies, fmt.Sprintf("bili_jct=%s", c.BiliJct))
		}
		if c.DedeUserID != "" {
			cookies = append(cookies, fmt.Sprintf("DedeUserID=%s", c.DedeUserID))
		}
		headers["Cookie"] = strings.Join(cookies, "; ")
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

// WatchLaterItem represents a single entry in the Bilibili watch later list.
type WatchLaterItem struct {
	Aid      int64           `json:"aid"`
	Bvid     string          `json:"bvid"`
	Title    string          `json:"title"`
	Pic      string          `json:"pic"`
	Desc     string          `json:"desc"`
	Duration int             `json:"duration"`
	Tid      int             `json:"tid"`
	Tname    string          `json:"tname"`
	Owner    VideoOwner      `json:"owner"`
	Stat     VideoStat       `json:"stat"`
	AddAt    int64           `json:"add_at"`
	Pubdate  int64           `json:"pubdate"`
}

// WatchLaterData is the raw payload returned by /x/v2/history/toview.
type WatchLaterData struct {
	Count int               `json:"count"`
	List  []WatchLaterItem  `json:"list"`
}

// GetWatchLaterList fetches the user's full watch later list from Bilibili.
// The official API returns up to 1000 items in a single response.
func (c *Client) GetWatchLaterList() (*WatchLaterData, error) {
	body, err := c.Get(WatchLaterURL, nil)
	if err != nil {
		return nil, err
	}

	var resp BiliResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal response error: %w", err)
	}
	if resp.Code != 0 {
		return nil, &ApiError{Code: resp.Code, Message: resp.Message}
	}

	var data WatchLaterData
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		return nil, fmt.Errorf("unmarshal data error: %w", err)
	}
	return &data, nil
}

// PostForm sends a POST request with application/x-www-form-urlencoded body,
// which is required by Bilibili write APIs such as removing a watch later item.
func (c *Client) PostForm(urlStr string, form url.Values) ([]byte, error) {
	req, err := http.NewRequest("POST", urlStr, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, fmt.Errorf("create request error: %w", err)
	}
	headers := c.getHeaders()
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

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

// RemoveFromWatchLater removes a single video from the watch later list using its aid.
// The Bilibili API requires the bili_jct (csrf) token.
func (c *Client) RemoveFromWatchLater(aid int64) error {
	if c.BiliJct == "" {
		return fmt.Errorf("bili_jct (csrf) is required to remove watch later items")
	}
	form := url.Values{}
	form.Set("aid", fmt.Sprintf("%d", aid))
	form.Set("platform", "web")
	form.Set("csrf", c.BiliJct)

	body, err := c.PostForm(WatchLaterDelURL, form)
	if err != nil {
		return err
	}
	var resp BiliResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return fmt.Errorf("unmarshal response error: %w", err)
	}
	if resp.Code != 0 {
		return &ApiError{Code: resp.Code, Message: resp.Message}
	}
	return nil
}

// ApiError represents a Bilibili API-level error with its code and message.
type ApiError struct {
	Code    int
	Message string
}

func (e *ApiError) Error() string {
	return fmt.Sprintf("api error: code=%d, message=%s", e.Code, e.Message)
}

type DynamicRawItem struct {
	IDStr   string          `json:"id_str"`
	Type    string          `json:"type"`
	Modules json.RawMessage `json:"modules"`
}

type DynamicSpaceResponse struct {
	HasMore bool             `json:"has_more"`
	Offset  string           `json:"offset"`
	Items   []DynamicRawItem `json:"items"`
}

func (c *Client) GetDynamicList(hostMid string, offset string, ps int) (*DynamicSpaceResponse, error) {
	params := map[string]string{
		"host_mid": hostMid,
		"ps":       fmt.Sprintf("%d", ps),
	}
	if offset != "" {
		params["offset"] = offset
	}

	body, err := c.Get(DynamicSpaceURL, params)
	if err != nil {
		return nil, err
	}

	var resp BiliResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal response error: %w", err)
	}

	if resp.Code != 0 {
		return nil, &ApiError{Code: resp.Code, Message: resp.Message}
	}

	var result DynamicSpaceResponse
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("unmarshal data error: %w", err)
	}

	return &result, nil
}

// ===== Favorites =====

type FavFolderInfo struct {
	ID         int64  `json:"id"`
	Fid        int64  `json:"fid"`
	Mid        int64  `json:"mid"`
	Title      string `json:"title"`
	Cover      string `json:"cover"`
	Attr       int    `json:"attr"`
	Intro      string `json:"intro"`
	Ctime      int64  `json:"ctime"`
	Mtime      int64  `json:"mtime"`
	State      int    `json:"state"`
	MediaCount int    `json:"media_count"`
	FavState   int    `json:"fav_state"`
	LikeState  int    `json:"like_state"`
}

type FavFolderListData struct {
	Count  int             `json:"count"`
	List   []FavFolderInfo `json:"list"`
}

func (c *Client) GetFavoriteFolderList() (*FavFolderListData, error) {
	params := map[string]string{
		"up_mid": c.DedeUserID,
	}
	body, err := c.Get(FavoriteFolderListURL, params)
	if err != nil {
		return nil, err
	}
	var resp BiliResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal response error: %w", err)
	}
	if resp.Code != 0 {
		return nil, &ApiError{Code: resp.Code, Message: resp.Message}
	}
	var data FavFolderListData
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		return nil, fmt.Errorf("unmarshal data error: %w", err)
	}
	return &data, nil
}

type FavResourceItem struct {
	ID       int64  `json:"id"`
	Type     int    `json:"type"`
	Title    string `json:"title"`
	Cover    string `json:"cover"`
	Intro    string `json:"intro"`
	Page     int    `json:"page"`
	Duration int    `json:"duration"`
	Upper    FavUpper  `json:"upper"`
	Ctime    int64  `json:"ctime"`
	Pubtime  int64  `json:"pubtime"`
	FavTime  int64  `json:"fav_time"`
	Attr     int    `json:"attr"`
	UGC      *FavUGC `json:"ugc,omitempty"`
	Stat     *FavStat `json:"stat,omitempty"`
	Cid      int64  `json:"cid"`
}

type FavUpper struct {
	Mid  int64  `json:"mid"`
	Name string `json:"name"`
	Face string `json:"face"`
}

type FavUGC struct {
	Bvid string `json:"bvid"`
}

type FavStat struct {
	View     int `json:"view"`
	Danmaku  int `json:"danmaku"`
	Reply    int `json:"reply"`
	Favorite int `json:"favorite"`
	Coin     int `json:"coin"`
	Share    int `json:"share"`
	Like     int `json:"like"`
}

type FavResourcePage struct {
	Num   int `json:"num"`
	Size  int `json:"size"`
	Count int `json:"count"`
	Total int `json:"total"`
}

type FavResourceData struct {
	Info   *FavFolderInfo    `json:"info"`
	Media  []FavResourceItem `json:"medias"`
	Page   FavResourcePage   `json:"page"`
}

func (c *Client) GetFavoriteResources(mediaID int64, pn, ps int) (*FavResourceData, error) {
	params := map[string]string{
		"media_id": fmt.Sprintf("%d", mediaID),
		"pn":       fmt.Sprintf("%d", pn),
		"ps":       fmt.Sprintf("%d", ps),
		"order":    "fav_time",
	}
	body, err := c.Get(FavoriteResourceListURL, params)
	if err != nil {
		return nil, err
	}
	var resp BiliResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal response error: %w", err)
	}
	if resp.Code != 0 {
		return nil, &ApiError{Code: resp.Code, Message: resp.Message}
	}
	var data FavResourceData
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		return nil, fmt.Errorf("unmarshal data error: %w", err)
	}
	return &data, nil
}

func (c *Client) DealFavoriteResource(resources string, mediaIDs string) error {
	if c.BiliJct == "" {
		return fmt.Errorf("bili_jct (csrf) is required for favorite operations")
	}
	form := url.Values{}
	form.Set("resources", resources)
	form.Set("media_ids", mediaIDs)
	form.Set("csrf", c.BiliJct)

	body, err := c.PostForm(FavoriteDealURL, form)
	if err != nil {
		return err
	}
	var resp BiliResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return fmt.Errorf("unmarshal response error: %w", err)
	}
	if resp.Code != 0 {
		return &ApiError{Code: resp.Code, Message: resp.Message}
	}
	return nil
}

// ===== Likes =====

type LikedVideoItem struct {
	ID       int64  `json:"id"`
	Title    string `json:"title"`
	Pic      string `json:"pic"`
	Desc     string `json:"desc"`
	Duration int    `json:"duration"`
	Tid      int    `json:"tid"`
	Tname    string `json:"tname"`
	Owner    VideoOwner `json:"owner"`
	Stat     VideoStat  `json:"stat"`
	Pubdate  int64  `json:"pubdate"`
	Bvid     string `json:"bvid"`
	Aid      int64  `json:"aid"`
}

type LikedVideoPage struct {
	Total int `json:"total"`
}

type LikedVideoData struct {
	List  []LikedVideoItem `json:"list"`
	Page  LikedVideoPage   `json:"page"`
}

func (c *Client) GetLikedVideos(vmid int64, pn, ps int) (*LikedVideoData, error) {
	params := map[string]string{
		"vmid": fmt.Sprintf("%d", vmid),
		"pn":   fmt.Sprintf("%d", pn),
		"ps":   fmt.Sprintf("%d", ps),
	}
	body, err := c.Get(LikedVideoURL, params)
	if err != nil {
		return nil, err
	}
	var resp BiliResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal response error: %w", err)
	}
	if resp.Code != 0 {
		return nil, &ApiError{Code: resp.Code, Message: resp.Message}
	}
	var data LikedVideoData
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		return nil, fmt.Errorf("unmarshal data error: %w", err)
	}
	return &data, nil
}

func (c *Client) LikeVideo(bvid string, like bool) error {
	if c.BiliJct == "" {
		return fmt.Errorf("bili_jct (csrf) is required for like operations")
	}
	form := url.Values{}
	form.Set("bvid", bvid)
	if like {
		form.Set("like", "1")
	} else {
		form.Set("like", "0")
	}
	form.Set("csrf", c.BiliJct)

	body, err := c.PostForm(LikeURL, form)
	if err != nil {
		return err
	}
	var resp BiliResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return fmt.Errorf("unmarshal response error: %w", err)
	}
	if resp.Code != 0 {
		return &ApiError{Code: resp.Code, Message: resp.Message}
	}
	return nil
}
