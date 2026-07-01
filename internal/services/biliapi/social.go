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
	ID    int64  `json:"id"`
	Title string `json:"title"`
	Count int    `json:"media_count"`
}

// FetchFavorites fetches the user's favorite folders.
func (c *Client) FetchFavorites() ([]FavoriteFolder, error) {
	url := "https://api.bilibili.com/x/v3/fav/folder/created/list-all"
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
	var folders []FavoriteFolder
	if listJSON != nil {
		json.Unmarshal(listJSON, &folders)
	}

	return folders, nil
}

// WatchLater represents a watch later item.
type WatchLater struct {
	BVID    string `json:"bvid"`
	Title   string `json:"title"`
	Owner   struct {
		Name string `json:"name"`
	} `json:"owner"`
}

// FetchWatchLater fetches the watch later list.
func (c *Client) FetchWatchLater() ([]WatchLater, error) {
	url := "https://api.bilibili.com/x/v2/history/toview/web"
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
	var items []WatchLater
	if listJSON != nil {
		json.Unmarshal(listJSON, &items)
	}

	return items, nil
}

// LikeVideo represents a liked video.
type LikeVideo struct {
	BVID  string `json:"bvid"`
	Title string `json:"title"`
}

// FetchLikes fetches the user's liked videos.
func (c *Client) FetchLikes() ([]LikeVideo, error) {
	url := "https://api.bilibili.com/x/v2/like/web/list?like_av=1"
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

	listJSON, _ := json.Marshal(respData["item"])
	var items []LikeVideo
	if listJSON != nil {
		json.Unmarshal(listJSON, &items)
	}

	return items, nil
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
