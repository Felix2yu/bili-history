package services

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"bilibili-history-go/biliapi"
	"bilibili-history-go/config"
	"bilibili-history-go/database"
	"bilibili-history-go/utils"
)

type DynamicFetchStatus struct {
	IsRunning    bool   `json:"is_running"`
	HostMid      string `json:"host_mid"`
	TotalFetched int    `json:"total_fetched"`
	TotalPages   int    `json:"total_pages"`
	Message      string `json:"message,omitempty"`
}

var (
	dynamicFetchStatus DynamicFetchStatus
	dynamicFetchMutex  sync.Mutex
	dynamicStopFlags   sync.Map
	dynamicProgressCh  = make(chan string, 100)
)

func GetDynamicFetchStatus(hostMid string) DynamicFetchStatus {
	dynamicFetchMutex.Lock()
	defer dynamicFetchMutex.Unlock()
	return dynamicFetchStatus
}

func setDynamicFetchStatus(status DynamicFetchStatus) {
	dynamicFetchMutex.Lock()
	defer dynamicFetchMutex.Unlock()
	dynamicFetchStatus = status
}

func StopDynamicFetch(hostMid string) {
	dynamicStopFlags.Store(hostMid, true)
}

func GetDynamicProgressChan() chan string {
	return dynamicProgressCh
}

// DynamicType 表示动态类型常量
const (
	DynamicTypeVideo   = "DYNAMIC_TYPE_AV"      // 视频动态
	DynamicTypeDraw    = "DYNAMIC_TYPE_DRAW"    // 图文动态
	DynamicTypeOpus    = "DYNAMIC_TYPE_OPUS"    // 文章动态
	DynamicTypeForward = "DYNAMIC_TYPE_FORWARD" // 转发动态
	DynamicTypeNone    = "DYNAMIC_TYPE_NONE"    // 纯文本动态
)

func FetchDynamicSpace(hostMid string, needTop, saveToDB, saveMedia bool, dynamicTypes []string) {
	cfg := config.GetConfig()
	if cfg == nil || cfg.SESSDATA == "" {
		setDynamicFetchStatus(DynamicFetchStatus{Message: "SESSDATA 未配置"})
		return
	}

	// 构建类型过滤映射
	typeFilter := make(map[string]bool)
	if len(dynamicTypes) > 0 {
		for _, t := range dynamicTypes {
			typeFilter[t] = true
		}
	}

	setDynamicFetchStatus(DynamicFetchStatus{
		IsRunning: true,
		HostMid:   hostMid,
		Message:   "正在抓取...",
	})

	dynamicStopFlags.Store(hostMid, false)

	go func() {
		defer func() {
			status := GetDynamicFetchStatus(hostMid)
			status.IsRunning = false
			if status.Message == "正在抓取..." {
				status.Message = "抓取完成"
			}
			setDynamicFetchStatus(status)
			dynamicStopFlags.Delete(hostMid)
		}()

		client := biliapi.NewClientWithConfig(cfg.SESSDATA, cfg.BiliJct, cfg.DedeUserID)
		offset := ""
		totalFetched := 0
		totalSkipped := 0
		totalPages := 0
		maxPages := 100

		for {
			if shouldStop, _ := dynamicStopFlags.Load(hostMid); shouldStop == true {
				sendDynamicProgress(fmt.Sprintf("[停止] 用户请求停止，已获取 %d 条动态", totalFetched))
				setDynamicFetchStatus(DynamicFetchStatus{
					HostMid:      hostMid,
					TotalFetched: totalFetched,
					TotalPages:   totalPages,
					Message:      fmt.Sprintf("已停止，共获取 %d 条动态", totalFetched),
				})
				return
			}

			totalPages++
			sendDynamicProgress(fmt.Sprintf("[第 %d 页] 正在获取...", totalPages))

			result, err := client.GetDynamicList(hostMid, offset, 30)
			if err != nil {
				sendDynamicProgress(fmt.Sprintf("[错误] %v", err))
				setDynamicFetchStatus(DynamicFetchStatus{
					HostMid:      hostMid,
					TotalFetched: totalFetched,
					TotalPages:   totalPages,
					Message:      fmt.Sprintf("抓取出错: %v", err),
				})
				return
			}

			if len(result.Items) == 0 {
				sendDynamicProgress("[完成] 没有更多动态")
				break
			}

			// Parse and optionally download media
			var dbItems []database.DynamicItem
			for _, rawItem := range result.Items {
				item := parseDynamicItem(rawItem, hostMid)

				// 应用类型过滤
				if len(typeFilter) > 0 && !typeFilter[item.Type] {
					totalSkipped++
					continue
				}

				if saveMedia {
					downloadDynamicMedia(&item, hostMid)
				}

				dbItems = append(dbItems, item)
			}

			if saveToDB && len(dbItems) > 0 {
				inserted, err := database.SaveDynamics(hostMid, dbItems)
				if err != nil {
					sendDynamicProgress(fmt.Sprintf("[错误] 保存失败: %v", err))
				} else {
					sendDynamicProgress(fmt.Sprintf("[第 %d 页] 获取 %d 条，新增 %d 条", totalPages, len(dbItems), inserted))
				}
			} else if len(dbItems) > 0 {
				sendDynamicProgress(fmt.Sprintf("[第 %d 页] 获取 %d 条", totalPages, len(dbItems)))
			}

			totalFetched += len(dbItems)

			setDynamicFetchStatus(DynamicFetchStatus{
				IsRunning:    true,
				HostMid:      hostMid,
				TotalFetched: totalFetched,
				TotalPages:   totalPages,
			})

			if !result.HasMore || totalPages >= maxPages {
				break
			}

			offset = result.Offset
			time.Sleep(500 * time.Millisecond)
		}

		skipMsg := ""
		if totalSkipped > 0 {
			skipMsg = fmt.Sprintf("，跳过 %d 条", totalSkipped)
		}
		sendDynamicProgress(fmt.Sprintf("[全部抓取完毕] 抓取完成！共获取 %d 条动态%s，总计 %d 页", totalFetched, skipMsg, totalPages))
		setDynamicFetchStatus(DynamicFetchStatus{
			HostMid:      hostMid,
			TotalFetched: totalFetched,
			TotalPages:   totalPages,
			Message:      fmt.Sprintf("抓取完成，共 %d 条动态%s", totalFetched, skipMsg),
		})
	}()
}

func sendDynamicProgress(msg string) {
	select {
	case dynamicProgressCh <- msg:
	default:
	}
}

func parseDynamicItem(raw biliapi.DynamicRawItem, hostMid string) database.DynamicItem {
	item := database.DynamicItem{
		ID:      raw.IDStr,
		Type:    raw.Type,
		HostMid: hostMid,
	}

	var modules map[string]json.RawMessage
	json.Unmarshal(raw.Modules, &modules)

	// Parse module_author for author info
	if authorRaw, ok := modules["module_author"]; ok {
		var author struct {
			Name string `json:"name"`
		}
		json.Unmarshal(authorRaw, &author)
		item.AuthorName = author.Name
	}

	// Parse module_dynamic for content
	if dynRaw, ok := modules["module_dynamic"]; ok {
		var dyn struct {
			Major struct {
				Type string `json:"type"`
				Draw struct {
					Items []struct {
						Src string `json:"src"`
					} `json:"items"`
				} `json:"draw"`
				Archive struct {
					Title string `json:"title"`
					Desc  string `json:"desc"`
					Cover struct {
						Src string `json:"src"`
					} `json:"cover"`
					Bvid string `json:"bvid"`
				} `json:"archive"`
				Opus struct {
					Title   string `json:"title"`
					Summary struct {
						Text string `json:"text"`
					} `json:"summary"`
				} `json:"opus"`
			} `json:"major"`
			Desc struct {
				Text string `json:"text"`
			} `json:"desc"`
		}
		json.Unmarshal(dynRaw, &dyn)

		item.Txt = dyn.Desc.Text

		switch dyn.Major.Type {
		case "MAJOR_TYPE_ARCHIVE":
			item.Bvid = dyn.Major.Archive.Bvid
			item.Title = dyn.Major.Archive.Title
			item.Desc = dyn.Major.Archive.Desc
			item.Cover = dyn.Major.Archive.Cover.Src
		case "MAJOR_TYPE_DRAW":
			item.OpusTitle = dyn.Major.Opus.Title
			item.OpusSummaryText = dyn.Major.Opus.Summary.Text
			if item.OpusSummaryText == "" {
				item.OpusSummaryText = dyn.Desc.Text
			}
		case "MAJOR_TYPE_OPUS":
			item.OpusTitle = dyn.Major.Opus.Title
			item.OpusSummaryText = dyn.Major.Opus.Summary.Text
		}
	}

	// Parse module_dynamic.topic for additional text
	if topicRaw, ok := modules["module_dynamic"]; ok {
		var topic struct {
			Desc struct {
				Text string `json:"text"`
			} `json:"desc"`
		}
		json.Unmarshal(topicRaw, &topic)
		if topic.Desc.Text != "" && item.Txt == "" {
			item.Txt = topic.Desc.Text
		}
	}

	// Parse publish time from module_author
	if authorRaw, ok := modules["module_author"]; ok {
		var author struct {
			PubTs int64 `json:"pub_ts"`
		}
		json.Unmarshal(authorRaw, &author)
		item.PublishTS = author.PubTs
	}

	item.RawJSON = string(raw.Modules)
	return item
}

func downloadDynamicMedia(item *database.DynamicItem, hostMid string) {
	outputDir := utils.GetOutputPath("dynamic", hostMid, "images")

	cfg := config.GetConfig()
	sessdata := ""
	if cfg != nil {
		sessdata = cfg.SESSDATA
	}

	// Download cover for video dynamics
	if item.Cover != "" {
		ext := extractExt(item.Cover)
		filename := fmt.Sprintf("%s_cover%s", sanitizeID(item.ID), ext)
		localPath := filepath.Join(outputDir, filename)
		if err := downloadFile(item.Cover, localPath, sessdata); err == nil {
			relPath := fmt.Sprintf("dynamic/%s/images/%s", hostMid, filename)
			item.Cover = relPath
			item.MediaLocals = append(item.MediaLocals, relPath)
		}
	}

	// Download images from draw/opus dynamics
	// The images are not in the parsed item yet, we need to extract them from raw JSON
	var modules map[string]json.RawMessage
	json.Unmarshal([]byte(item.RawJSON), &modules)
	if dynRaw, ok := modules["module_dynamic"]; ok {
		var dyn struct {
			Major struct {
				Draw struct {
					Items []struct {
						Src string `json:"src"`
					} `json:"items"`
				} `json:"draw"`
			} `json:"major"`
		}
		json.Unmarshal(dynRaw, &dyn)

		for idx, img := range dyn.Major.Draw.Items {
			if img.Src == "" {
				continue
			}
			ext := extractExt(img.Src)
			filename := fmt.Sprintf("%s_%d%s", sanitizeID(item.ID), idx, ext)
			localPath := filepath.Join(outputDir, filename)
			if err := downloadFile(img.Src, localPath, sessdata); err == nil {
				relPath := fmt.Sprintf("dynamic/%s/images/%s", hostMid, filename)
				item.MediaLocals = append(item.MediaLocals, relPath)
			}
		}
	}
}

func downloadFile(url, path, sessdata string) error {
	// Use the existing image download function
	_, err := DownloadImage(url, "dynamic", filepath.Base(path))
	return err
}

func extractExt(url string) string {
	// Remove query params
	if idx := strings.Index(url, "?"); idx > 0 {
		url = url[:idx]
	}
	ext := filepath.Ext(url)
	if ext == "" {
		return ".jpg"
	}
	return ext
}

func sanitizeID(id string) string {
	return strings.ReplaceAll(id, "/", "_")
}
