package routers

// features_cov_test.go drives every RegisterFavoriteRoutes handler
// (features.go:22) offline. biliapi/client.go publishes its endpoint URLs as
// package-level vars, so each test stands up an httptest server, repoints the
// vars at it and restores them through t.Cleanup - no request can leave the
// process. Seeded rows use the reserved bvid prefix BVFT and numeric ids built
// on 2013 so they never collide with the other fixture namespaces.

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"bilibili-history-go/biliapi"
	"bilibili-history-go/config"
	"bilibili-history-go/database"
	"bilibili-history-go/scheduler"
	"bilibili-history-go/services"

	"github.com/gin-gonic/gin"
)

// ---------------------------------------------------------------------------
// fixture identifiers
// ---------------------------------------------------------------------------

const (
	ftFolderCreated1 int64 = 20131001
	ftFolderCreated2 int64 = 20131002
	ftFolderEmpty    int64 = 20131003
	ftFolderStale    int64 = 20131004
	ftSeasonFolder   int64 = 20131011
	ftCollectedPlain int64 = 20131012
	ftUnmatchedFID   int64 = 20131050
	ftOrphanMediaID  int64 = 20131051

	ftContent1  int64 = 20132001
	ftContent2  int64 = 20132002
	ftContent3  int64 = 20132003
	ftSeasonA   int64 = 20132011
	ftSeasonB   int64 = 20132012
	ftOrphanOid int64 = 20132013

	ftLikeAID int64  = 20133001
	ftLikeA   string = "BVFTLK0000000001"
	ftLikeBID int64  = 20133002
	ftLikeB   string = "BVFTLK0000000002"

	ftWLAID       int64 = 20134001
	ftWLA               = "BVFTWL0000000001"
	ftWLBIT       int64 = 20134002
	ftWLB               = "BVFTWL0000000002"
	ftWLFail      int64 = 20134003
	ftWLFailBvid        = "BVFTWL0000000003"
	ftWLStale           = "BVFTWL0000000009"
	ftWLUntracked int64 = 20134004

	ftFavVideo1 = "BVFTFA0000000001"
	ftFavVideo2 = "BVFTFA0000000002"
	ftSeasonV1  = "BVFTSE0000000001"
	ftSeasonV2  = "BVFTSE0000000002"
	ftDynVideo  = "BVFTDY0000000001"

	ftMidSelf        int64 = 2013900001
	ftMidOther       int64 = 2013900002
	ftUpName               = "FT功能UP"
	ftDynHost              = "2013777001"
	ftDynHostBusy          = "2013777002"
	ftDynHostStopped       = "2013777003"
	ftDynHostDeleted       = "2013777004"
	ftDynID1               = "2013600000000000001"
	ftDynID2               = "2013600000000000002"
	ftDynID3               = "2013600000000000003"
	ftDynIDAbsent          = "2013600000000000999"
)

// ---------------------------------------------------------------------------
// offline biliapi stub
// ---------------------------------------------------------------------------

const (
	ftPathCreatedList   = "/x/v3/fav/folder/created/list-all"
	ftPathCollectedList = "/x/v3/fav/folder/collected/list"
	ftPathResourceList  = "/x/v3/fav/resource/list"
	ftPathSeasonList    = "/x/space/fav/season/list"
	ftPathFavoriteDeal  = "/x/v3/fav/resource/deal"
	ftPathLikedVideos   = "/x/space/like/video"
	ftPathLike          = "/x/web-interface/archive/like"
	ftPathWatchLater    = "/x/v2/history/toview"
	ftPathWatchLaterDel = "/x/v2/history/toview/del"
	ftPathUserCard      = "/x/web-interface/card"
	ftPathDynamicSpace  = "/x/polymer/web-dynamic/v1/feed/space"
	ftPathNav           = "/x/web-interface/nav"
)

// ftStubPayloads are the canned BiliBili responses served by ftBiliStub. The
// fav ids line up with the fixture constants above.
const (
	ftCreatedFoldersJSON = `{"code":0,"message":"0","data":{"count":3,"list":[
		{"id":20131001,"fid":20131001,"mid":2013900001,"title":"FT同步收藏夹","cover":"https://cdn.bilibili.com/ft/cov1.jpg","attr":0,"intro":"同步一号","ctime":1357000000,"mtime":1357000010,"state":0,"media_count":1,"fav_state":1,"like_state":0,"type":0,"upper":{"mid":2013900001,"name":"` + ftUpName + `","face":"https://cdn.bilibili.com/ft/face1.jpg"}},
		{"id":20131002,"fid":20131002,"mid":2013900001,"title":"FT补封面收藏夹","cover":"","attr":0,"intro":"同步二号","ctime":1357000001,"mtime":1357000011,"state":0,"media_count":2,"fav_state":1,"like_state":0,"type":0,"upper":{"mid":2013900001,"name":"` + ftUpName + `","face":""}},
		{"id":20131003,"fid":20131003,"mid":2013900001,"title":"FT空收藏夹","cover":"","attr":0,"intro":"","ctime":1357000002,"mtime":1357000012,"state":0,"media_count":0,"fav_state":1,"like_state":0,"type":0,"upper":{"mid":2013900001,"name":"` + ftUpName + `","face":""}}
	]}}`

	ftCollectedFoldersJSON = `{"code":0,"message":"0","data":{"count":2,"list":[
		{"id":20131011,"fid":20131011,"mid":2013900002,"title":"FT收藏的合集","cover":"","attr":0,"intro":"别人的合集","ctime":1357000020,"mtime":1357000021,"state":0,"media_count":2,"fav_state":0,"like_state":0,"type":21,"link":"https://www.bilibili.com/medialist/play/20131011","upper":{"mid":2013900002,"name":"FT他人UP","face":""}},
		{"id":20131012,"fid":20131012,"mid":2013900002,"title":"FT收藏的空收藏夹","cover":"","attr":0,"intro":"空","ctime":1357000022,"mtime":1357000023,"state":0,"media_count":0,"fav_state":0,"like_state":0,"type":0,"upper":{"mid":2013900002,"name":"FT他人UP","face":""}}
	]}}`

	ftResource1JSON = `{"code":0,"message":"0","data":{"info":{"id":20131001,"fid":20131001,"mid":2013900001,"title":"FT在线内容收藏夹","cover":"https://cdn.bilibili.com/ft/cov1.jpg","media_count":1,"state":0},"page":{"num":1,"size":20,"count":1,"total":1},"medias":[
		{"id":20132001,"type":2,"title":"FT收藏内容一","cover":"https://cdn.bilibili.com/ft/media1.jpg","intro":"内容一","page":1,"duration":120,"upper":{"mid":2013900002,"name":"FT内容UP","face":"https://cdn.bilibili.com/ft/up2.jpg"},"ctime":1357000000,"pubtime":1357000001,"fav_time":1357000020,"attr":0,"cid":43000001,"ugc":{"bvid":"` + ftFavVideo1 + `"},"stat":{"view":700,"danmaku":7,"reply":3,"favorite":9,"coin":1,"share":1,"like":11}}
	]}}`

	ftResource2JSON = `{"code":0,"message":"0","data":{"info":{"id":20131002,"fid":20131002,"mid":2013900001,"title":"FT多内容收藏夹","cover":"","media_count":2,"state":0},"page":{"num":1,"size":20,"count":2,"total":2},"medias":[
		{"id":20132002,"type":2,"title":"FT收藏内容二","cover":"https://cdn.bilibili.com/ft/media2.jpg","intro":"内容二","page":1,"duration":240,"upper":{"mid":2013900002,"name":"FT内容UP","face":""},"ctime":1357000002,"pubtime":1357000003,"fav_time":1357000021,"attr":0,"cid":43000002,"ugc":{"bvid":"` + ftFavVideo2 + `"},"stat":{"view":800,"danmaku":8,"reply":4,"favorite":10,"like":12}},
		{"id":20132003,"type":2,"title":"已失效视频","cover":"","intro":"","page":1,"duration":0,"upper":{"mid":0,"name":"","face":""},"ctime":1357000004,"pubtime":1357000005,"fav_time":1357000022,"attr":9,"cid":0,"ugc":null,"stat":null}
	]}}`

	ftEmptyResourceJSON = `{"code":0,"message":"0","data":{"info":{"id":0,"fid":0,"mid":0,"title":"","cover":"","media_count":0,"state":0},"page":{"num":1,"size":20,"count":0,"total":0},"medias":[]}}`

	ftSeasonJSON = `{"code":0,"message":"0","data":{"info":{"id":20131011,"fid":20131011,"mid":2013900002,"title":"FT收藏的合集","cover":"https://cdn.bilibili.com/ft/season.jpg","media_count":2,"state":0},"page":{"num":1,"size":100,"count":2,"total":2},"medias":[
		{"id":20132011,"title":"FT合集内容一","cover":"https://cdn.bilibili.com/ft/season1.jpg","duration":210,"pubtime":1357000006,"bvid":"` + ftSeasonV1 + `","upper":{"mid":2013900002,"name":"FT内容UP","face":"https://cdn.bilibili.com/ft/up2.jpg"},"cnt_info":{"view":900,"danmaku":9,"reply":5,"favorite":11,"like":13}},
		{"id":20132012,"title":"FT合集内容二","cover":"","duration":60,"pubtime":1357000007,"bvid":"` + ftSeasonV2 + `","upper":{"mid":2013900002,"name":"FT内容UP","face":""},"cnt_info":null}
	]}}`

	ftLikedJSON = `{"code":0,"message":"0","data":{"list":[
		{"aid":20133001,"bvid":"` + ftLikeA + `","title":"FT点赞一","pic":"https://cdn.bilibili.com/ft/like1.jpg","desc":"赞一","duration":50,"tid":1,"tname":"FT赞分区","owner":{"mid":2013900002,"name":"FT点赞UP","face":"https://cdn.bilibili.com/ft/up2.jpg"},"stat":{"view":100,"danmaku":1,"reply":1,"favorite":2,"coin":1,"share":1,"like":5},"pubdate":1357000000},
		{"aid":20133002,"bvid":"` + ftLikeB + `","title":"FT点赞二","pic":"","desc":"赞二","duration":600,"tid":2,"tname":"FT赞分区二","owner":{"mid":2013900003,"name":"FT点赞UP二","face":""},"stat":{"view":200,"danmaku":2,"reply":2,"favorite":3,"coin":2,"share":2,"like":8},"pubdate":1357000001}
	]}}`

	ftWatchLaterJSON = `{"code":0,"message":"0","data":{"count":2,"list":[
		{"aid":20134001,"bvid":"` + ftWLA + `","title":"FT稍后再看一","pic":"https://cdn.bilibili.com/ft/wl1.jpg","desc":"稍后一","duration":30,"tid":3,"tname":"FT稍后分区","owner":{"mid":2013900002,"name":"FT稍后UP","face":""},"stat":{"view":500,"danmaku":5,"like":50,"favorite":5},"add_at":1357000000,"pubdate":1357000000},
		{"aid":20134002,"bvid":"` + ftWLB + `","title":"FT稍后再看二","pic":"","desc":"稍后二","duration":90,"tid":4,"tname":"FT稍后分区二","owner":{"mid":2013900004,"name":"FT稍后UP二","face":""},"stat":{"view":600,"danmaku":6,"like":60,"favorite":6},"add_at":1357000001,"pubdate":1357000001}
	]}}`

	ftUserCardJSON = `{"code":0,"message":"0","data":{"card":{"mid":"2013900002","name":"` + ftUpName + `","face":"https://cdn.bilibili.com/ft/face1.jpg","sign":"FT签名","level_info":{"current_level":6},"fans":1234,"attention":12}}}`

	ftNavJSON = `{"code":0,"message":"0","data":{"wbi_img":{"img_url":"https://i0.hdslb.com/bfs/wbi/ftimgkeyftimgkeyftimgkeyftimgkey.png","sub_url":"https://i0.hdslb.com/bfs/wbi/ftsubkeyftsubkeyftsubkeyftsubkey.png"}}}`

	ftDynamicOKJSON = `{"code":0,"message":"0","data":{"has_more":false,"offset":"ft-offset-1","items":[
		{"id_str":"` + ftDynID1 + `","type":"DYNAMIC_TYPE_AV","modules":{"module_author":{"name":"` + ftUpName + `","pub_ts":1357000100,"pub_time":"2013-01-01"},"module_dynamic":{"desc":{"text":"FT动态正文"},"major":{"type":"MAJOR_TYPE_ARCHIVE","archive":{"title":"FT动态视频","desc":"FT简介","bvid":"` + ftDynVideo + `","cover":{"src":"https://cdn.bilibili.com/ft/dyn1.jpg"}}}}}}
	]}}`

	ftDynamicDrawJSON = `{"code":0,"message":"0","data":{"has_more":false,"offset":"ft-offset-2","items":[
		{"id_str":"` + ftDynID2 + `","type":"DYNAMIC_TYPE_DRAW","modules":{"module_author":{"name":"` + ftUpName + `","pub_ts":1357000101},"module_dynamic":{"desc":{"text":"FT图文正文"},"major":{"type":"MAJOR_TYPE_DRAW","draw":{"items":[{"src":"https://i0.hdslb.com/ft/remote-1.jpg","width":100,"height":80}]}}}}}
	]}}`

	ftDynamicEmptyJSON = `{"code":0,"message":"0","data":{"has_more":false,"offset":"","items":[]}}`

	ftOKEnvelopeJSON = `{"code":0,"message":"0","data":null}`
)

// ftBiliStub answers the canned payloads above and records what was asked for.
// codes/membership flags are mutable so a single test can flip an endpoint into
// an error state, and block pauses the dynamic feed so the "fetch already
// running" branch becomes deterministic.
type ftBiliStub struct {
	server   *httptest.Server
	mu       sync.Mutex
	codes    map[string]int
	hits     map[string]int
	queries  map[string][]url.Values
	forms    map[string][]url.Values
	members  map[int64]bool
	failAid  map[int64]bool
	block    chan struct{}
	onlySeen []string
}

func ftStartStub(t *testing.T) *ftBiliStub {
	t.Helper()
	stub := &ftBiliStub{
		codes:   make(map[string]int),
		hits:    make(map[string]int),
		queries: make(map[string][]url.Values),
		forms:   make(map[string][]url.Values),
		members: make(map[int64]bool),
		failAid: make(map[int64]bool),
	}
	stub.server = httptest.NewServer(http.HandlerFunc(stub.serve))

	// biliapi keeps its endpoints in package-level vars so tests can repoint
	// them; restore every value once the test is over.
	base := stub.server.URL
	repoint := map[*string]string{
		&biliapi.HistoryURL:               "/x/web-interface/history/cursor",
		&biliapi.HistoryDelURL:            "/x/web-interface/history/del",
		&biliapi.VideoInfoURL:             "/x/web-interface/view",
		&biliapi.WatchLaterURL:            "/x/v2/history/toview",
		&biliapi.WatchLaterDelURL:         "/x/v2/history/toview/del",
		&biliapi.DynamicSpaceURL:          "/x/polymer/web-dynamic/v1/feed/space",
		&biliapi.UserCardURL:              "/x/web-interface/card",
		&biliapi.FavoriteFolderListURL:    "/x/v3/fav/folder/created/list-all",
		&biliapi.FavoriteCollectedListURL: "/x/v3/fav/folder/collected/list",
		&biliapi.FavoriteResourceListURL:  "/x/v3/fav/resource/list",
		&biliapi.FavoriteSeasonListURL:    "/x/space/fav/season/list",
		&biliapi.FavoriteDealURL:          "/x/v3/fav/resource/deal",
		&biliapi.LikedVideoURL:            "/x/space/like/video",
		&biliapi.LikeURL:                  "/x/web-interface/archive/like",
		&biliapi.WbiNavURL:                "/x/web-interface/nav",
	}
	for target, path := range repoint {
		ptr := *target
		previous := ptr
		*target = base + path
		t.Cleanup(func() { *target = previous })
	}
	t.Cleanup(stub.server.Close)
	return stub
}

// ftPrevURLs is captured before repointing so cleanup restores production values.
func (s *ftBiliStub) setCode(path string, code int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.codes[path] = code
}

func (s *ftBiliStub) setMember(mediaID int64, isMember bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.members[mediaID] = isMember
}

func (s *ftBiliStub) setFailAid(aid int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.failAid[aid] = true
}

func (s *ftBiliStub) blockDynamic(ch chan struct{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.block = ch
}

func (s *ftBiliStub) hitsFor(path string) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.hits[path]
}

func (s *ftBiliStub) lastQuery(path string) url.Values {
	s.mu.Lock()
	defer s.mu.Unlock()
	list := s.queries[path]
	if len(list) == 0 {
		return url.Values{}
	}
	return list[len(list)-1]
}

func (s *ftBiliStub) lastForm(path string) url.Values {
	s.mu.Lock()
	defer s.mu.Unlock()
	list := s.forms[path]
	if len(list) == 0 {
		return url.Values{}
	}
	return list[len(list)-1]
}

func (s *ftBiliStub) formCount(path string) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.forms[path])
}

func (s *ftBiliStub) serve(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	s.mu.Lock()
	s.hits[path]++
	code := s.codes[path]
	block := s.block
	var form url.Values
	if r.Method == http.MethodPost {
		if err := r.ParseForm(); err == nil {
			form = r.Form
			s.forms[path] = append(s.forms[path], r.Form)
		}
	}
	s.queries[path] = append(s.queries[path], r.URL.Query())
	members := s.members
	failAids := s.failAid
	s.mu.Unlock()

	if path == ftPathDynamicSpace && block != nil {
		<-block
	}

	w.Header().Set("Content-Type", "application/json")
	if code != 0 {
		fmt.Fprintf(w, `{"code":%d,"message":"ft-stub-error-%d","data":null}`, code, code)
		return
	}

	// The watch later delete API is graded per aid so one batch call can mix a
	// success, a not-member style failure and a missing local row.
	if path == ftPathWatchLaterDel && form.Get("aid") != "" {
		var aid int64
		fmt.Sscanf(form.Get("aid"), "%d", &aid)
		if failAids[aid] {
			io.WriteString(w, `{"code":-404,"message":"ft not found","data":null}`)
			return
		}
	}
	if path == ftPathFavoriteDeal {
		io.WriteString(w, ftOKEnvelopeJSON)
		return
	}
	if path == ftPathLike {
		if form.Get("bvid") == ftLikeB {
			io.WriteString(w, `{"code":-9,"message":"ft 未点赞","data":null}`)
			return
		}
		io.WriteString(w, `{"code":0,"message":"0","data":{"like":true}}`)
		return
	}

	switch path {
	case ftPathCreatedList:
		io.WriteString(w, ftCreatedFoldersJSON)
	case ftPathCollectedList:
		if mid := r.URL.Query().Get("up_mid"); mid != "" && members[0] {
			// membership flag 0 is unused; collected list is static
		}
		io.WriteString(w, ftCollectedFoldersJSON)
	case ftPathResourceList:
		switch r.URL.Query().Get("media_id") {
		case fmt.Sprint(ftFolderCreated1):
			io.WriteString(w, ftResource1JSON)
		case fmt.Sprint(ftFolderCreated2):
			io.WriteString(w, ftResource2JSON)
		default:
			io.WriteString(w, ftEmptyResourceJSON)
		}
	case ftPathSeasonList:
		io.WriteString(w, ftSeasonJSON)
	case ftPathLikedVideos:
		io.WriteString(w, ftLikedJSON)
	case ftPathWatchLater:
		io.WriteString(w, ftWatchLaterJSON)
	case ftPathWatchLaterDel:
		io.WriteString(w, ftOKEnvelopeJSON)
	case ftPathUserCard:
		io.WriteString(w, ftUserCardJSON)
	case ftPathDynamicSpace:
		switch r.URL.Query().Get("host_mid") {
		case ftDynHostStopped:
			io.WriteString(w, ftDynamicDrawJSON)
		case ftDynHostDeleted:
			io.WriteString(w, ftDynamicEmptyJSON)
		default:
			io.WriteString(w, ftDynamicOKJSON)
		}
	case ftPathNav:
		io.WriteString(w, ftNavJSON)
	default:
		t := fmt.Sprintf(`{"code":-412,"message":"ft unexpected path %s","data":null}`, path)
		io.WriteString(w, t)
	}
}

// ftCreds installs full credentials on the config singleton and restores the
// previous values afterwards, the same way setSESSDATA does for SESSDATA alone.
func ftCreds(t *testing.T) {
	t.Helper()
	cfg := loadCfg(t)
	prevSess, prevJct, prevUID := cfg.SESSDATA, cfg.BiliJct, cfg.DedeUserID
	cfg.SESSDATA = "ft-sessdata-2013"
	cfg.BiliJct = "ft-jct-2013"
	cfg.DedeUserID = fmt.Sprint(ftMidSelf)
	t.Cleanup(func() {
		cfg.SESSDATA, cfg.BiliJct, cfg.DedeUserID = prevSess, prevJct, prevUID
	})
}

// ftNoCreds clears every credential the features handlers check for.
func ftNoCreds(t *testing.T) {
	t.Helper()
	cfg := loadCfg(t)
	prevSess, prevJct, prevUID := cfg.SESSDATA, cfg.BiliJct, cfg.DedeUserID
	cfg.SESSDATA, cfg.BiliJct, cfg.DedeUserID = "", "", ""
	t.Cleanup(func() {
		cfg.SESSDATA, cfg.BiliJct, cfg.DedeUserID = prevSess, prevJct, prevUID
	})
}

// ftPlain decodes the {success,message,data} / {status,message} bodies the
// dynamic handlers emit instead of the models.ApiResponse envelope.
func ftPlain(t *testing.T, w *httptest.ResponseRecorder) map[string]interface{} {
	t.Helper()
	var m map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &m); err != nil {
		t.Fatalf("decode plain body %q: %v", w.Body.String(), err)
	}
	return m
}

// ftSuccess drives a dynamic route and requires success:true in the plain body.
func ftSuccess(t *testing.T, e *gin.Engine, method, path, body string) map[string]interface{} {
	t.Helper()
	w := doRaw(t, e, method, path, body)
	if w.Code != http.StatusOK {
		t.Fatalf("%s %s: http %d, want 200 (body %q)", method, path, w.Code, w.Body.String())
	}
	payload := ftPlain(t, w)
	if payload["success"] != true {
		t.Fatalf("%s %s: success = %v, want true (body %q)", method, path, payload["success"], w.Body.String())
	}
	return payload
}

// ftFailure drives a dynamic route and requires success:false plus a message.
func ftFailure(t *testing.T, e *gin.Engine, method, path, body, wantMsg string) map[string]interface{} {
	t.Helper()
	w := doRaw(t, e, method, path, body)
	payload := ftPlain(t, w)
	if payload["success"] != false {
		t.Fatalf("%s %s: success = %v, want false (body %q)", method, path, payload["success"], w.Body.String())
	}
	if msg, _ := payload["message"].(string); !strings.Contains(msg, wantMsg) {
		t.Fatalf("%s %s: message = %q, want substring %q", method, path, msg, wantMsg)
	}
	return payload
}

// ftWaitTask polls an async sync task until it leaves the running state.
func ftWaitTask(t *testing.T, taskID string) map[string]interface{} {
	t.Helper()
	deadline := time.Now().Add(15 * time.Second)
	var status map[string]interface{}
	for time.Now().Before(deadline) {
		status = scheduler.GetAsyncTaskStatus(taskID)
		if status == nil {
			t.Fatalf("task %s vanished from the async registry", taskID)
		}
		switch state, _ := status["status"].(string); state {
		case "completed", "failed":
			return status
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("task %s never finished: %v", taskID, status)
	return nil
}

// ftRunTask completes a doSync* goroutine through the scheduler registry so the
// deferred CompleteAsyncTask branch is exercised as well.
func ftRunTask(t *testing.T, name string, body func(cfg *config.Config, taskID string)) map[string]interface{} {
	t.Helper()
	cfg := loadCfg(t)
	taskID, _ := scheduler.StartAsyncTask(name)
	done := make(chan struct{})
	go func() {
		defer close(done)
		body(cfg, taskID)
	}()
	select {
	case <-done:
	case <-time.After(20 * time.Second):
		t.Fatalf("sync task %s did not return within 20s", name)
	}
	return ftWaitTask(t, taskID)
}

// ---------------------------------------------------------------------------
// side database seeding helpers (favorites / likes / watchlater / dynamic)
// ---------------------------------------------------------------------------

func ftCount(t *testing.T, handle *sql.DB, query string, args ...interface{}) int {
	t.Helper()
	if handle == nil {
		t.Fatal("side database is nil")
	}
	var n int
	if err := handle.QueryRow(query, args...).Scan(&n); err != nil {
		t.Fatalf("count %q: %v", query, err)
	}
	return n
}

func ftExecDB(t *testing.T, handle *sql.DB, query string, args ...interface{}) {
	t.Helper()
	if handle == nil {
		t.Fatal("side database is nil")
	}
	if _, err := handle.Exec(query, args...); err != nil {
		t.Fatalf("exec %q: %v", query, err)
	}
}

// ftSeedFavorites writes the folders and contents the local read endpoints are
// expected to report. Every media_id starts with 2013 so a foreign fixture can
// never be mistaken for one of these rows.
func ftSeedFavorites(t *testing.T) {
	t.Helper()
	fav := database.GetFavoritesDB()
	if fav == nil {
		t.Skip("favorites database unavailable")
	}
	ftExecDB(t, fav, `INSERT OR REPLACE INTO favorites_content
		(id, media_id, content_id, type, title, cover, bvid, duration, upper_mid, fav_time, fetch_time, creator_name)
		VALUES (2013900001, ?, ?, 2, 'FT本地内容一', 'https://cdn.bilibili.com/ft/local1.jpg', ?, 100, ?, ?, ?, ?)`,
		ftFolderCreated1, ftContent1, ftFavVideo1, ftMidOther, 1357000101, 1357000000, ftUpName)
	ftExecDB(t, fav, `INSERT OR REPLACE INTO favorites_content
		(id, media_id, content_id, type, title, cover, bvid, duration, upper_mid, fav_time, fetch_time, creator_name)
		VALUES (2013900002, ?, ?, 2, 'FT本地内容二', 'https://cdn.bilibili.com/ft/local2.jpg', ?, 200, ?, ?, ?, ?)`,
		ftFolderCreated1, ftContent2, ftFavVideo2, ftMidOther, 1357000102, 1357000000, ftUpName)
	// The rows have to go again: they share media_id/content_id with the stub
	// payloads the favourite sync test counts on, so leaving them behind would
	// make that test depend on run order.
	t.Cleanup(func() {
		ftExecDB(t, fav, `DELETE FROM favorites_content WHERE id IN (2013900001, 2013900002)`)
	})
}

// ftFavoriteFolderRow writes one folder; id doubles as the favorites_folder.id
// primary key because batchCheckFavoriteStatus joins on it.
func ftFavoriteFolderRow(t *testing.T, id, mediaID int64, title string, folderType, mediaCount int, mtime int64) {
	t.Helper()
	fav := database.GetFavoritesDB()
	if fav == nil {
		t.Skip("favorites database unavailable")
	}
	ftExecDB(t, fav, `INSERT OR REPLACE INTO favorites_folder
		(id, media_id, fid, mid, title, cover, attr, intro, ctime, mtime, state, media_count, fav_state, like_state, folder_type, fetch_time)
		VALUES (?, ?, ?, ?, ?, 'https://cdn.bilibili.com/ft/folder.jpg', 0, 'FT本地收藏', ?, ?, 0, ?, 1, 0, ?, ?)`,
		id, mediaID, mediaID, ftMidSelf, title, 1357000000, mtime, mediaCount, folderType, mtime)
	t.Cleanup(func() {
		ftExecDB(t, fav, `DELETE FROM favorites_folder WHERE media_id = ?`, mediaID)
	})
}

func ftFavoriteContentRow(t *testing.T, id, mediaID, contentID int64, bvid, title string, favTime int64) {
	t.Helper()
	fav := database.GetFavoritesDB()
	if fav == nil {
		t.Skip("favorites database unavailable")
	}
	ftExecDB(t, fav, `INSERT OR REPLACE INTO favorites_content
		(id, media_id, content_id, type, title, cover, bvid, duration, upper_mid, fav_time, fetch_time, creator_name)
		VALUES (?, ?, ?, 2, ?, 'https://cdn.bilibili.com/ft/c.jpg', ?, 90, ?, ?, 1357000000, ?)`,
		id, mediaID, contentID, title, bvid, ftMidOther, favTime, ftUpName)
	t.Cleanup(func() {
		ftExecDB(t, fav, `DELETE FROM favorites_content WHERE id = ?`, id)
	})
}

func ftSeedLikes(t *testing.T) {
	t.Helper()
	likes := database.GetLikesDB()
	if likes == nil {
		t.Skip("likes database unavailable")
	}
	ftExecDB(t, likes, `INSERT OR REPLACE INTO liked_videos
		(bvid, aid, title, pic, desc, duration, tid, tname, owner_name, owner_mid, owner_face, pubdate, view, danmaku, like_count, link, fetch_time, is_seen)
		VALUES (?, ?, 'FT本地点赞一', 'https://cdn.bilibili.com/ft/l1.jpg', '赞一', 120, 18, 'FT分区', 'FT点赞UP', 2013900010, 'https://cdn.bilibili.com/ft/up1.jpg', ?, 5000, 20, 300, ?, ?, 0)`,
		ftLikeA, ftLikeAID, 1357000100, "https://www.bilibili.com/video/"+ftLikeA, 1357000200)
	ftExecDB(t, likes, `INSERT OR REPLACE INTO liked_videos
		(bvid, aid, title, pic, desc, duration, tid, tname, owner_name, owner_mid, owner_face, pubdate, view, danmaku, like_count, link, fetch_time, is_seen)
		VALUES (?, ?, 'FT本地点赞二', 'https://cdn.bilibili.com/ft/l2.jpg', '赞二', 60, 17, 'FT分区二', 'FT点赞UP二', 2013900011, '', ?, 8000, 40, 600, ?, ?, 1)`,
		ftLikeB, ftLikeBID, 1357000101, "https://www.bilibili.com/video/"+ftLikeB, 1357000201)
	t.Cleanup(func() {
		ftExecDB(t, likes, `DELETE FROM liked_videos WHERE bvid IN (?, ?)`, ftLikeA, ftLikeB)
	})
}

func ftWatchLaterRow(t *testing.T, bvid string, aid, addAt int64, title string) {
	t.Helper()
	wl := database.GetWatchLaterDB()
	if wl == nil {
		t.Skip("watchlater database unavailable")
	}
	ftExecDB(t, wl, `INSERT OR REPLACE INTO watchlater_videos
		(bvid, aid, title, pic, desc, duration, tid, tname, owner_name, owner_mid, owner_face, add_at, pubdate, view, danmaku, link, fetch_time)
		VALUES (?, ?, ?, '', '', 60, 1, 'FT稍后分区', 'FT稍后UP', 2013900020, '', ?, ?, 42, 3, ?, 1357000000)`,
		bvid, aid, title, addAt, addAt-10, "https://www.bilibili.com/video/"+bvid)
	t.Cleanup(func() {
		ftExecDB(t, wl, `DELETE FROM watchlater_videos WHERE bvid = ?`, bvid)
	})
}

// ftDynItem writes a dynamics row plus its host record.
func ftDynItem(t *testing.T, idStr, hostMid, dynType, bvid, title string, publishTS int64, mediaLocals string) {
	t.Helper()
	dynDB := database.GetDynamicDB()
	if dynDB == nil {
		t.Skip("dynamic database unavailable")
	}
	ftExecDB(t, dynDB, `INSERT OR REPLACE INTO dynamics
		(id_str, type, host_mid, author_name, author_face, txt, opus_title, opus_summary_text, bvid, title,
		 desc, cover, publish_ts, media_locals, live_media_locals, raw_json, fetch_time)
		VALUES (?, ?, ?, ?, '', 'FT动态正文', '', '', ?, ?, 'FT简介', 'https://cdn.bilibili.com/ft/dyn.jpg', ?, ?, '[]', '', ?)`,
		idStr, dynType, hostMid, ftUpName, bvid, title, publishTS, mediaLocals, 1357000000)
	t.Cleanup(func() {
		ftExecDB(t, dynDB, `DELETE FROM dynamics WHERE id_str = ?`, idStr)
	})
}

func ftSeedDynHost(t *testing.T, hostMid, upName string, itemCount, coreCount int, lastPublishTS int64) {
	t.Helper()
	if err := database.SaveDynamicHost(database.DynamicHost{
		HostMid: hostMid, UpName: upName, FacePath: "https://cdn.bilibili.com/ft/face.jpg",
		ItemCount: itemCount, CoreCount: coreCount, LastPublishTS: lastPublishTS, LastFetchTime: 1357000000,
	}); err != nil {
		t.Fatalf("seed dynamic host %s: %v", hostMid, err)
	}
	t.Cleanup(func() {
		if db := database.GetDynamicDB(); db != nil {
			ftExecDB(t, db, `DELETE FROM dynamic_hosts WHERE host_mid = ?`, hostMid)
		}
	})
}

// ftEmptyJSONList reports whether a decoded list field carries no entries. The
// local read helpers return a nil slice for an empty window, which encodes as
// JSON null, while a clamped paging window yields an empty array, so both
// shapes mean "nothing here".
func ftEmptyJSONList(v interface{}) bool {
	switch l := v.(type) {
	case nil:
		return true
	case []interface{}:
		return len(l) == 0
	default:
		return false
	}
}

// ftNum renders a value decoded from JSON the same way the fixture constants
// are written. encoding/json turns every JSON number into a float64, so a plain
// fmt.Sprint(id) would print 2.0132011e+07 and never match "20132011"; integral
// floats are therefore converted back to int64 first. Booleans and strings pass
// through unchanged.
func ftNum(v interface{}) string {
	switch n := v.(type) {
	case float64:
		if n == math.Trunc(n) && math.Abs(n) < 1<<63 {
			return strconv.FormatInt(int64(n), 10)
		}
		return strconv.FormatFloat(n, 'f', -1, 64)
	case json.Number:
		return n.String()
	default:
		return fmt.Sprint(v)
	}
}

// ftListEntry finds a JSON array element whose key holds the wanted value.
func ftListEntry(t *testing.T, raw interface{}, key, want string) map[string]interface{} {
	t.Helper()
	list, ok := raw.([]interface{})
	if !ok {
		t.Fatalf("value = %T, want a JSON array", raw)
	}
	for _, item := range list {
		entry, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		if ftNum(entry[key]) == want {
			return entry
		}
	}
	t.Fatalf("no entry with %s=%q in %v", key, want, list)
	return nil
}

func ftHasEntry(raw interface{}, key, want string) bool {
	list, ok := raw.([]interface{})
	if !ok {
		return false
	}
	for _, item := range list {
		entry, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		if ftNum(entry[key]) == want {
			return true
		}
	}
	return false
}

// ftIndexOfEntry returns the position of the first entry matching key=want, or
// -1. Used to assert relative ordering without depending on foreign rows.
func ftIndexOfEntry(raw interface{}, key, want string) int {
	list, ok := raw.([]interface{})
	// An explicit JSON null prunes the slice to nil, which is still empty.
	if !ok {
		if raw == nil {
			return -1
		}
		return -2
	}
	for i, item := range list {
		entry, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		if ftNum(entry[key]) == want {
			return i
		}
	}
	return -1
}

// ---------------------------------------------------------------------------
// credentials gating
// ---------------------------------------------------------------------------

// The deal endpoints check SESSDATA (and bili_jct) before they bind the body, so
// the request-validation branches are only reachable with credentials in place.
// ftStartStub repoints every endpoint at the local server as well, which keeps a
// request that slips past validation inside the process.
func TestFTFavoriteResourceDealNeedsMembership(t *testing.T) {
	stub := ftStartStub(t)
	e := newAPI(t, RegisterFavoriteRoutes)

	t.Run("missing resources", func(t *testing.T) {
		ftCreds(t)
		resp := expectStatus(t, e, "POST", "/api/favorite/resource/deal", `{}`, 400, "error")
		expectMessageContains(t, resp, "缺少必要参数")
	})
	t.Run("malformed json", func(t *testing.T) {
		ftCreds(t)
		resp := expectStatus(t, e, "POST", "/api/favorite/resource/deal", `{"resources":`, 400, "error")
		expectMessageContains(t, resp, "参数错误")
	})
	t.Run("no credentials", func(t *testing.T) {
		ftNoCreds(t)
		resp := expectStatus(t, e, "POST", "/api/favorite/resource/deal", `{"resources":"1:2","media_ids":"3"}`, 200, "error")
		expectMessage(t, resp, "未配置 SESSDATA，无法操作收藏")
	})
	t.Run("validation never reaches the api", func(t *testing.T) {
		ftCreds(t)
		expectStatus(t, e, "POST", "/api/favorite/resource/deal", `{}`, 400, "error")
		if got := stub.hitsFor(ftPathFavoriteDeal); got != 0 {
			t.Fatalf("deal requests = %d, want none for an invalid body", got)
		}
	})
}

func TestFTBatchFavoriteAndToggleValidation(t *testing.T) {
	stub := ftStartStub(t)
	e := newAPI(t, RegisterFavoriteRoutes)

	t.Run("batch needs both fields", func(t *testing.T) {
		ftCreds(t)
		resp := expectStatus(t, e, "POST", "/api/favorite/resource/batch-deal", `{"resources":"1:2"}`, 400, "error")
		expectMessageContains(t, resp, "resources 和 media_ids 不能为空")
	})
	t.Run("batch malformed json", func(t *testing.T) {
		ftCreds(t)
		resp := expectStatus(t, e, "POST", "/api/favorite/resource/batch-deal", `{"media_ids":`, 400, "error")
		expectMessageContains(t, resp, "参数错误")
	})
	t.Run("batch no credentials", func(t *testing.T) {
		ftNoCreds(t)
		resp := expectStatus(t, e, "POST", "/api/favorite/resource/batch-deal", `{"resources":"1:2","media_ids":"3"}`, 200, "error")
		expectMessage(t, resp, "未配置 SESSDATA，无法操作收藏")
	})

	// ToggleLikeRequest marks bvid as required (features.go:957), so an empty
	// body is a 400 before the network is ever touched. bili_jct has to be set
	// for the guard to let the request reach that bind step.
	t.Run("toggle missing bvid", func(t *testing.T) {
		ftCreds(t)
		resp := expectStatus(t, e, "POST", "/api/like/toggle", `{"like":true}`, 400, "error")
		expectMessageContains(t, resp, "参数错误")
	})
	t.Run("toggle malformed json", func(t *testing.T) {
		ftCreds(t)
		resp := expectStatus(t, e, "POST", "/api/like/toggle", `{"bvid":`, 400, "error")
		expectMessageContains(t, resp, "参数错误")
	})
	t.Run("toggle no credentials", func(t *testing.T) {
		ftNoCreds(t)
		resp := expectStatus(t, e, "POST", "/api/like/toggle", `{"bvid":"BVFTLK0000000001","like":true}`, 200, "error")
		expectMessage(t, resp, "未配置 SESSDATA / bili_jct，无法点赞")
	})
	t.Run("validation never reaches the api", func(t *testing.T) {
		ftCreds(t)
		expectStatus(t, e, "POST", "/api/favorite/resource/batch-deal", `{"resources":"1:2"}`, 400, "error")
		expectStatus(t, e, "POST", "/api/like/toggle", `{"like":true}`, 400, "error")
		if got := stub.hitsFor(ftPathFavoriteDeal); got != 0 {
			t.Fatalf("deal requests = %d, want none for an invalid body", got)
		}
		if got := stub.hitsFor(ftPathLike); got != 0 {
			t.Fatalf("like requests = %d, want none for an invalid body", got)
		}
	})
}

func TestFTBatchCheckFavoriteStatusValidation(t *testing.T) {
	e := newAPI(t, RegisterFavoriteRoutes)

	t.Run("malformed json", func(t *testing.T) {
		resp := expectStatus(t, e, "POST", "/api/favorite/check/batch", `{"oids":`, 400, "error")
		expectMessageContains(t, resp, "参数错误")
	})

	// oids may be a comma separated string or an array; unparsable parts are
	// skipped silently and an unknown shape yields an empty result set.
	cases := []struct{ name, body string }{
		{"empty object", `{}`},
		{"numeric array", `{"oids":[20132001,20132002]}`},
		{"string array", `{"oids":["20132001","20132002"]}`},
		{"mixed array", `{"oids":[20132001,"20132002","not-a-number"]}`},
		{"comma string", `{"oids":"20132001, 20132002 ,,oops"}`},
		{"object oids", `{"oids":{"a":1}}`},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			resp := expectStatus(t, e, "POST", "/api/favorite/check/batch", tc.body, 200, "success")
			data := resp.dataMap(t)
			if _, ok := data["results"]; !ok {
				t.Fatalf("payload has no results: %v", data)
			}
			if got := len(dataArray(t, data["results"])); got > 2 {
				t.Fatalf("%s: results = %v, want at most the two valid oids", tc.name, data["results"])
			}
		})
	}
}

// ---------------------------------------------------------------------------
// favorites: local reads
// ---------------------------------------------------------------------------

func TestFTLocalFavoriteFolderReads(t *testing.T) {
	e := newAPI(t, RegisterFavoriteRoutes)
	ftSeedFavorites(t)
	fav := database.GetFavoritesDB()
	if fav == nil {
		t.Skip("favorites database unavailable")
	}

	// getLocalFavoriteFolders reads every created folder.
	resp := expectStatus(t, e, "GET", "/api/favorite/local/list", "", 200, "success").dataMap(t)
	for _, key := range []string{"list", "total", "page", "size"} {
		if _, ok := resp[key]; !ok {
			t.Fatalf("favorite/local/list payload missing %q: %v", key, resp)
		}
	}
	if int(resp["page"].(float64)) != 1 || int(resp["size"].(float64)) != 20 {
		t.Fatalf("default paging = page %v size %v, want 1/20", resp["page"], resp["size"])
	}
	total := int(resp["total"].(float64))
	if want := ftCount(t, fav, `SELECT COUNT(*) FROM favorites_folder`); total != want {
		t.Fatalf("favorite/local/list total = %d, want the %d stored folders", total, want)
	}
	if int(resp["size"].(float64)) != 20 {
		t.Fatalf("favorite/local/list size = %v, want the 20 default", resp["size"])
	}

	// pn/ps are accepted aliases and page two must return the tail of the list.
	aliased := expectStatus(t, e, "GET", "/api/favorite/local/list?pn=2&ps=2", "", 200, "success").dataMap(t)
	if int(aliased["page"].(float64)) != 2 || int(aliased["size"].(float64)) != 2 {
		t.Fatalf("aliased paging = %v/%v, want 2/2", aliased["page"], aliased["size"])
	}

	// A page past the end clamps the window instead of panicking
	// (features.go:166-174).
	overflow := expectStatus(t, e, "GET", "/api/favorite/local/list?page=99999&size=10", "", 200, "success").dataMap(t)
	if !ftEmptyJSONList(overflow["list"]) {
		t.Fatalf("overflow page list = %v, want no entries", overflow["list"])
	}
}

func TestFTLocalCollectedFoldersAndContents(t *testing.T) {
	e := newAPI(t, RegisterFavoriteRoutes)
	ftFavoriteFolderRow(t, ftUnmatchedFID, ftSeasonFolder, "FT收藏的合集", 1, 2, 1357000200)
	ftFavoriteContentRow(t, 2013910001, ftSeasonFolder, ftSeasonA, ftSeasonV1, "FT合集内容一", 1357000300)
	ftFavoriteContentRow(t, 2013910002, ftSeasonFolder, ftSeasonB, ftSeasonV2, "FT合集内容二", 1357000301)

	// getLocalCollectedFolders -> database.GetFavoriteFoldersByType(1, page, size)
	collected := expectStatus(t, e, "GET", "/api/favorite/collected/local/list?page=1&size=5", "", 200, "success").dataMap(t)
	if !ftHasEntry(collected["list"], "title", "FT收藏的合集") {
		t.Fatalf("collected list = %v, want the seeded collected folder", collected["list"])
	}
	if int(collected["total"].(float64)) < 1 {
		t.Fatalf("collected total = %v, want the seeded folder counted", collected["total"])
	}

	// getLocalFavoriteContents orders by fav_time DESC.
	contents := expectStatus(t, e, "GET", fmt.Sprintf("/api/favorite/content/list?media_id=%d&page=1&size=5", ftSeasonFolder), "", 200, "success").dataMap(t)
	if int(contents["total"].(float64)) != 2 {
		t.Fatalf("contents total = %v, want the two seeded contents", contents["total"])
	}
	first := ftListEntry(t, contents["list"], "bvid", ftSeasonV2)
	if first["title"] != "FT合集内容二" {
		t.Fatalf("first content = %v, want the newest fav_time", first)
	}

	// The second page holds the older row.
	second := expectStatus(t, e, "GET", fmt.Sprintf("/api/favorite/content/list?media_id=%d&page=2&size=1", ftSeasonFolder), "", 200, "success").dataMap(t)
	if len(dataArray(t, second["list"])) != 1 {
		t.Fatalf("contents page 2 = %v, want a single row", second["list"])
	}
	if !ftHasEntry(second["list"], "bvid", ftSeasonV1) {
		t.Fatalf("contents page 2 = %v, want %s", second["list"], ftSeasonV1)
	}

	// pn/ps aliases resolve the same window; an unparsable page falls back to
	// the pn default of 1 rather than failing the request (features.go:247-258).
	fallback := expectStatus(t, e, "GET", fmt.Sprintf("/api/favorite/content/list?media_id=%d&page=abc&size=xyz", ftSeasonFolder), "", 200, "success").dataMap(t)
	if int(fallback["page"].(float64)) != 1 || int(fallback["size"].(float64)) != 20 {
		t.Fatalf("fallback paging = %v/%v, want 1/20", fallback["page"], fallback["size"])
	}

	// An unknown media id yields an empty window with total 0.
	missing := expectStatus(t, e, "GET", "/api/favorite/content/list?media_id=20991234", "", 200, "success").dataMap(t)
	if int(missing["total"].(float64)) != 0 || !ftEmptyJSONList(missing["list"]) {
		t.Fatalf("unknown media payload = %v, want total 0 and an empty list", missing)
	}
}

// TestFTBatchCheckFavoriteStatusFolderJoin pins down the folder-title lookup of
// batchCheckFavoriteStatus, whose query is
//
//	SELECT DISTINCT fc.media_id, ff.title
//	FROM favorites_content fc LEFT JOIN favorites_folder ff ON fc.media_id = ff.id
//	WHERE CAST(fc.bvid AS INTEGER) = ?
//
// i.e. the oid is compared against a numeric bvid column and the title comes
// from a join on the folder rowid.
func TestFTBatchCheckFavoriteStatusFolderJoin(t *testing.T) {
	fav := database.GetFavoritesDB()
	if fav == nil {
		t.Skip("favorites database unavailable")
	}

	// Case one: a folder written with id == media_id, so the join matches and the
	// title resolves (batchCheckFavoriteStatus matches on
	// CAST(fc.bvid AS INTEGER), hence the numeric bvid).
	ftFavoriteFolderRow(t, ftSeasonFolder, ftSeasonFolder, "FT编号收藏夹", 1, 1, 1357000200)
	ftFavoriteContentRow(t, 2013910003, ftSeasonFolder, 2013910003, fmt.Sprint(ftSeasonA), "FT编号内容", 1357000302)

	e := newAPI(t, RegisterFavoriteRoutes)
	resp := expectStatus(t, e, "POST", "/api/favorite/check/batch", fmt.Sprintf(`{"oids":[%d]}`, ftSeasonA), 200, "success").dataMap(t)
	result := ftListEntry(t, resp["results"], "oid", fmt.Sprint(ftSeasonA))
	if result["is_favorited"] != true {
		t.Fatalf("result = %v, want the numeric bvid row to be recognised", result)
	}
	folders := dataArray(t, result["favorite_folders"])
	if len(folders) != 1 {
		t.Fatalf("favorite_folders = %v, want exactly the seeded folder", folders)
	}
	entry := folders[0].(map[string]interface{})
	if int64(entry["media_id"].(float64)) != ftSeasonFolder {
		t.Fatalf("folder media_id = %v, want %d", entry["media_id"], ftSeasonFolder)
	}
	if entry["title"] != "FT编号收藏夹" {
		t.Fatalf("folder title = %v, want the joined folder title", entry["title"])
	}

	// Case two: the row shape database.SaveFavoriteFolders really writes - the
	// autoincrement id differs from media_id.
	// 回归 features.go 的 join 修复：过去它按 ff.id（自增 rowid）连接，而入库
	// 时从不显式给 id（database/extras.go:408），于是每次真实同步都连不上；
	// LEFT JOIN 出来的 NULL title 又让非空 `var title string` 的 Scan 报错、整行
	// 被丢弃，结果是明明在收藏夹里的视频回答 is_favorited=false。现在按
	// fc.media_id = ff.media_id 连接，并把 title 按可空列扫描。
	ftFavoriteFolderRow(t, ftUnmatchedFID, ftCollectedPlain, "FT收藏的合集", 1, 1, 1357000201)
	ftFavoriteContentRow(t, 2013910004, ftCollectedPlain, 2013910004, fmt.Sprint(ftSeasonB), "FT编号内容二", 1357000303)
	unmatched := expectStatus(t, e, "POST", "/api/favorite/check/batch", fmt.Sprintf(`{"oids":[%d]}`, ftSeasonB), 200, "success").dataMap(t)
	fixed := ftListEntry(t, unmatched["results"], "oid", fmt.Sprint(ftSeasonB))
	if fixed["is_favorited"] != true {
		t.Fatalf("folder-id join result = %v, want the row recognised after the media_id join fix", fixed)
	}
	joined := dataArray(t, fixed["favorite_folders"])
	if len(joined) != 1 {
		t.Fatalf("favorite_folders = %v, want exactly the seeded folder", joined)
	}
	je := joined[0].(map[string]interface{})
	if int64(je["media_id"].(float64)) != ftCollectedPlain || je["title"] != "FT收藏的合集" {
		t.Fatalf("joined folder = %v, want media_id=%d title=FT收藏的合集", je, ftCollectedPlain)
	}
	// The content row itself is there, so only the join is at fault.
	if n := ftCount(t, fav, `SELECT COUNT(*) FROM favorites_content WHERE media_id = ? AND content_id = ?`, ftCollectedPlain, 2013910004); n != 1 {
		t.Fatalf("content rows for the joined media_id = %d, want the seeded row", n)
	}

	// Case three: a content row whose folder is not in favorites_folder. The
	// LEFT JOIN then yields a NULL title, which must not discard the row (it
	// used to fail the non-nullable scan and lose the is_favorited hit).
	ftFavoriteContentRow(t, 2013910005, ftOrphanMediaID, 2013910005, fmt.Sprint(ftOrphanOid), "FT孤儿内容", 1357000304)
	orphan := expectStatus(t, e, "POST", "/api/favorite/check/batch", fmt.Sprintf(`{"oids":[%d]}`, ftOrphanOid), 200, "success").dataMap(t)
	nullTitle := ftListEntry(t, orphan["results"], "oid", fmt.Sprint(ftOrphanOid))
	if nullTitle["is_favorited"] != true {
		t.Fatalf("orphan-media result = %v, want is_favorited true with a NULL folder title", nullTitle)
	}
	orphanFolders := dataArray(t, nullTitle["favorite_folders"])
	if len(orphanFolders) != 1 {
		t.Fatalf("orphan favorite_folders = %v, want one entry with an empty title", orphanFolders)
	}
	of := orphanFolders[0].(map[string]interface{})
	if int64(of["media_id"].(float64)) != ftOrphanMediaID || of["title"] != "" {
		t.Fatalf("orphan folder = %v, want media_id=%d title=\"\"", of, ftOrphanMediaID)
	}

	// An oid with no matching content row is reported as not favorited.
	other := expectStatus(t, e, "POST", "/api/favorite/check/batch", `{"oids":[2013999999]}`, 200, "success").dataMap(t)
	miss := ftListEntry(t, other["results"], "oid", "2013999999")
	if miss["is_favorited"] != false || len(dataArray(t, miss["favorite_folders"])) != 0 {
		t.Fatalf("unknown oid result = %v, want is_favorited false with no folders", miss)
	}
}

// ---------------------------------------------------------------------------
// favorites: online reads
// ---------------------------------------------------------------------------

func TestFTFavoriteListAndCollectedFolders(t *testing.T) {
	stub := ftStartStub(t)
	ftCreds(t)
	e := newAPI(t, RegisterFavoriteRoutes)

	// getFavoriteList proxies the created-folder list.
	resp := expectStatus(t, e, "GET", "/api/favorite/list", "", 200, "success").dataMap(t)
	list := dataArray(t, resp["list"])
	if len(list) != 3 {
		t.Fatalf("favorite/list length = %d, want the three stub folders", len(list))
	}
	if int(resp["total"].(float64)) != 3 {
		t.Fatalf("favorite/list total = %v, want the stub count", resp["total"])
	}
	folder := ftListEntry(t, list, "title", "FT同步收藏夹")
	if ftNum(folder["id"]) != fmt.Sprint(ftFolderCreated1) {
		t.Fatalf("folder id = %v, want %d", folder["id"], ftFolderCreated1)
	}
	// up_mid is taken from the configured DedeUserID.
	if got := stub.lastQuery(ftPathCreatedList).Get("up_mid"); got != fmt.Sprint(ftMidSelf) {
		t.Fatalf("created list up_mid = %q, want the configured %d", got, ftMidSelf)
	}

	// getCollectedFavoriteFolders returns everything when no keyword is set.
	collected := expectStatus(t, e, "GET", "/api/favorite/collected/list", "", 200, "success").dataMap(t)
	if len(dataArray(t, collected["list"])) != 2 {
		t.Fatalf("collected list = %v, want both stub folders", collected["list"])
	}
	if collected["has_more"] != false {
		t.Fatalf("collected has_more = %v, want the hard coded false", collected["has_more"])
	}

	// The keyword filter matches title or intro (features.go:227-232).
	byTitle := expectStatus(t, e, "GET", "/api/favorite/collected/list?keyword=合集", "", 200, "success").dataMap(t)
	if len(dataArray(t, byTitle["list"])) != 1 || !ftHasEntry(byTitle["list"], "title", "FT收藏的合集") {
		t.Fatalf("title filter = %v, want only the 合集 folder", byTitle["list"])
	}
	byIntro := expectStatus(t, e, "GET", "/api/favorite/collected/list?keyword=空", "", 200, "success").dataMap(t)
	if !ftHasEntry(byIntro["list"], "title", "FT收藏的空收藏夹") {
		t.Fatalf("intro filter = %v, want the folder whose intro matches", byIntro["list"])
	}
	// Nothing matches: the slice stays nil, so data.list is serialised as null
	// rather than an empty array (features.go:226).
	none := expectStatus(t, e, "GET", "/api/favorite/collected/list?keyword=FT无此关键字", "", 200, "success").dataMap(t)
	if none["list"] != nil {
		t.Fatalf("keyword miss list = %v, want null", none["list"])
	}

	// Cookie expiry is reported with its own message (features.go:217-220).
	stub.setCode(ftPathCollectedList, -6)
	expired := expectStatus(t, e, "GET", "/api/favorite/collected/list", "", 200, "error")
	expectMessage(t, expired, "Cookie 已过期，请重新登录")
	stub.setCode(ftPathCollectedList, -352)
	blocked := expectStatus(t, e, "GET", "/api/favorite/collected/list", "", 200, "error")
	expectMessageContains(t, blocked, "获取收藏的收藏夹失败")
	expectMessageContains(t, blocked, "-352")
	stub.setCode(ftPathCollectedList, 0)

	stub.setCode(ftPathCreatedList, -6)
	expired = expectStatus(t, e, "GET", "/api/favorite/list", "", 200, "error")
	expectMessage(t, expired, "Cookie 已过期，请重新登录")
	stub.setCode(ftPathCreatedList, -101)
	failed := expectStatus(t, e, "GET", "/api/favorite/list", "", 200, "error")
	expectMessageContains(t, failed, "获取收藏夹列表失败")
	stub.setCode(ftPathCreatedList, 0)
}

func TestFTOnlineFavoriteContents(t *testing.T) {
	stub := ftStartStub(t)
	ftCreds(t)
	e := newAPI(t, RegisterFavoriteRoutes)

	// A plain folder goes through GetFavoriteResources.
	resp := expectStatus(t, e, "GET", fmt.Sprintf("/api/favorite/content/online?media_id=%d&pn=1&ps=20", ftFolderCreated1), "", 200, "success").dataMap(t)
	list := dataArray(t, resp["list"])
	if len(list) != 1 {
		t.Fatalf("resource list = %v, want the single stub media", list)
	}
	entry := list[0].(map[string]interface{})
	if entry["bvid"] != ftFavVideo1 || ftNum(entry["id"]) != fmt.Sprint(ftContent1) {
		t.Fatalf("resource entry = %v, want bvid %s and id %d", entry, ftFavVideo1, ftContent1)
	}
	if ftNum(entry["fav_time"]) != "1357000020" || ftNum(entry["pub_time"]) != "1357000001" {
		t.Fatalf("resource timestamps = %v", entry)
	}
	if entry["intro"] != "内容一" {
		t.Fatalf("resource intro = %v, want the stub intro", entry["intro"])
	}
	upper := entry["upper_mid"]
	if ftNum(upper) != fmt.Sprint(ftMidOther) {
		t.Fatalf("upper_mid = %v, want %d", upper, ftMidOther)
	}
	if ftNum(resp["total"]) != "1" {
		t.Fatalf("resource total = %v, want the page count", resp["total"])
	}
	info := resp["info"].(map[string]interface{})
	if info["title"] != "FT在线内容收藏夹" || ftNum(info["media_count"]) != "1" {
		t.Fatalf("resource info = %v, want the stub folder header", info)
	}

	// A media with ugc:null has no bvid to report (features.go:351-354).
	withNull := expectStatus(t, e, "GET", fmt.Sprintf("/api/favorite/content/online?media_id=%d", ftFolderCreated2), "", 200, "success").dataMap(t)
	items := dataArray(t, withNull["list"])
	if len(items) != 2 {
		t.Fatalf("second folder list = %v, want both stub medias", items)
	}
	invalid := ftListEntry(t, items, "title", "已失效视频")
	if invalid["bvid"] != "" {
		t.Fatalf("失效视频 bvid = %v, want the empty string from the nil ugc", invalid["bvid"])
	}

	// season_id switches to the season API (features.go:296-336).
	season := expectStatus(t, e, "GET", fmt.Sprintf("/api/favorite/content/online?season_id=%d&pn=1&ps=100", ftSeasonFolder), "", 200, "success").dataMap(t)
	seasonItems := dataArray(t, season["list"])
	if len(seasonItems) != 2 {
		t.Fatalf("season list = %v, want the two stub episodes", seasonItems)
	}
	first := seasonItems[0].(map[string]interface{})
	if first["bvid"] != ftSeasonV1 || first["title"] != "FT合集内容一" {
		t.Fatalf("season entry = %v, want the first stub episode", first)
	}
	if first["upper"].(map[string]interface{})["name"] != "FT内容UP" {
		t.Fatalf("season upper = %v", first["upper"])
	}
	if ftNum(season["total"]) != "2" {
		t.Fatalf("season total = %v, want info.media_count", season["total"])
	}
	if ftNum(season["page"]) != "1" || ftNum(season["size"]) != "100" {
		t.Fatalf("season paging = %v/%v, want the echoed pn/ps", season["page"], season["size"])
	}

	// Error branches for both variants.
	stub.setCode(ftPathResourceList, -6)
	expired := expectStatus(t, e, "GET", fmt.Sprintf("/api/favorite/content/online?media_id=%d", ftFolderCreated1), "", 200, "error")
	expectMessage(t, expired, "Cookie 已过期，请重新登录")
	stub.setCode(ftPathResourceList, -403)
	forbidden := expectStatus(t, e, "GET", fmt.Sprintf("/api/favorite/content/online?media_id=%d", ftFolderCreated1), "", 200, "error")
	expectMessageContains(t, forbidden, "获取收藏夹内容失败")
	stub.setCode(ftPathResourceList, 0)

	stub.setCode(ftPathSeasonList, -6)
	expired = expectStatus(t, e, "GET", fmt.Sprintf("/api/favorite/content/online?season_id=%d", ftSeasonFolder), "", 200, "error")
	expectMessage(t, expired, "Cookie 已过期，请重新登录")
	stub.setCode(ftPathSeasonList, -101)
	failed := expectStatus(t, e, "GET", fmt.Sprintf("/api/favorite/content/online?season_id=%d", ftSeasonFolder), "", 200, "error")
	expectMessageContains(t, failed, "获取合集内容失败")
	stub.setCode(ftPathSeasonList, 0)
}

// ---------------------------------------------------------------------------
// favorites: write operations
// ---------------------------------------------------------------------------

func TestFTFavoriteResourceDeal(t *testing.T) {
	stub := ftStartStub(t)
	ftCreds(t)
	e := newAPI(t, RegisterFavoriteRoutes)

	// The resources/media_ids shorthand form.
	resp := expectStatus(t, e, "POST", "/api/favorite/resource/deal",
		fmt.Sprintf(`{"resources":"%d:2","media_ids":"%d"}`, ftFolderCreated1, ftContent1), 200, "success").dataMap(t)
	if resp["message"] != "操作成功" {
		t.Fatalf("deal response = %v, want 操作成功", resp)
	}
	form := stub.lastForm(ftPathFavoriteDeal)
	if form.Get("resources") != fmt.Sprintf("%d:2", ftFolderCreated1) {
		t.Fatalf("deal resources = %q", form.Get("resources"))
	}
	if form.Get("csrf") != "ft-jct-2013" {
		t.Fatalf("deal csrf = %q, want the configured bili_jct", form.Get("csrf"))
	}

	// The rid/add_media_ids form defaults to media type 2 (features.go:416-421).
	before := stub.formCount(ftPathFavoriteDeal)
	added := expectStatus(t, e, "POST", "/api/favorite/resource/deal",
		fmt.Sprintf(`{"rid":%d,"add_media_ids":"%d"}`, ftFolderCreated2, ftContent2), 200, "success")
	if added.Status != "success" {
		t.Fatalf("rid form response = %v", added)
	}
	if got := stub.lastForm(ftPathFavoriteDeal).Get("resources"); got != fmt.Sprintf("%d:2", ftFolderCreated2) {
		t.Fatalf("rid form resources = %q, want the default type 2", got)
	}

	// An explicit type wins over the default.
	expectStatus(t, e, "POST", "/api/favorite/resource/deal",
		fmt.Sprintf(`{"rid":%d,"type":12,"del_media_ids":"%d"}`, ftFolderCreated2, ftContent3), 200, "success")
	if got := stub.lastForm(ftPathFavoriteDeal).Get("resources"); got != fmt.Sprintf("%d:12", ftFolderCreated2) {
		t.Fatalf("typed resources = %q, want the supplied type 12", got)
	}
	if stub.formCount(ftPathFavoriteDeal) != before+2 {
		t.Fatalf("deal calls = %d, want the two rid-form posts", stub.formCount(ftPathFavoriteDeal))
	}

	// API failures map onto the two error messages.
	stub.setCode(ftPathFavoriteDeal, -6)
	expired := expectStatus(t, e, "POST", "/api/favorite/resource/deal",
		fmt.Sprintf(`{"resources":"%d:2","media_ids":"%d"}`, ftFolderCreated1, ftContent1), 200, "error")
	expectMessage(t, expired, "Cookie 已过期，请重新登录")
	stub.setCode(ftPathFavoriteDeal, -403)
	denied := expectStatus(t, e, "POST", "/api/favorite/resource/deal",
		fmt.Sprintf(`{"rid":%d,"add_media_ids":"%d"}`, ftFolderCreated2, ftContent2), 200, "error")
	expectMessageContains(t, denied, "收藏操作失败")
	stub.setCode(ftPathFavoriteDeal, 0)
}

func TestFTBatchFavoriteResource(t *testing.T) {
	stub := ftStartStub(t)
	ftCreds(t)
	e := newAPI(t, RegisterFavoriteRoutes)

	resp := expectStatus(t, e, "POST", "/api/favorite/resource/batch-deal",
		fmt.Sprintf(`{"resources":"%d:2,%d:2","media_ids":"%d"}`, ftFolderCreated1, ftFolderCreated2, ftContent1), 200, "success").dataMap(t)
	if resp["message"] != "操作成功" {
		t.Fatalf("batch-deal response = %v, want 操作成功", resp)
	}
	form := stub.lastForm(ftPathFavoriteDeal)
	if !strings.Contains(form.Get("resources"), fmt.Sprint(ftFolderCreated2)) {
		t.Fatalf("batch-deal resources = %q, want both folders", form.Get("resources"))
	}

	stub.setCode(ftPathFavoriteDeal, -6)
	expired := expectStatus(t, e, "POST", "/api/favorite/resource/batch-deal",
		fmt.Sprintf(`{"resources":"%d:2","media_ids":"%d"}`, ftFolderCreated1, ftContent1), 200, "error")
	expectMessage(t, expired, "Cookie 已过期，请重新登录")
	stub.setCode(ftPathFavoriteDeal, -412)
	blocked := expectStatus(t, e, "POST", "/api/favorite/resource/batch-deal",
		fmt.Sprintf(`{"resources":"%d:2","media_ids":"%d"}`, ftFolderCreated1, ftContent1), 200, "error")
	expectMessageContains(t, blocked, "批量收藏操作失败")
	stub.setCode(ftPathFavoriteDeal, 0)
}

// TestFTLocalBatchFavoriteResource drives the purely local add/remove branch
// pair of features.go:480.
func TestFTLocalBatchFavoriteResource(t *testing.T) {
	e := newAPI(t, RegisterFavoriteRoutes)
	fav := database.GetFavoritesDB()
	if fav == nil {
		t.Skip("favorites database unavailable")
	}

	t.Run("malformed json", func(t *testing.T) {
		resp := expectStatus(t, e, "POST", "/api/favorite/resource/local-batch-deal", `{"media_id":`, 400, "error")
		expectMessageContains(t, resp, "参数错误")
	})
	t.Run("missing fields", func(t *testing.T) {
		resp := expectStatus(t, e, "POST", "/api/favorite/resource/local-batch-deal", `{"bvids":["BVFTLB0000000001"]}`, 400, "error")
		expectMessageContains(t, resp, "media_id 和 bvids 不能为空")
		resp = expectStatus(t, e, "POST", "/api/favorite/resource/local-batch-deal",
			fmt.Sprintf(`{"media_id":%d,"bvids":[]}`, ftFolderCreated1), 400, "error")
		expectMessageContains(t, resp, "media_id 和 bvids 不能为空")
	})

	// Removing deletes the matching rows and counts the affected ones.
	ftFavoriteContentRow(t, 2013910010, ftFolderCreated1, 2013910010, "BVFTLB0000000001", "FT批量内容一", 1357000310)
	ftFavoriteContentRow(t, 2013910011, ftFolderCreated1, 2013910011, "BVFTLB0000000002", "FT批量内容二", 1357000311)
	removed := expectStatus(t, e, "POST", "/api/favorite/resource/local-batch-deal",
		fmt.Sprintf(`{"media_id":%d,"bvids":["BVFTLB0000000001","BVFTLB0000000002","BVFTLB0000000099"],"action":"remove"}`, ftFolderCreated1),
		200, "success").dataMap(t)
	if int(removed["affected"].(float64)) != 2 {
		t.Fatalf("remove affected = %v, want the two stored bvids", removed["affected"])
	}
	if left := ftCount(t, fav, `SELECT COUNT(*) FROM favorites_content WHERE media_id = ? AND bvid LIKE 'BVFTLB%'`, ftFolderCreated1); left != 0 {
		t.Fatalf("rows after remove = %d, want both deleted", left)
	}

	// Adding writes placeholder rows.
	// PRODUCTION BUG: features.go:514-523 counts one affected row per Exec
	// call, but every placeholder is written with content_id = 0 and the table
	// declares UNIQUE(media_id, content_id), so all three inserts collide and
	// only one row survives while the response claims three; using the bvid (or
	// a hash of it) as the conflict key, or accumulating RowsAffected, would
	// report the truth.
	added := expectStatus(t, e, "POST", "/api/favorite/resource/local-batch-deal",
		fmt.Sprintf(`{"media_id":%d,"bvids":["BVFTLB0000000003","BVFTLB0000000004","BVFTLB0000000005"],"action":"add"}`, ftFolderCreated2),
		200, "success").dataMap(t)
	if int(added["affected"].(float64)) != 3 {
		t.Fatalf("add affected = %v, want the buggy per-insert count of 3", added["affected"])
	}
	if rows := ftCount(t, fav, `SELECT COUNT(*) FROM favorites_content WHERE media_id = ? AND content_id = 0`, ftFolderCreated2); rows != 1 {
		t.Fatalf("placeholder rows = %d, want exactly one stored row for three bvids", rows)
	}
	t.Cleanup(func() {
		ftExecDB(t, fav, `DELETE FROM favorites_content WHERE media_id = ? AND content_id = 0`, ftFolderCreated2)
	})

	// An unknown action takes the add branch as well.
	expectStatus(t, e, "POST", "/api/favorite/resource/local-batch-deal",
		fmt.Sprintf(`{"media_id":%d,"bvids":["BVFTLB0000000006"],"action":"keep"}`, ftFolderCreated2), 200, "success")
}

// ---------------------------------------------------------------------------
// likes
// ---------------------------------------------------------------------------

func TestFTLikeReadsAndToggle(t *testing.T) {
	stub := ftStartStub(t)
	ftCreds(t)
	e := newAPI(t, RegisterFavoriteRoutes)
	ftSeedLikes(t)

	// getLikeList / getLikeLocal read the cached likes.
	resp := expectStatus(t, e, "GET", "/api/like/list?page=1&size=100&sort=pubdate&order=desc", "", 200, "success").dataMap(t)
	list := dataArray(t, resp["list"])
	first := ftListEntry(t, list, "bvid", ftLikeB)
	if first["title"] != "FT本地点赞二" {
		t.Fatalf("like list first row = %v, want the newest pubdate row", first)
	}
	if ftNum(resp["page"]) != "1" || ftNum(resp["size"]) != "100" {
		t.Fatalf("like paging = %v/%v, want the echoed query values", resp["page"], resp["size"])
	}

	// has_more compares total against page*size (features.go:846).
	narrow := expectStatus(t, e, "GET", "/api/like/list?page=1&size=1", "", 200, "success").dataMap(t)
	if int(narrow["total"].(float64)) < 2 {
		t.Fatalf("like total = %v, want the two seeded rows", narrow["total"])
	}
	if len(dataArray(t, narrow["list"])) != 1 {
		t.Fatalf("like size=1 page = %v, want a single row", narrow["list"])
	}
	overflow := expectStatus(t, e, "GET", "/api/like/list?page=99999&size=1", "", 200, "success").dataMap(t)
	if overflow["has_more"] != false {
		t.Fatalf("far page has_more = %v, want false", overflow["has_more"])
	}

	local := expectStatus(t, e, "GET", "/api/like/local", "", 200, "success").dataMap(t)
	if !ftHasEntry(local["list"], "bvid", ftLikeA) {
		t.Fatalf("local likes = %v, want %s", local["list"], ftLikeA)
	}

	// toggleLike posts the csrf signed like/unlike form.
	liked := expectStatus(t, e, "POST", "/api/like/toggle", fmt.Sprintf(`{"bvid":"%s","like":true}`, ftLikeA), 200, "success").dataMap(t)
	if liked["action"] != "点赞" || liked["bvid"] != ftLikeA || liked["success"] != true {
		t.Fatalf("toggle response = %v, want a 点赞 acknowledgement", liked)
	}
	form := stub.lastForm(ftPathLike)
	if form.Get("like") != "1" || form.Get("bvid") != ftLikeA || form.Get("csrf") != "ft-jct-2013" {
		t.Fatalf("toggle form = %v, want bvid/like/csrf", form)
	}

	unliked := expectStatus(t, e, "POST", "/api/like/toggle", fmt.Sprintf(`{"bvid":"%s","like":false}`, ftLikeA), 200, "success").dataMap(t)
	if unliked["action"] != "取消点赞" {
		t.Fatalf("unlike response = %v, want 取消点赞", unliked)
	}
	if got := stub.lastForm(ftPathLike).Get("like"); got != "0" {
		t.Fatalf("unlike form like = %q, want 0", got)
	}

	stub.setCode(ftPathLike, -6)
	expired := expectStatus(t, e, "POST", "/api/like/toggle", fmt.Sprintf(`{"bvid":"%s","like":true}`, ftLikeA), 200, "error")
	expectMessage(t, expired, "Cookie 已过期，请重新登录")
	stub.setCode(ftPathLike, 0)

	// A non -6 API error keeps the generic message.
	failed := expectStatus(t, e, "POST", "/api/like/toggle", fmt.Sprintf(`{"bvid":"%s","like":true}`, ftLikeB), 200, "error")
	expectMessageContains(t, failed, "点赞操作失败")
	expectMessageContains(t, failed, "code=-9")
}

// ---------------------------------------------------------------------------
// watch later
// ---------------------------------------------------------------------------

func TestFTWatchLaterReadsAndDeletes(t *testing.T) {
	stub := ftStartStub(t)
	ftCreds(t)
	e := newAPI(t, RegisterFavoriteRoutes)

	// getWatchLaterList refreshes the cache and returns the remote rows.
	resp := expectStatus(t, e, "GET", "/api/watchlater/list", "", 200, "success").dataMap(t)
	list := dataArray(t, resp["list"])
	if len(list) != 2 {
		t.Fatalf("watchlater list = %v, want the two stub items", list)
	}
	if ftNum(resp["total"]) != "2" {
		t.Fatalf("watchlater total = %v, want the remote count", resp["total"])
	}
	entry := ftListEntry(t, list, "bvid", ftWLA)
	if entry["title"] != "FT稍后再看一" || ftNum(entry["aid"]) != fmt.Sprint(ftWLAID) {
		t.Fatalf("watchlater entry = %v, want the stub row", entry)
	}
	if entry["link"] != "https://www.bilibili.com/video/"+ftWLA {
		t.Fatalf("watchlater link = %v, want the derived video url", entry["link"])
	}
	if entry["owner_name"] != "FT稍后UP" {
		t.Fatalf("watchlater owner = %v", entry["owner_name"])
	}

	// The same rows are readable from the local cache afterwards.
	local := expectStatus(t, e, "GET", "/api/watchlater/local?page=1&size=100", "", 200, "success").dataMap(t)
	if !ftHasEntry(local["list"], "bvid", ftWLB) {
		t.Fatalf("local watchlater = %v, want the synced rows", local["list"])
	}

	// deleteWatchLaterVideo needs a local row to learn the aid.
	missing := expectStatus(t, e, "DELETE", "/api/watchlater/BVFTWL0000000099", "", 200, "error")
	expectMessage(t, missing, "本地未找到该视频，请先同步稍后再看列表")

	del := expectStatus(t, e, "DELETE", "/api/watchlater/"+ftWLA, "", 200, "success").dataMap(t)
	if del["bvid"] != ftWLA || ftNum(del["aid"]) != fmt.Sprint(ftWLAID) {
		t.Fatalf("delete response = %v, want the bvid/aid echo", del)
	}
	if got := stub.lastForm(ftPathWatchLaterDel).Get("aid"); got != fmt.Sprint(ftWLAID) {
		t.Fatalf("delete form aid = %q, want %d", got, ftWLAID)
	}
	if got := stub.lastForm(ftPathWatchLaterDel).Get("csrf"); got != "ft-jct-2013" {
		t.Fatalf("delete form csrf = %q", got)
	}
	wl := database.GetWatchLaterDB()
	if wl == nil {
		t.Skip("watchlater database unavailable")
	}
	if n := ftCount(t, wl, `SELECT COUNT(*) FROM watchlater_videos WHERE bvid = ?`, ftWLA); n != 0 {
		t.Fatalf("local row survived the remote delete (%d)", n)
	}

	// Cookie expiry is the -6 branch.
	stub.setCode(ftPathWatchLaterDel, -6)
	expired := expectStatus(t, e, "DELETE", "/api/watchlater/"+ftWLB, "", 200, "error")
	expectMessage(t, expired, "Cookie 已过期，请重新登录")
	stub.setCode(ftPathWatchLaterDel, 0)
	if n := ftCount(t, wl, `SELECT COUNT(*) FROM watchlater_videos WHERE bvid = ?`, ftWLB); n != 1 {
		t.Fatalf("failed delete removed the local row (%d)", n)
	}

	stub.setCode(ftPathWatchLaterDel, -404)
	failed := expectStatus(t, e, "DELETE", "/api/watchlater/"+ftWLB, "", 200, "error")
	expectMessageContains(t, failed, "删除失败")
	stub.setCode(ftPathWatchLaterDel, 0)
}

func TestFTBatchDeleteWatchLaterVideos(t *testing.T) {
	stub := ftStartStub(t)
	ftCreds(t)
	e := newAPI(t, RegisterFavoriteRoutes)

	ftWatchLaterRow(t, ftWLFailBvid, ftWLFail, 1357000401, "FT批量稍后失败")
	// The bvids deleted below are seeded explicitly (and the read test above
	// deletes some of them again), so this test never depends on which other
	// test last filled the watch-later cache.
	ftWatchLaterRow(t, ftWLA, ftWLAID, 1357000402, "FT批量稍后一")
	ftWatchLaterRow(t, ftWLB, ftWLBIT, 1357000403, "FT批量稍后二")
	stub.setFailAid(ftWLFail)

	// bvids accepts a JSON array, a comma separated string, and tolerates
	// whitespace; unparsable shapes are rejected before any API call.
	t.Run("malformed json", func(t *testing.T) {
		resp := expectStatus(t, e, "POST", "/api/watchlater/batch-delete", `{"bvids":`, 400, "error")
		expectMessageContains(t, resp, "参数错误")
	})
	t.Run("wrong shape", func(t *testing.T) {
		resp := expectStatus(t, e, "POST", "/api/watchlater/batch-delete", `{"bvids":123}`, 400, "error")
		expectMessageContains(t, resp, "bvids 参数格式错误")
		resp = expectStatus(t, e, "POST", "/api/watchlater/batch-delete", `{}`, 400, "error")
		expectMessageContains(t, resp, "bvids 参数格式错误")
	})
	t.Run("empty selection", func(t *testing.T) {
		resp := expectStatus(t, e, "POST", "/api/watchlater/batch-delete", `{"bvids":["  ", ""]}`, 400, "error")
		expectMessageContains(t, resp, "未提供要删除的 bvid")
	})

	body := fmt.Sprintf(`{"bvids":["%s","%s",42,"%s"]}`, ftWLA, ftWLFailBvid, ftWLUntrackedBvid)
	resp := expectStatus(t, e, "POST", "/api/watchlater/batch-delete", body, 200, "success").dataMap(t)
	if ftNum(resp["total"]) != "3" {
		t.Fatalf("batch total = %v, want the three string bvids", resp["total"])
	}
	if ftNum(resp["success"]) != "1" || ftNum(resp["failed"]) != "2" {
		t.Fatalf("batch counters = success %v failed %v, want 1/2", resp["success"], resp["failed"])
	}
	results := dataArray(t, resp["results"])
	ok := ftListEntry(t, results, "bvid", ftWLA)
	if ok["success"] != true {
		t.Fatalf("deleted row = %v, want success true", ok)
	}
	failed := ftListEntry(t, results, "bvid", ftWLFailBvid)
	if failed["success"] == true {
		t.Fatalf("failing row = %v, want success false", failed)
	}
	missingRow := ftListEntry(t, results, "bvid", ftWLUntrackedBvid)
	if msg, _ := missingRow["error"].(string); !strings.Contains(msg, "本地未找到该视频") {
		t.Fatalf("missing row error = %q, want the local-cache message", msg)
	}

	// The comma separated form is accepted too.
	comma := expectStatus(t, e, "POST", "/api/watchlater/batch-delete",
		fmt.Sprintf(`{"bvids":"%s, %s"}`, ftWLB, ftWLUntrackedBvid), 200, "success").dataMap(t)
	if ftNum(comma["total"]) != "2" {
		t.Fatalf("comma total = %v, want two bvids", comma["total"])
	}
	if ftNum(comma["success"]) != "1" || ftNum(comma["failed"]) != "1" {
		t.Fatalf("comma counters = success %v failed %v, want the seeded row deleted and the unknown one refused",
			comma["success"], comma["failed"])
	}
	if n := ftCount(t, database.GetWatchLaterDB(), `SELECT COUNT(*) FROM watchlater_videos WHERE bvid = ?`, ftWLB); n != 0 {
		t.Fatalf("comma delete left the local %s row in place (%d)", ftWLB, n)
	}
}

const ftWLUntrackedBvid = "BVFTWL0000000099"

func TestFTWatchLaterListErrors(t *testing.T) {
	stub := ftStartStub(t)
	ftCreds(t)
	e := newAPI(t, RegisterFavoriteRoutes)

	stub.setCode(ftPathWatchLater, -6)
	expired := expectStatus(t, e, "GET", "/api/watchlater/list", "", 200, "error")
	expectMessage(t, expired, "Cookie 已过期，请重新登录")

	stub.setCode(ftPathWatchLater, -352)
	blocked := expectStatus(t, e, "GET", "/api/watchlater/list", "", 200, "error")
	expectMessageContains(t, blocked, "获取稍后再看列表失败")
	expectMessageContains(t, blocked, "-352")
	stub.setCode(ftPathWatchLater, 0)

	// With no credentials configured the remote list is refused outright.
	cfg := loadCfg(t)
	prev := cfg.SESSDATA
	cfg.SESSDATA = ""
	t.Cleanup(func() { cfg.SESSDATA = prev })
	resp := expectStatus(t, e, "GET", "/api/watchlater/list", "", 200, "error")
	expectMessage(t, resp, "未配置 SESSDATA，无法访问 B 站稍后再看")
}

// ---------------------------------------------------------------------------
// async sync tasks
// ---------------------------------------------------------------------------

func TestFTFavoriteSync(t *testing.T) {
	stub := ftStartStub(t)
	ftCreds(t)
	e := newAPI(t, RegisterFavoriteRoutes)
	fav := database.GetFavoritesDB()
	if fav == nil {
		t.Skip("favorites database unavailable")
	}

	// A stale local folder must be pruned by the diff in doSyncFavorites.
	ftFavoriteFolderRow(t, ftFolderStale, ftFolderStale, "FT陈旧收藏夹", 0, 1, 1357000100)

	// Run one keeps the collected-folder endpoint in an error state so only the
	// created folders are written; the collected save is exercised on the second
	// run, once the created folders are already in the table.
	stub.setCode(ftPathCollectedList, -101)

	resp := expectStatus(t, e, "POST", "/api/favorite/sync", "", 200, "success").dataMap(t)
	taskID, _ := resp["task_id"].(string)
	if taskID == "" {
		t.Fatalf("sync task_id missing: %v", resp)
	}
	if resp["message"] != "收藏夹同步任务已启动，正在后台执行" {
		t.Fatalf("sync message = %v", resp["message"])
	}

	status := ftWaitTask(t, taskID)
	if status["status"] != "completed" {
		t.Fatalf("task status = %v, want completed", status)
	}
	result, ok := status["result"].(string)
	if !ok || !strings.Contains(result, fmt.Sprint(`"folders":3`)) {
		t.Fatalf("task result = %q, want the three created folders", result)
	}
	// 1 media for folder one, 2 for folder two, none for the empty folder.
	if !strings.Contains(result, `"contents":3`) {
		t.Fatalf("task result = %q, want the three created-folder medias", result)
	}

	// The stub folders landed in the cache.
	if n := ftCount(t, fav, `SELECT COUNT(*) FROM favorites_folder WHERE media_id = ?`, ftFolderCreated1); n != 1 {
		t.Fatalf("created folder count = %d, want the synced row", n)
	}
	title := ""
	if err := fav.QueryRow(`SELECT title FROM favorites_folder WHERE media_id = ?`, ftFolderCreated1).Scan(&title); err != nil {
		t.Fatalf("read synced folder: %v", err)
	}
	if title != "FT同步收藏夹" {
		t.Fatalf("synced title = %q, want the stub title", title)
	}

	// The cover-less folder with media_count 2 had its cover backfilled from the
	// first media of its resource page (features.go:620-637).
	cover := ""
	if err := fav.QueryRow(`SELECT cover FROM favorites_folder WHERE media_id = ?`, ftFolderCreated2).Scan(&cover); err != nil {
		t.Fatalf("read backfilled folder: %v", err)
	}
	if cover != "https://cdn.bilibili.com/ft/media2.jpg" {
		t.Fatalf("backfilled cover = %q, want the first media cover", cover)
	}
	// The folder that already carries a cover keeps it.
	if err := fav.QueryRow(`SELECT cover FROM favorites_folder WHERE media_id = ?`, ftFolderCreated1).Scan(&cover); err != nil {
		t.Fatalf("read synced cover: %v", err)
	}
	if cover != "https://cdn.bilibili.com/ft/cov1.jpg" {
		t.Fatalf("synced cover = %q, want the stub cover untouched", cover)
	}

	// The empty folder keeps its own empty cover and stores nothing.
	if err := fav.QueryRow(`SELECT cover FROM favorites_folder WHERE media_id = ?`, ftFolderEmpty).Scan(&cover); err != nil {
		t.Fatalf("read empty folder: %v", err)
	}
	if cover != "" {
		t.Fatalf("empty folder cover = %q, want none", cover)
	}
	if n := ftCount(t, fav, `SELECT COUNT(*) FROM favorites_content WHERE media_id = ?`, ftFolderEmpty); n != 0 {
		t.Fatalf("empty folder contents = %d, want none", n)
	}

	// The medias of each folder are stored by content id, with the bvid taken
	// from the ugc block (the second folder also carries a 已失效 media).
	for _, want := range []struct {
		folder, content int64
		bvid            string
	}{
		{ftFolderCreated1, ftContent1, ftFavVideo1},
		{ftFolderCreated2, ftContent2, ftFavVideo2},
		{ftFolderCreated2, ftContent3, ""},
	} {
		var storedBvid string
		err := fav.QueryRow(`SELECT bvid FROM favorites_content WHERE media_id = ? AND content_id = ?`,
			want.folder, want.content).Scan(&storedBvid)
		if err != nil {
			t.Fatalf("read synced content %d/%d: %v", want.folder, want.content, err)
		}
		if storedBvid != want.bvid {
			t.Fatalf("content %d/%d bvid = %q, want %q", want.folder, want.content, storedBvid, want.bvid)
		}
	}
	if n := ftCount(t, fav, `SELECT COUNT(*) FROM favorites_content WHERE media_id = ?`, ftSeasonFolder); n != 0 {
		t.Fatalf("season contents = %d, want none while the collected list fails", n)
	}

	// The stale folder is gone again.
	if n := ftCount(t, fav, `SELECT COUNT(*) FROM favorites_folder WHERE media_id = ?`, ftFolderStale); n != 0 {
		t.Fatalf("stale folder survived the prune (%d)", n)
	}

	stub.setCode(ftPathCollectedList, 0)

	// Second run takes the incremental branch: every created-folder page is
	// already known, so only the collected season contents are read.
	incr := ftRunTask(t, "同步收藏夹", doSyncFavorites)
	if incr["status"] != "completed" {
		t.Fatalf("incremental task = %v, want completed", incr)
	}
	if res, _ := incr["result"].(string); !strings.Contains(res, `"contents":2`) {
		t.Fatalf("incremental result = %q, want only the two season contents", res)
	}

	// The collected folder is stored with folder_type 1 and its season episodes
	// went through the season API.
	if n := ftCount(t, fav, `SELECT COUNT(*) FROM favorites_folder WHERE media_id = ? AND folder_type = 1`, ftSeasonFolder); n != 1 {
		t.Fatalf("collected folder rows = %d, want the type 1 marker", n)
	}
	if n := ftCount(t, fav, `SELECT COUNT(*) FROM favorites_content WHERE media_id = ?`, ftSeasonFolder); n != 2 {
		t.Fatalf("season contents = %d, want the two stub episodes", n)
	}
	// 回归 database.SaveFavoriteFolders 的类型隔离裁剪：过去它按
	// `media_id NOT IN (本次列表)` 全局删除，于是 features.go:788 那次只带
	// folder_type=1 的收藏合集保存，会把同一个 doSyncFavorites 里先写入的
	// folder_type=0 创建收藏夹全部删掉（内容行留着、收藏夹行没了）。现在裁剪
	// 只作用于本次列表覆盖的 folder_type。
	if n := ftCount(t, fav, `SELECT COUNT(*) FROM favorites_folder WHERE media_id = ? AND folder_type = 0`, ftFolderCreated1); n != 1 {
		t.Fatalf("created folder rows = %d, want the collected save to leave folder_type 0 alone", n)
	}
	if n := ftCount(t, fav, `SELECT COUNT(*) FROM favorites_content WHERE media_id = ? AND content_id = ?`, ftFolderCreated1, ftContent1); n != 1 {
		t.Fatalf("synced content rows after the collected save = %d, want them kept", n)
	}

	// Cookie expiry and generic failures are recorded on the task.
	stub.setCode(ftPathCreatedList, -6)
	expired := ftRunTask(t, "同步收藏夹", doSyncFavorites)
	if expired["status"] != "failed" {
		t.Fatalf("expired task status = %v, want failed", expired["status"])
	}
	if msg, _ := expired["error"].(string); msg != "Cookie 已过期，请重新登录" {
		t.Fatalf("expired task error = %q", msg)
	}
	stub.setCode(ftPathCreatedList, -352)
	blocked := ftRunTask(t, "同步收藏夹", doSyncFavorites)
	if msg, _ := blocked["error"].(string); !strings.Contains(msg, "获取收藏夹列表失败") {
		t.Fatalf("blocked task error = %q, want the folder list failure", msg)
	}
	stub.setCode(ftPathCreatedList, 0)

	// A folder whose resource page errors out simply stops paging (features.go:685-687).
	stub.setCode(ftPathResourceList, -412)
	partial := ftRunTask(t, "同步收藏夹", doSyncFavorites)
	if partial["status"] != "completed" {
		t.Fatalf("partial task = %v, want completed", partial["status"])
	}
	stub.setCode(ftPathResourceList, 0)
}

func TestFTLikeSync(t *testing.T) {
	stub := ftStartStub(t)
	ftCreds(t)
	e := newAPI(t, RegisterFavoriteRoutes)
	likes := database.GetLikesDB()
	if likes == nil {
		t.Skip("likes database unavailable")
	}

	resp := expectStatus(t, e, "POST", "/api/like/sync", "", 200, "success").dataMap(t)
	taskID, _ := resp["task_id"].(string)
	if taskID == "" {
		t.Fatalf("like sync task_id missing: %v", resp)
	}
	if resp["message"] != "点赞同步任务已启动，正在后台执行" {
		t.Fatalf("like sync message = %v", resp["message"])
	}
	status := ftWaitTask(t, taskID)
	if status["status"] != "completed" {
		t.Fatalf("like task = %v, want completed", status)
	}
	if res, _ := status["result"].(string); !strings.Contains(res, `"total":2`) {
		t.Fatalf("like result = %q, want the two stub videos", res)
	}
	if n := ftCount(t, likes, `SELECT COUNT(*) FROM liked_videos WHERE bvid = ?`, ftLikeA); n != 1 {
		t.Fatalf("synced like rows = %d, want %s stored", n, ftLikeA)
	}
	var owner, tname string
	if err := likes.QueryRow(`SELECT owner_name, tname FROM liked_videos WHERE bvid = ?`, ftLikeA).Scan(&owner, &tname); err != nil {
		t.Fatalf("read synced like: %v", err)
	}
	if owner != "FT点赞UP" || tname != "FT赞分区" {
		t.Fatalf("synced like metadata = %q/%q", owner, tname)
	}

	stub.setCode(ftPathLikedVideos, -6)
	expired := ftRunTask(t, "同步点赞列表", doSyncLikes)
	if msg, _ := expired["error"].(string); msg != "Cookie 已过期，请重新登录" {
		t.Fatalf("expired like task error = %q", msg)
	}
	stub.setCode(ftPathLikedVideos, -101)
	failed := ftRunTask(t, "同步点赞列表", doSyncLikes)
	if msg, _ := failed["error"].(string); !strings.Contains(msg, "获取点赞列表失败") {
		t.Fatalf("failed like task error = %q", msg)
	}
	stub.setCode(ftPathLikedVideos, 0)

	// A DedeUserID that is not numeric fails before any request is made
	// (features.go:900-905).
	before := stub.hitsFor(ftPathLikedVideos)
	bad := ftRunTask(t, "同步点赞列表", func(_ *config.Config, taskID string) {
		doSyncLikes(&config.Config{SESSDATA: "ft-sessdata-2013", DedeUserID: "not-a-uid"}, taskID)
	})
	if bad["status"] != "failed" {
		t.Fatalf("bad uid task = %v, want failed", bad["status"])
	}
	if msg, _ := bad["error"].(string); msg != "DedeUserID 无效" {
		t.Fatalf("bad uid error = %q, want the uid validation message", msg)
	}
	if stub.hitsFor(ftPathLikedVideos) != before {
		t.Fatal("an invalid DedeUserID must not reach the likes API")
	}

	// Without DedeUserID the endpoint refuses the request outright (the guard at
	// features.go:876-878 runs before any task is started, so no task id comes
	// back and the error envelope carries no data).
	cfg := loadCfg(t)
	prev := cfg.DedeUserID
	cfg.DedeUserID = ""
	t.Cleanup(func() { cfg.DedeUserID = prev })
	noUID := expectStatus(t, e, "POST", "/api/like/sync", "", 200, "error")
	expectMessage(t, noUID, "未配置 DedeUserID")
}

func TestFTWatchLaterSync(t *testing.T) {
	stub := ftStartStub(t)
	ftCreds(t)
	e := newAPI(t, RegisterFavoriteRoutes)
	wl := database.GetWatchLaterDB()
	if wl == nil {
		t.Skip("watchlater database unavailable")
	}

	// A local row the remote list does not contain is pruned.
	ftWatchLaterRow(t, ftWLStale, ftWLUntracked, 1357000400, "FT陈旧稍后再看")

	resp := expectStatus(t, e, "POST", "/api/watchlater/sync", "", 200, "success").dataMap(t)
	taskID, _ := resp["task_id"].(string)
	if resp["message"] != "稍后再看同步任务已启动，正在后台执行" {
		t.Fatalf("sync message = %v", resp["message"])
	}
	status := ftWaitTask(t, taskID)
	if status["status"] != "completed" {
		t.Fatalf("watchlater task = %v, want completed", status)
	}
	if res, _ := status["result"].(string); !strings.Contains(res, `"total":2`) {
		t.Fatalf("watchlater result = %q, want the two stub items", res)
	}
	if n := ftCount(t, wl, `SELECT COUNT(*) FROM watchlater_videos WHERE bvid = ?`, ftWLB); n != 1 {
		t.Fatalf("synced watchlater rows = %d, want %s stored", n, ftWLB)
	}
	if n := ftCount(t, wl, `SELECT COUNT(*) FROM watchlater_videos WHERE bvid = ?`, ftWLStale); n != 0 {
		t.Fatalf("stale watchlater row survived the sync (%d)", n)
	}

	stub.setCode(ftPathWatchLater, -6)
	expired := ftRunTask(t, "同步稍后再看", doSyncWatchLater)
	if msg, _ := expired["error"].(string); msg != "Cookie 已过期，请重新登录" {
		t.Fatalf("expired watchlater error = %q", msg)
	}
	stub.setCode(ftPathWatchLater, -352)
	blocked := ftRunTask(t, "同步稍后再看", doSyncWatchLater)
	if msg, _ := blocked["error"].(string); !strings.Contains(msg, "同步稍后再看失败") {
		t.Fatalf("blocked watchlater error = %q", msg)
	}
	stub.setCode(ftPathWatchLater, 0)
}

// ---------------------------------------------------------------------------
// dynamics
// ---------------------------------------------------------------------------

func TestFTDynamicLegacyAndReads(t *testing.T) {
	e := newAPI(t, RegisterFavoriteRoutes)

	// getDynamicListLegacy answers with the stub payload.
	w := doRaw(t, e, "GET", "/api/dynamic/list", "")
	payload := ftPlain(t, w)
	if payload["status"] != "success" {
		t.Fatalf("legacy list = %v, want a success envelope", payload)
	}
	data := payload["data"].(map[string]interface{})
	if ftNum(data["total"]) != "0" {
		t.Fatalf("legacy total = %v, want 0", data["total"])
	}
	if len(dataArray(t, data["records"])) != 0 {
		t.Fatalf("legacy records = %v, want an empty list", data["records"])
	}
	if msg, _ := data["message"].(string); !strings.Contains(msg, "/dynamic/db/hosts") {
		t.Fatalf("legacy hint = %q, want the replacement endpoint", msg)
	}

	// syncDynamicLegacy returns a bare {status, message} object.
	synced := expectStatus(t, e, "POST", "/api/dynamic/sync", "", 200, "success")
	expectMessage(t, synced, "请使用 /dynamic/space/auto/{host_mid} 端点")

	// getDynamicDbHosts reports the cached hosts.
	ftDynItem(t, ftDynID1, ftDynHost, "DYNAMIC_TYPE_AV", ftDynVideo, "FT动态视频", 1357000100, "[]")
	ftDynItem(t, ftDynID2, ftDynHost, "DYNAMIC_TYPE_DRAW", "", "FT图文", 1357000101, `["https://i0.hdslb.com/ft/remote-1.jpg"]`)
	ftSeedDynHost(t, ftDynHost, ftUpName, 2, 2, 1357000101)

	payload = ftSuccess(t, e, "GET", "/api/dynamic/db/hosts?limit=100&offset=0", "")
	list := dataArray(t, payload["data"])
	if !ftHasEntry(list, "host_mid", ftDynHost) {
		t.Fatalf("hosts = %v, want the seeded host", list)
	}
	entry := ftListEntry(t, list, "host_mid", ftDynHost)
	if ftNum(entry["item_count"]) != "2" || ftNum(entry["core_count"]) != "2" {
		t.Fatalf("host counters = %v, want the seeded 2/2", entry)
	}

	// getDynamicDbSpace pages through one host's items newest first.
	space := ftSuccess(t, e, "GET", fmt.Sprintf("/api/dynamic/db/space/%s?limit=1&offset=0", ftDynHost), "")
	if ftNum(space["total"]) != "2" {
		t.Fatalf("space total = %v, want both seeded items", space["total"])
	}
	items := dataArray(t, space["items"])
	if len(items) != 1 {
		t.Fatalf("space items = %v, want the requested page size", items)
	}
	if items[0].(map[string]interface{})["id_str"] != ftDynID2 {
		t.Fatalf("first item = %v, want the newest publish_ts", items[0])
	}
	pageTwo := ftSuccess(t, e, "GET", fmt.Sprintf("/api/dynamic/db/space/%s?limit=1&offset=1", ftDynHost), "")
	if dataArray(t, pageTwo["items"])[0].(map[string]interface{})["id_str"] != ftDynID1 {
		t.Fatalf("offset window = %v, want the older item", pageTwo["items"])
	}
	// A host with no rows yields an empty slice, not null (features.go:1327).
	empty := ftSuccess(t, e, "GET", "/api/dynamic/db/space/2013777999", "")
	if ftNum(empty["total"]) != "0" || len(dataArray(t, empty["items"])) != 0 {
		t.Fatalf("unknown host payload = %v, want total 0 and an empty item list", empty)
	}
}

func TestFTDynamicUserCard(t *testing.T) {
	stub := ftStartStub(t)
	e := newAPI(t, RegisterFavoriteRoutes)

	// Without credentials the handler reports the configuration gap.
	ftNoCreds(t)
	ftFailure(t, e, "GET", "/api/dynamic/user_card/2013900002", "", "SESSDATA 未配置")

	ftCreds(t)
	payload := ftSuccess(t, e, "GET", "/api/dynamic/user_card/2013900002", "")
	card := payload["data"].(map[string]interface{})
	if card["name"] != ftUpName {
		t.Fatalf("user card = %v, want the stub profile", card)
	}
	if card["mid"] != "2013900002" {
		t.Fatalf("user card mid = %v, want the stub mid", card["mid"])
	}
	if got := stub.lastQuery(ftPathUserCard).Get("mid"); got != "2013900002" {
		t.Fatalf("card request mid = %q", got)
	}

	// Any API error is surfaced verbatim; -101 means the user does not exist.
	stub.setCode(ftPathUserCard, -101)
	failed := ftFailure(t, e, "GET", "/api/dynamic/user_card/2013900002", "", "code=-101")
	if failed["success"] != false {
		t.Fatalf("failed card payload = %v", failed)
	}
	stub.setCode(ftPathUserCard, -352)
	ftFailure(t, e, "GET", "/api/dynamic/user_card/2013900002", "", "code=-352")
	stub.setCode(ftPathUserCard, 0)
}

func TestFTDynamicAutoFetchAndProgress(t *testing.T) {
	stub := ftStartStub(t)
	ftCreds(t)
	e := newAPI(t, RegisterFavoriteRoutes)
	ftDrainProgress(t)

	// Hold the feed so the second start request observes the running flag.
	release := make(chan struct{})
	stub.blockDynamic(release)
	t.Cleanup(func() {
		select {
		case <-release:
		default:
			close(release)
		}
	})

	started := ftSuccess(t, e, "GET", fmt.Sprintf("/api/dynamic/space/auto/%s?save_media=false&need_top=true", ftDynHostBusy), "")
	if started["message"] != "开始抓取动态" {
		t.Fatalf("start payload = %v, want the acknowledgement", started)
	}
	// startDynamicAutoFetch spawns services.FetchDynamicSpace, which raises the
	// IsRunning flag inside that goroutine, so neither the flag nor the first
	// feed request is synchronous with the handler reply. Waiting for the
	// blocked page to arrive makes the guard (features.go:1355-1359) deterministic
	// instead of racing it.
	if !ftWaitFlag(func() bool { return stub.hitsFor(ftPathDynamicSpace) == 1 }) {
		t.Fatalf("the fetch goroutine never issued the blocked feed request (hits = %d)", stub.hitsFor(ftPathDynamicSpace))
	}
	if !services.GetDynamicFetchStatus(ftDynHostBusy).IsRunning {
		t.Fatal("the running flag was not raised before the first feed request")
	}
	running := ftFailure(t, e, "GET", fmt.Sprintf("/api/dynamic/space/auto/%s?save_media=false", ftDynHostBusy), "", "该 UP 的动态抓取正在进行中")
	if running["success"] != false {
		t.Fatalf("guard payload = %v", running)
	}
	// The duplicate is rejected by the running flag, so it never reaches B 站:
	// the blocked page above is still the only feed request.
	if got := stub.hitsFor(ftPathDynamicSpace); got != 1 {
		t.Fatalf("feed requests = %d, want the guard to keep the duplicate off the network", got)
	}
	close(release)
	ftWaitFetchDone(t, ftDynHostBusy)

	// The fetched archive is stored under the requested host.
	space := ftSuccess(t, e, "GET", fmt.Sprintf("/api/dynamic/db/space/%s", ftDynHostBusy), "")
	if ftNum(space["total"]) != "1" {
		t.Fatalf("fetched total = %v, want the single archive dynamic", space["total"])
	}
	item := dataArray(t, space["items"])[0].(map[string]interface{})
	if item["id_str"] != ftDynID1 || item["bvid"] != ftDynVideo {
		t.Fatalf("fetched item = %v, want the stub archive", item)
	}
	// The archive branch of the parser (services/dynamic.go:300-304) copies the
	// major block, so the stored row must carry the title and the cover src.
	if item["title"] != "FT动态视频" || item["cover"] != "https://cdn.bilibili.com/ft/dyn1.jpg" {
		t.Fatalf("fetched archive = title %v cover %v, want the stub values", item["title"], item["cover"])
	}
	if item["txt"] != "FT动态正文" {
		t.Fatalf("fetched desc text = %v, want the module desc", item["txt"])
	}

	// The second host's fetch stores a DRAW dynamic; clean it up so the fixture
	// does not leak into the shared database.
	t.Cleanup(func() {
		if db := database.GetDynamicDB(); db != nil {
			ftExecDB(t, db, `DELETE FROM dynamics WHERE host_mid = ?`, ftDynHostStopped)
		}
	})
	// An offset page jumps straight to the stop branch (features.go:1340-1342).
	ftSuccess(t, e, "GET", fmt.Sprintf("/api/dynamic/space/auto/%s?save_media=false&need_top=false", ftDynHostStopped), "")
	ftWaitFetchDone(t, ftDynHostStopped)

	// stopDynamicAutoFetch always answers with the stop-signal envelope.
	stopped := ftSuccess(t, e, "POST", fmt.Sprintf("/api/dynamic/space/auto/%s/stop", ftDynHostStopped), "")
	if stopped["message"] != "已发送停止信号" {
		t.Fatalf("stop payload = %v", stopped)
	}

	// dynamicProgressSSE streams whatever the service published and returns on
	// the terminal message. The buffered channel is drained first so the reply
	// is deterministic.
	ftDrainProgress(t)
	services.GetDynamicProgressChan() <- "FT测试进度：全部抓取完毕"
	w := doRaw(t, e, "GET", fmt.Sprintf("/api/dynamic/space/auto/%s/progress", ftDynHost), "")
	if w.Code != http.StatusOK {
		t.Fatalf("sse status = %d, want 200 (body %q)", w.Code, w.Body.String())
	}
	body := w.Body.String()
	if !strings.Contains(body, "Content-Type") && !strings.Contains(body, "event: progress") {
		t.Fatalf("sse body = %q, want an event: progress frame", body)
	}
	if !strings.Contains(body, "event: progress") {
		t.Fatalf("sse body = %q, want an event: progress frame", body)
	}
	if !strings.Contains(body, "全部抓取完毕") {
		t.Fatalf("sse body = %q, want the terminal message", body)
	}
	if !strings.HasPrefix(body, "event: progress") {
		t.Fatalf("sse body = %q, want the frame written before any body", body)
	}

	// deleteDynamicSpace drops both the items and the host record.
	ftSuccess(t, e, "DELETE", fmt.Sprintf("/api/dynamic/space/%s", ftDynHostBusy), "")
	after := ftSuccess(t, e, "GET", fmt.Sprintf("/api/dynamic/db/space/%s", ftDynHostBusy), "")
	if ftNum(after["total"]) != "0" {
		t.Fatalf("space after delete = %v, want it emptied", after["total"])
	}
}

func TestFTDynamicItemDelete(t *testing.T) {
	e := newAPI(t, RegisterFavoriteRoutes)
	ftDynItem(t, ftDynID3, ftDynHostDeleted, "DYNAMIC_TYPE_AV", ftDynVideo, "FT待删动态", 1357000102, `["https://i0.hdslb.com/ft/remote-1.jpg"]`)
	ftSeedDynHost(t, ftDynHostDeleted, ftUpName, 1, 1, 1357000102)

	// The unknown id fails the lookup before anything is removed.
	ftFailure(t, e, "DELETE", fmt.Sprintf("/api/dynamic/item/%s", ftDynIDAbsent), "", "动态不存在")

	deleted := ftSuccess(t, e, "DELETE", fmt.Sprintf("/api/dynamic/item/%s", ftDynID3), "")
	if deleted["message"] != "已删除" {
		t.Fatalf("delete payload = %v, want 已删除", deleted)
	}
	dynDB := database.GetDynamicDB()
	if dynDB == nil {
		t.Skip("dynamic database unavailable")
	}
	if n := ftCount(t, dynDB, `SELECT COUNT(*) FROM dynamics WHERE id_str = ?`, ftDynID3); n != 0 {
		t.Fatalf("dynamic row survived the delete (%d)", n)
	}
	// UpdateDynamicHostStats recomputed the counters to zero.
	var itemCount int
	if err := dynDB.QueryRow(`SELECT item_count FROM dynamic_hosts WHERE host_mid = ?`, ftDynHostDeleted).Scan(&itemCount); err != nil {
		t.Fatalf("read host stats: %v", err)
	}
	if itemCount != 0 {
		t.Fatalf("host item_count = %d, want the recomputed zero", itemCount)
	}
}

// ftDrainProgress empties the shared buffered progress channel.
func ftDrainProgress(t *testing.T) {
	t.Helper()
	ch := services.GetDynamicProgressChan()
	for {
		select {
		case <-ch:
		default:
			return
		}
	}
}

// ftWaitFlag polls cond for up to 10s so a goroutine-owned flag becomes a
// deterministic precondition instead of a race.
func ftWaitFlag(cond func() bool) bool {
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		if cond() {
			return true
		}
		time.Sleep(5 * time.Millisecond)
	}
	return false
}

// ftWaitFetchDone blocks until the global dynamic fetch flag clears.
func ftWaitFetchDone(t *testing.T, hostMid string) {
	t.Helper()
	deadline := time.Now().Add(15 * time.Second)
	for time.Now().Before(deadline) {
		if !services.GetDynamicFetchStatus(hostMid).IsRunning {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("dynamic fetch for %s never finished", hostMid)
}
