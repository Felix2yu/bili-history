package biliapi

import (
	"compress/gzip"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
)

const (
	HistoryURL              = "https://api.bilibili.com/x/web-interface/history/cursor"
	HistoryDelURL           = "https://api.bilibili.com/x/web-interface/history/del"
	VideoInfoURL            = "https://api.bilibili.com/x/web-interface/view"
	WatchLaterURL           = "https://api.bilibili.com/x/v2/history/toview"
	WatchLaterDelURL        = "https://api.bilibili.com/x/v2/history/toview/del"
	DynamicSpaceURL         = "https://api.bilibili.com/x/polymer/web-dynamic/v1/feed/space"
	UserCardURL             = "https://api.bilibili.com/x/web-interface/card"
	FavoriteFolderListURL        = "https://api.bilibili.com/x/v3/fav/folder/created/list-all"
	FavoriteCollectedListURL     = "https://api.bilibili.com/x/v3/fav/folder/collected/list"
	FavoriteResourceListURL = "https://api.bilibili.com/x/v3/fav/resource/list"
	FavoriteSeasonListURL   = "https://api.bilibili.com/x/space/fav/season/list"
	FavoriteDealURL         = "https://api.bilibili.com/x/v3/fav/resource/deal"
	LikedVideoURL           = "https://api.bilibili.com/x/space/like/video"
	LikeURL                 = "https://api.bilibili.com/x/web-interface/archive/like"
	WbiNavURL               = "https://api.bilibili.com/x/web-interface/nav"
)

// WBI mixin key 的混淆表
var mixinKeyEncTab = []int{
	46, 47, 18, 2, 53, 8, 23, 32, 15, 50, 10, 31, 58, 3, 45, 35, 27, 43, 5, 49,
	33, 9, 42, 19, 29, 28, 14, 39, 12, 38, 41, 13, 37, 48, 7, 16, 24, 55, 40,
	61, 26, 17, 0, 1, 60, 51, 30, 4, 22, 25, 54, 21, 56, 59, 6, 63, 57, 62, 11,
	36, 20, 34, 44, 52,
}

// WBI 签名
func getMixinKey(orig string) string {
	var buf strings.Builder
	for _, v := range mixinKeyEncTab {
		if v < len(orig) {
			buf.WriteByte(orig[v])
		}
	}
	return buf.String()[:32]
}

// WbiSign 对参数进行 wbi 签名
func wbiSign(params map[string]string, imgKey, subKey string) string {
	mixinKey := getMixinKey(imgKey + subKey)
	currTime := fmt.Sprintf("%d", time.Now().Unix())
	params["wts"] = currTime

	// 按 key 排序
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// 过滤特殊字符并拼接
	var buf strings.Builder
	for i, k := range keys {
		v := params[k]
		v = strings.ReplaceAll(v, "'", "")
		v = strings.ReplaceAll(v, "(", "")
		v = strings.ReplaceAll(v, ")", "")
		v = strings.ReplaceAll(v, "!", "")
		if i > 0 {
			buf.WriteByte('&')
		}
		buf.WriteString(url.QueryEscape(k))
		buf.WriteByte('=')
		buf.WriteString(url.QueryEscape(v))
	}

	// 计算 md5
	hash := md5.Sum([]byte(buf.String() + mixinKey))
	return fmt.Sprintf("%x", hash)
}

// generateDmImgList 生成 dm_img_list 参数
func generateDmImgList() string {
	x := 1245 + rand.Intn(10) - 5
	if x < 0 {
		x = 0
	}
	y := 1285 + rand.Intn(10) - 5
	if y < 0 {
		y = 0
	}
	timestamp := 30 + rand.Intn(10) - 5
	if timestamp < 0 {
		timestamp = 0
	}
	return fmt.Sprintf(`[{"x":%d,"y":%d,"z":0,"timestamp":%d,"type":0}]`,
		3*x+2*y, 4*x-5*y, timestamp)
}

// addDmVerifyInfo 添加反爬验证参数
func addDmVerifyInfo(params string) string {
	dmImgList := generateDmImgList()
	// base64 编码 "no webgl"
	dmImgStr := "bm8gd2ViZ2w"
	dmCoverImgStr := "bm8gd2ViZ2w"
	return fmt.Sprintf("%s&dm_img_list=%s&dm_img_str=%s&dm_cover_img_str=%s",
		params, url.QueryEscape(dmImgList), dmImgStr, dmCoverImgStr)
}

type Client struct {
	SESSDATA   string
	BiliJct    string
	DedeUserID string
	Buvid3     string
	UserAgent  string
	ImgKey     string
	SubKey     string
	client     *http.Client
}

type BiliResponse struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

type HistoryEntry struct {
	Title      string      `json:"title"`
	LongTitle  string      `json:"long_title"`
	Cover      string      `json:"cover"`
	URI        string      `json:"uri"`
	History    HistoryInfo `json:"history"`
	ViewAt     int64       `json:"view_at"`
	Progress   int         `json:"progress"`
	Badge      string      `json:"badge"`
	ShowTitle  string      `json:"show_title"`
	Icon       string      `json:"icon"`
	Business   string      `json:"business"`
	Bvid       string      `json:"bvid"`
	DTotal     int         `json:"duration"`
	AuthorName string      `json:"author_name"`
	AuthorFace string      `json:"author_face"`
	AuthorMid  int64       `json:"author_mid"`
}

type HistoryInfo struct {
	Bvid     string `json:"bvid"`
	Page     int    `json:"page"`
	Cid      int    `json:"cid"`
	Part     string `json:"part"`
	Business string `json:"business"`
	Dt       int    `json:"dt"`
}

type HistoryCursorData struct {
	Cursor HistoryCursor  `json:"cursor"`
	List   []HistoryEntry `json:"list"`
}

type HistoryCursor struct {
	Max      int64  `json:"max"`
	ViewAt   int64  `json:"view_at"`
	Business string `json:"business"`
	Ps       int    `json:"ps"`
}

type VideoInfo struct {
	Bvid      string      `json:"bvid"`
	Aid       int         `json:"aid"`
	Videos    int         `json:"videos"`
	Tid       int         `json:"tid"`
	Tname     string      `json:"tname"`
	Copyright int         `json:"copyright"`
	Pic       string      `json:"pic"`
	Title     string      `json:"title"`
	Pubdate   int64       `json:"pubdate"`
	Ctime     int64       `json:"ctime"`
	Desc      string      `json:"desc"`
	Duration  int         `json:"duration"`
	Owner     VideoOwner  `json:"owner"`
	Stat      VideoStat   `json:"stat"`
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

// generateBuvid3 生成 buvid3 UUID (格式: xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx)
func generateBuvid3() string {
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		time.Now().UnixNano()%0x100000000,
		time.Now().UnixNano()%0x10000,
		time.Now().UnixNano()%0x10000,
		time.Now().UnixNano()%0x10000,
		time.Now().UnixNano()%0x100000000000)
}

func NewClient(sessdata string) *Client {
	buvid3 := generateBuvid3()
	return &Client{
		SESSDATA:  sessdata,
		Buvid3:    buvid3,
		UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36",
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

// FetchWbiKeys 从B站获取 wbi 签名所需的 img_key 和 sub_key
func (c *Client) FetchWbiKeys() error {
	body, err := c.GetWithDm(WbiNavURL, nil)
	if err != nil {
		return fmt.Errorf("fetch wbi keys error: %w", err)
	}

	var resp struct {
		Code int `json:"code"`
		Data struct {
			WbiImg struct {
				ImgURL string `json:"img_url"`
				SubURL string `json:"sub_url"`
			} `json:"wbi_img"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return fmt.Errorf("unmarshal wbi keys error: %w", err)
	}
	if resp.Code != 0 {
		return fmt.Errorf("fetch wbi keys failed: code=%d", resp.Code)
	}

	// 从 URL 中提取 key (格式: https://i0.hdslb.com/bfs/wbi/xxx.png -> xxx)
	imgURL := resp.Data.WbiImg.ImgURL
	subURL := resp.Data.WbiImg.SubURL

	// 提取文件名（去掉扩展名）作为 key
	imgKey := imgURL[strings.LastIndex(imgURL, "/")+1:]
	imgKey = imgKey[:strings.LastIndex(imgKey, ".")]
	subKey := subURL[strings.LastIndex(subURL, "/")+1:]
	subKey = subKey[:strings.LastIndex(subKey, ".")]

	c.ImgKey = imgKey
	c.SubKey = subKey
	return nil
}

// SignWbi 对参数进行 wbi 签名（如果已获取到 keys）
func (c *Client) SignWbi(params map[string]string) string {
	if c.ImgKey == "" || c.SubKey == "" {
		// 尝试获取 keys
		c.FetchWbiKeys()
	}
	if c.ImgKey != "" && c.SubKey != "" {
		return wbiSign(params, c.ImgKey, c.SubKey)
	}
	return ""
}

func (c *Client) getHeaders() map[string]string {
	headers := map[string]string{
		"User-Agent":         c.UserAgent,
		"Referer":            "https://www.bilibili.com",
		"Origin":             "https://www.bilibili.com",
		"Accept":             "application/json, text/plain, */*",
		"Accept-Language":    "zh-CN,zh;q=0.9,en;q=0.8",
		"Accept-Encoding":    "gzip, deflate",
		"Connection":         "keep-alive",
		"Sec-Ch-Ua":          `"Google Chrome";v="131", "Chromium";v="131", "Not_A Brand";v="24"`,
		"Sec-Ch-Ua-Mobile":   "?0",
		"Sec-Ch-Ua-Platform": `"Windows"`,
		"Sec-Fetch-Dest":     "empty",
		"Sec-Fetch-Mode":     "cors",
		"Sec-Fetch-Site":     "same-site",
	}
	if c.SESSDATA != "" {
		// 生成 buvid4
		buvid4 := generateBuvid3()
		// 生成 b_lsid
		blsid := fmt.Sprintf("%08x_%010x", time.Now().UnixNano()%0x100000000, time.Now().UnixNano()%0x10000000000)
		cookies := []string{
			fmt.Sprintf("SESSDATA=%s", c.SESSDATA),
			fmt.Sprintf("buvid3=%s", c.Buvid3),
			fmt.Sprintf("buvid4=%s", buvid4),
			fmt.Sprintf("b_lsid=%s", blsid),
			"b_nut=1234567890",
			"bili_ticket=",
			"bili_ticket_mid=",
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

	// 处理 gzip 压缩
	var reader io.Reader = resp.Body
	if resp.Header.Get("Content-Encoding") == "gzip" {
		gz, err := gzip.NewReader(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("gzip reader error: %w", err)
		}
		defer gz.Close()
		reader = gz
	}

	body, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("read body error: %w", err)
	}

	return body, nil
}

// GetWithDm 发送 GET 请求并添加 dm 反爬验证参数
func (c *Client) GetWithDm(urlStr string, params map[string]string) ([]byte, error) {
	u, err := url.Parse(urlStr)
	if err != nil {
		return nil, fmt.Errorf("parse url error: %w", err)
	}

	// 构建查询字符串
	queryParts := []string{}
	if params != nil {
		for k, v := range params {
			queryParts = append(queryParts, fmt.Sprintf("%s=%s", url.QueryEscape(k), url.QueryEscape(v)))
		}
	}
	queryStr := strings.Join(queryParts, "&")

	// 添加 dm 验证参数
	queryStr = addDmVerifyInfo(queryStr)
	u.RawQuery = queryStr

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

	// 处理 gzip 压缩
	var reader io.Reader = resp.Body
	if resp.Header.Get("Content-Encoding") == "gzip" {
		gz, err := gzip.NewReader(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("gzip reader error: %w", err)
		}
		defer gz.Close()
		reader = gz
	}

	body, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("read body error: %w", err)
	}

	return body, nil
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
	Aid      int64      `json:"aid"`
	Bvid     string     `json:"bvid"`
	Title    string     `json:"title"`
	Pic      string     `json:"pic"`
	Desc     string     `json:"desc"`
	Duration int        `json:"duration"`
	Tid      int        `json:"tid"`
	Tname    string     `json:"tname"`
	Owner    VideoOwner `json:"owner"`
	Stat     VideoStat  `json:"stat"`
	AddAt    int64      `json:"add_at"`
	Pubdate  int64      `json:"pubdate"`
}

// WatchLaterData is the raw payload returned by /x/v2/history/toview.
type WatchLaterData struct {
	Count int              `json:"count"`
	List  []WatchLaterItem `json:"list"`
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

// DeleteBiliHistory 删除B站观看历史记录
// bvids: 要删除的视频 BV 号列表
func (c *Client) DeleteBiliHistory(bvids []string) error {
	if c.BiliJct == "" {
		return fmt.Errorf("bili_jct (csrf) is required to delete history")
	}
	if len(bvids) == 0 {
		return fmt.Errorf("bvids is empty")
	}

	form := url.Values{}
	form.Set("bvid", strings.Join(bvids, ","))
	form.Set("csrf", c.BiliJct)

	body, err := c.PostForm(HistoryDelURL, form)
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

type UserCardInfo struct {
	Mid       string `json:"mid"`
	Name      string `json:"name"`
	Face      string `json:"face"`
	Sign      string `json:"sign"`
	Level     int    `json:"level"`
	Fans      int    `json:"fans"`
	Attention int    `json:"attention"`
	Archive   int    `json:"archive"`
}

type UserCardResponse struct {
	Card UserCardInfo `json:"card"`
}

func (c *Client) GetUserCard(mid string) (*UserCardInfo, error) {
	params := map[string]string{
		"mid": mid,
	}

	body, err := c.GetWithDm(UserCardURL, params)
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

	var result UserCardResponse
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("unmarshal data error: %w", err)
	}

	return &result.Card, nil
}

func (c *Client) GetDynamicList(hostMid string, offset string, ps int) (*DynamicSpaceResponse, error) {
	// 先确保获取 wbi keys
	if c.ImgKey == "" || c.SubKey == "" {
		c.FetchWbiKeys()
	}

	params := map[string]string{
		"host_mid": hostMid,
		"ps":       fmt.Sprintf("%d", ps),
		"features": "itemOpusStyle,listOnlyfans,opusBigCover,onlyfansVote,decorationCard,onlyfansAssetsV2,forwardListHidden,ugcDelete,commentsNewVersion",
		"platform": "web",
	}
	if offset != "" {
		params["offset"] = offset
	}

	// 如果有 wbi keys，添加签名
	if c.ImgKey != "" && c.SubKey != "" {
		w_rid := wbiSign(params, c.ImgKey, c.SubKey)
		params["w_rid"] = w_rid
	}

	body, err := c.GetWithDm(DynamicSpaceURL, params)
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
	Type       int    `json:"type"`
	Link       string `json:"link"`
}

type FavFolderListData struct {
	Count int             `json:"count"`
	List  []FavFolderInfo `json:"list"`
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

func (c *Client) GetCollectedFavoriteFolders(upMid string, pn, ps int) (*FavFolderListData, error) {
	params := map[string]string{
		"up_mid":      upMid,
		"pn":          fmt.Sprintf("%d", pn),
		"ps":          fmt.Sprintf("%d", ps),
		"platform":    "web",
		"web_location": "0.0",
	}

	// 确保获取 wbi keys
	if c.ImgKey == "" || c.SubKey == "" {
		_ = c.FetchWbiKeys()
	}

	// 添加 wbi 签名
	if c.ImgKey != "" && c.SubKey != "" {
		w_rid := wbiSign(params, c.ImgKey, c.SubKey)
		params["w_rid"] = w_rid
	}

	body, err := c.Get(FavoriteCollectedListURL, params)
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
	ID       int64    `json:"id"`
	Type     int      `json:"type"`
	Title    string   `json:"title"`
	Cover    string   `json:"cover"`
	Intro    string   `json:"intro"`
	Page     int      `json:"page"`
	Duration int      `json:"duration"`
	Upper    FavUpper `json:"upper"`
	Ctime    int64    `json:"ctime"`
	Pubtime  int64    `json:"pubtime"`
	FavTime  int64    `json:"fav_time"`
	Attr     int      `json:"attr"`
	UGC      *FavUGC  `json:"ugc,omitempty"`
	Stat     *FavStat `json:"stat,omitempty"`
	Cid      int64    `json:"cid"`
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
	Info  *FavFolderInfo    `json:"info"`
	Media []FavResourceItem `json:"medias"`
	Page  FavResourcePage   `json:"page"`
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

type SeasonMediaItem struct {
	ID       int64      `json:"id"`
	Title    string     `json:"title"`
	Cover    string     `json:"cover"`
	Duration int        `json:"duration"`
	Pubtime  int64      `json:"pubtime"`
	Bvid     string     `json:"bvid"`
	Upper    FavUpper   `json:"upper"`
	CntInfo  *FavStat   `json:"cnt_info"`
}

type SeasonData struct {
	Info  *FavFolderInfo    `json:"info"`
	Media []SeasonMediaItem `json:"medias"`
	Page  FavResourcePage   `json:"page"`
}

func (c *Client) GetSeasonContents(seasonID int64, pn, ps int) (*SeasonData, error) {
	params := map[string]string{
		"season_id":  fmt.Sprintf("%d", seasonID),
		"pn":         fmt.Sprintf("%d", pn),
		"ps":         fmt.Sprintf("%d", ps),
		"web_location": "0.0",
	}
	body, err := c.Get(FavoriteSeasonListURL, params)
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
	var data SeasonData
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
	Aid      int64      `json:"aid"`
	Title    string     `json:"title"`
	Pic      string     `json:"pic"`
	Desc     string     `json:"desc"`
	Duration int        `json:"duration"`
	Tid      int        `json:"tid"`
	Tname    string     `json:"tname"`
	Owner    VideoOwner `json:"owner"`
	Stat     VideoStat  `json:"stat"`
	Pubdate  int64      `json:"pubdate"`
	Bvid     string     `json:"bvid"`
}

type LikedVideoData struct {
	List []LikedVideoItem `json:"list"`
}

func (c *Client) GetLikedVideos(vmid int64) (*LikedVideoData, error) {
	params := map[string]string{
		"vmid": fmt.Sprintf("%d", vmid),
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

