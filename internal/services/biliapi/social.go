package biliapi

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// Comment represents a video comment.
type Comment struct {
	Rpid       int64  `json:"rpid"`
	Content    string `json:"content"`
	Member     CommentMember `json:"member"`
	Like       int    `json:"like"`
	ReplyCount int    `json:"rcount"`
	Ctime      int64  `json:"ctime"`
}

type CommentMember struct {
	Mid  int64  `json:"mid"`
	Name string `json:"name"`
	Face string `json:"face"`
}

// FetchComments fetches comments for a video.
func (c *Client) FetchComments(bvid string, pn, ps int) ([]Comment, int, error) {
	if pn <= 0 {
		pn = 1
	}
	if ps <= 0 {
		ps = 20
	}

	url := fmt.Sprintf("https://api.bilibili.com/x/v2/reply?type=1&oid=%s&pn=%d&ps=%d", bvid, pn, ps)
	data, err := c.doRequest(url)
	if err != nil {
		return nil, 0, err
	}

	code, _ := data["code"].(float64)
	if code != 0 {
		msg, _ := data["message"].(string)
		return nil, 0, fmt.Errorf("API error: %s", msg)
	}

	respData, _ := data["data"].(map[string]interface{})
	if respData == nil {
		return nil, 0, nil
	}

	// Parse page info for total
	pageInfo, _ := respData["page"].(map[string]interface{})
	total := 0
	if count, ok := pageInfo["count"].(float64); ok {
		total = int(count)
	}

	// Parse replies
	repliesJSON, _ := json.Marshal(respData["replies"])
	var comments []Comment
	if repliesJSON != nil {
		json.Unmarshal(repliesJSON, &comments)
	}

	return comments, total, nil
}

// FavoriteFolder represents a Bilibili favorite folder.
type FavoriteFolder struct {
	ID         int64  `json:"id"`
	FID        int64  `json:"fid"`
	MID        int64  `json:"mid"`
	Title      string `json:"title"`
	Cover      string `json:"cover"`
	Attr       int    `json:"attr"`
	Intro      string `json:"intro"`
	CTime      int64  `json:"ctime"`
	MTime      int64  `json:"mtime"`
	State      int    `json:"state"`
	MediaCount int    `json:"media_count"`
	FavState   int    `json:"fav_state"`
	LikeState  int    `json:"like_state"`
	Upper      struct {
		MID  int64  `json:"mid"`
		Name string `json:"name"`
		Face string `json:"face"`
	} `json:"upper"`
}

// FavoriteFolderListData is the response data for favorite folder list.
type FavoriteFolderListData struct {
	List []FavoriteFolder `json:"list"`
}

// FetchCreatedFavorites fetches all favorite folders created by a user.
func (c *Client) FetchCreatedFavorites(upMid int64) (*FavoriteFolderListData, error) {
	url := "https://api.bilibili.com/x/v3/fav/folder/created/list-all"
	if upMid > 0 {
		url += fmt.Sprintf("?up_mid=%d&type=0", upMid)
	}
	data, err := c.doRequest(url)
	if err != nil {
		return nil, err
	}

	code, _ := data["code"].(float64)
	if code != 0 {
		msg, _ := data["message"].(string)
		return nil, fmt.Errorf("API error: %s", msg)
	}

	respData, _ := data["data"].(map[string]interface{})
	if respData == nil {
		return &FavoriteFolderListData{}, nil
	}

	listJSON, _ := json.Marshal(respData["list"])
	var folders []FavoriteFolder
	if listJSON != nil {
		json.Unmarshal(listJSON, &folders)
	}

	return &FavoriteFolderListData{List: folders}, nil
}

// FetchCollectedFavorites fetches favorite folders collected by a user.
func (c *Client) FetchCollectedFavorites(upMid int64, pn, ps int) (*FavoriteFolderListData, error) {
	url := fmt.Sprintf("https://api.bilibili.com/x/v3/fav/folder/collected/list?up_mid=%d&pn=%d&ps=%d", upMid, pn, ps)
	data, err := c.doRequest(url)
	if err != nil {
		return nil, err
	}

	code, _ := data["code"].(float64)
	if code != 0 {
		msg, _ := data["message"].(string)
		return nil, fmt.Errorf("API error: %s", msg)
	}

	respData, _ := data["data"].(map[string]interface{})
	if respData == nil {
		return &FavoriteFolderListData{}, nil
	}

	listJSON, _ := json.Marshal(respData["list"])
	var folders []FavoriteFolder
	if listJSON != nil {
		json.Unmarshal(listJSON, &folders)
	}

	return &FavoriteFolderListData{List: folders}, nil
}

// FavoriteContent represents a single item in a favorite folder.
type FavoriteContent struct {
	ID        int64  `json:"id"`
	Type      int    `json:"type"`
	Title     string `json:"title"`
	Cover     string `json:"cover"`
	BVID      string `json:"bvid"`
	Intro     string `json:"intro"`
	Page      int    `json:"page"`
	Duration  int    `json:"duration"`
	UpperMid  int64  `json:"upper_mid"`
	Attr      int    `json:"attr"`
	CTime     int64  `json:"ctime"`
	PubTime   int64  `json:"pubtime"`
	FavTime   int64  `json:"fav_time"`
	Play      int    `json:"play"`
	Danmaku   int    `json:"danmaku"`
	Reply     int    `json:"reply"`
	Upper     struct {
		MID  int64  `json:"mid"`
		Name string `json:"name"`
		Face string `json:"face"`
	} `json:"upper"`
	CntInfo struct {
		Collect int `json:"collect"`
		Play    int `json:"play"`
		Danmaku int `json:"danmaku"`
		Reply   int `json:"reply"`
	} `json:"cnt_info"`
}

// FavoriteContentListData is the response data for favorite folder content.
type FavoriteContentListData struct {
	Info   FavoriteFolder    `json:"info"`
	Medias []FavoriteContent `json:"medias"`
	HasMore bool             `json:"has_more"`
}

// FetchFavoriteContent fetches content from a favorite folder.
func (c *Client) FetchFavoriteContent(mediaID int64, pn, ps int, keyword, order string) (*FavoriteContentListData, error) {
	if pn <= 0 {
		pn = 1
	}
	if ps <= 0 {
		ps = 20
	}
	if order == "" {
		order = "mtime"
	}

	url := fmt.Sprintf("https://api.bilibili.com/x/v3/fav/resource/list?media_id=%d&pn=%d&ps=%d&order=%s&platform=web", mediaID, pn, ps, order)
	if keyword != "" {
		url += "&keyword=" + keyword
	}

	data, err := c.doRequest(url)
	if err != nil {
		return nil, err
	}

	code, _ := data["code"].(float64)
	if code != 0 {
		msg, _ := data["message"].(string)
		return nil, fmt.Errorf("API error: %s", msg)
	}

	respData, _ := data["data"].(map[string]interface{})
	if respData == nil {
		return &FavoriteContentListData{}, nil
	}

	result := &FavoriteContentListData{}

	if hasMore, ok := respData["has_more"].(bool); ok {
		result.HasMore = hasMore
	}

	if infoJSON, ok := respData["info"]; ok {
		infoBytes, _ := json.Marshal(infoJSON)
		json.Unmarshal(infoBytes, &result.Info)
	}

	if mediasJSON, ok := respData["medias"]; ok {
		mediasBytes, _ := json.Marshal(mediasJSON)
		json.Unmarshal(mediasBytes, &result.Medias)
	}

	return result, nil
}

// CurrentUserInfo holds current logged-in user info.
type CurrentUserInfo struct {
	UID    int64  `json:"mid"`
	Uname  string `json:"uname"`
	IsLogin bool  `json:"isLogin"`
	Face   string `json:"face"`
}

// FetchCurrentUser fetches the current logged-in user info.
func (c *Client) FetchCurrentUser() (*CurrentUserInfo, error) {
	url := "https://api.bilibili.com/x/web-interface/nav"
	data, err := c.doRequest(url)
	if err != nil {
		return nil, err
	}

	code, _ := data["code"].(float64)
	if code != 0 {
		msg, _ := data["message"].(string)
		return nil, fmt.Errorf("API error: %s", msg)
	}

	respData, _ := data["data"].(map[string]interface{})
	if respData == nil {
		return nil, fmt.Errorf("empty data")
	}

	user := &CurrentUserInfo{}
	if uid, ok := respData["mid"].(float64); ok {
		user.UID = int64(uid)
	}
	if uname, ok := respData["uname"].(string); ok {
		user.Uname = uname
	}
	if isLogin, ok := respData["isLogin"].(bool); ok {
		user.IsLogin = isLogin
	}
	if face, ok := respData["face"].(string); ok {
		user.Face = face
	}

	return user, nil
}

// WatchLater represents a watch later item.
type WatchLater struct {
	AID       int64  `json:"aid"`
	BVID      string `json:"bvid"`
	Title     string `json:"title"`
	Pic       string `json:"pic"`
	Desc      string `json:"desc"`
	Duration  int    `json:"duration"`
	TID       int    `json:"tid"`
	TName     string `json:"tname"`
	AddAt     int64  `json:"add_at"`
	PubDate   int64  `json:"pubdate"`
	OwnerName string `json:"owner_name"`
	OwnerMid  int64  `json:"owner_mid"`
	OwnerFace string `json:"owner_face"`
	View      int    `json:"view"`
	Danmaku   int    `json:"danmaku"`
	Link      string `json:"link"`
	Owner     struct {
		MID  int64  `json:"mid"`
		Name string `json:"name"`
		Face string `json:"face"`
	} `json:"owner"`
	Stat struct {
		View    int `json:"view"`
		Danmaku int `json:"danmaku"`
	} `json:"stat"`
}

// WatchLaterListData is the response data for watch later list.
type WatchLaterListData struct {
	List []WatchLater `json:"list"`
}

// FetchWatchLater fetches the watch later list.
func (c *Client) FetchWatchLater() (*WatchLaterListData, error) {
	url := "https://api.bilibili.com/x/v2/history/toview"
	data, err := c.doRequest(url)
	if err != nil {
		return nil, err
	}

	code, _ := data["code"].(float64)
	if code != 0 {
		msg, _ := data["message"].(string)
		return nil, fmt.Errorf("API error: %s", msg)
	}

	respData, _ := data["data"].(map[string]interface{})
	if respData == nil {
		return &WatchLaterListData{}, nil
	}

	listJSON, _ := json.Marshal(respData["list"])
	var items []WatchLater
	if listJSON != nil {
		json.Unmarshal(listJSON, &items)
	}

	for i := range items {
		items[i].OwnerName = items[i].Owner.Name
		items[i].OwnerMid = items[i].Owner.MID
		items[i].OwnerFace = items[i].Owner.Face
		items[i].View = items[i].Stat.View
		items[i].Danmaku = items[i].Stat.Danmaku
		if items[i].BVID != "" {
			items[i].Link = "https://www.bilibili.com/video/" + items[i].BVID
		}
	}

	return &WatchLaterListData{List: items}, nil
}

// LikeVideo represents a liked video.
type LikeVideo struct {
	AID       int64  `json:"aid"`
	BVID      string `json:"bvid"`
	Title     string `json:"title"`
	Pic       string `json:"pic"`
	Desc      string `json:"desc"`
	Duration  int    `json:"duration"`
	TID       int    `json:"tid"`
	TName     string `json:"tname"`
	PubDate   int64  `json:"pubdate"`
	OwnerName string `json:"owner_name"`
	OwnerMid  int64  `json:"owner_mid"`
	OwnerFace string `json:"owner_face"`
	View      int    `json:"view"`
	Danmaku   int    `json:"danmaku"`
	LikeCount int    `json:"like_count"`
	Link      string `json:"link"`
	Owner     struct {
		MID  int64  `json:"mid"`
		Name string `json:"name"`
		Face string `json:"face"`
	} `json:"owner"`
	Stat struct {
		View    int `json:"view"`
		Danmaku int `json:"danmaku"`
		Like    int `json:"like"`
	} `json:"stat"`
}

// LikeListData is the response data for like list.
type LikeListData struct {
	List []LikeVideo `json:"list"`
}

// FetchLikes fetches the user's liked videos.
func (c *Client) FetchLikes() (*LikeListData, error) {
	if c.dedeUserID == "" {
		return nil, fmt.Errorf("user not logged in")
	}

	url := fmt.Sprintf("https://api.bilibili.com/x/space/like/video?vmid=%s&pn=1&ps=50", c.dedeUserID)
	data, err := c.doRequest(url)
	if err != nil {
		return nil, err
	}

	code, _ := data["code"].(float64)
	if code != 0 {
		msg, _ := data["message"].(string)
		return nil, fmt.Errorf("API error: %s", msg)
	}

	respData, _ := data["data"].(map[string]interface{})
	if respData == nil {
		return &LikeListData{}, nil
	}

	listJSON, _ := json.Marshal(respData["list"])
	var items []LikeVideo
	if listJSON != nil {
		json.Unmarshal(listJSON, &items)
	}

	for i := range items {
		items[i].OwnerName = items[i].Owner.Name
		items[i].OwnerMid = items[i].Owner.MID
		items[i].OwnerFace = items[i].Owner.Face
		items[i].View = items[i].Stat.View
		items[i].Danmaku = items[i].Stat.Danmaku
		items[i].LikeCount = items[i].Stat.Like
		if items[i].BVID != "" {
			items[i].Link = "https://www.bilibili.com/video/" + items[i].BVID
		}
	}

	return &LikeListData{List: items}, nil
}

// Dynamic represents a user dynamic entry.
type Dynamic struct {
	DynID  int64  `json:"dyn_id"`
	Type   int    `json:"type"`
	BVID   string `json:"bvid"`
	Title  string `json:"title"`
	Desc   string `json:"desc"`
}

// FetchDynamics fetches user dynamics.
func (c *Client) FetchDynamics() ([]Dynamic, error) {
	url := "https://api.bilibili.com/x/polymer/web-dynamic/v1/feed/all"
	data, err := c.doRequest(url)
	if err != nil {
		return nil, err
	}

	code, _ := data["code"].(float64)
	if code != 0 {
		msg, _ := data["message"].(string)
		return nil, fmt.Errorf("API error: %s", msg)
	}

	respData, _ := data["data"].(map[string]interface{})
	if respData == nil {
		return nil, nil
	}

	itemsJSON, _ := json.Marshal(respData["items"])
	var dynamics []Dynamic
	if itemsJSON != nil {
		json.Unmarshal(itemsJSON, &dynamics)
	}

	return dynamics, nil
}

// PopularVideo represents a popular video.
type PopularVideo struct {
	BVID   string `json:"bvid"`
	Title  string `json:"title"`
	Owner  struct {
		Name string `json:"name"`
	} `json:"owner"`
	Stat struct {
		View int `json:"view"`
	} `json:"stat"`
}

// FetchPopular fetches popular videos.
func (c *Client) FetchPopular() ([]PopularVideo, error) {
	url := "https://api.bilibili.com/x/web-interface/popular?ps=20&pn=1"
	data, err := c.doRequest(url)
	if err != nil {
		return nil, err
	}

	code, _ := data["code"].(float64)
	if code != 0 {
		msg, _ := data["message"].(string)
		return nil, fmt.Errorf("API error: %s", msg)
	}

	respData, _ := data["data"].(map[string]interface{})
	if respData == nil {
		return nil, nil
	}

	listJSON, _ := json.Marshal(respData["list"])
	var videos []PopularVideo
	if listJSON != nil {
		json.Unmarshal(listJSON, &videos)
	}

	return videos, nil
}

// DeleteHistory deletes a history record from Bilibili.
func (c *Client) DeleteHistory(bvid string) error {
	req, err := http.NewRequest("POST", "https://api.bilibili.com/x/web-interface/history/delete", nil)
	if err != nil {
		return err
	}

	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Referer", referer)
	if c.sessdata != "" {
		req.Header.Set("Cookie", "SESSDATA="+c.sessdata)
	}

	// Add form data
	q := req.URL.Query()
	q.Set("bvid", bvid)
	req.URL.RawQuery = q.Encode()

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	json.Unmarshal(body, &result)

	code, _ := result["code"].(float64)
	if code != 0 {
		msg, _ := result["message"].(string)
		return fmt.Errorf("API error: %s", msg)
	}

	return nil
}
