package biliapi

import (
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"testing"
)

// wbi64Keys is a 64-character alphabet (a-z, 0-9, A-Z, "_", "-") used to observe
// the mixin-key permutation: every index in mixinKeyEncTab is < 64, so the whole
// table is applied before truncation to 32 characters.
const wbi64Keys = "abcdefghijklmnopqrstuvwxyz0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ_-"

func TestGetMixinKey(t *testing.T) {
	got := getMixinKey(wbi64Keys)
	want := "KLscRix6pOk5WdJ91HfN7jGt32oDmCFn"
	if len(got) != 32 {
		t.Fatalf("mixin key length = %d, want 32 (%q)", len(got), got)
	}
	if got != want {
		t.Fatalf("mixin key = %q, want %q", got, want)
	}

	// A realistic key pair (32 + 32 characters, all indices below 64 are kept
	// only when the index is < len(orig)); the result must still be 32 chars.
	imgKey := "e8f8eb9f4f42f42d4e4e4f8b0f0f0f0f"
	subKey := "0f0f0f0fe8f8eb9f4f42f42d4e4e4f42"
	got = getMixinKey(imgKey + subKey)
	if len(got) != 32 {
		t.Fatalf("mixin key length for key pair = %d, want 32 (%q)", len(got), got)
	}
	// The mixin key is a permutation of the source characters, so it must be a
	// subset of the concatenated keys and contain no duplicates beyond those in
	// the source.
	src := imgKey + subKey
	for _, r := range got {
		if !strings.ContainsRune(src, r) {
			t.Fatalf("mixin key char %q not present in source keys", r)
		}
	}

	// Only table indices smaller than len(orig) contribute characters, so a
	// 64-char source yields exactly the first 32 permutations of the table.
	// Verifying the first few positions pins down the permutation order.
	first := getMixinKey(wbi64Keys)
	for i, idx := range []int{46, 47, 18, 2, 53, 8} {
		if first[i] != wbi64Keys[idx] {
			t.Fatalf("mixin key[%d] = %q, want %q (table index %d)", i, first[i], wbi64Keys[idx], idx)
		}
	}
}

// wbiExpected reproduces the canonical WBI query string: params sorted by key,
// values stripped of "'()!", url-escaped and joined with "&", then md5'd with the
// mixin key appended.
func wbiExpected(t *testing.T, ordered map[string]string, mixinKey string) string {
	t.Helper()
	parts := make([]string, 0, len(ordered))
	for _, k := range sortedKeys(ordered) {
		v := ordered[k]
		v = strings.ReplaceAll(v, "'", "")
		v = strings.ReplaceAll(v, "(", "")
		v = strings.ReplaceAll(v, ")", "")
		v = strings.ReplaceAll(v, "!", "")
		parts = append(parts, url.QueryEscape(k)+"="+url.QueryEscape(v))
	}
	sum := md5.Sum([]byte(strings.Join(parts, "&") + mixinKey))
	return fmt.Sprintf("%x", sum)
}

func sortedKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func TestWbiSign(t *testing.T) {
	imgKey := "e8f8eb9f4f42f42d4e4e4f8b0f0f0f0f"
	subKey := "0f0f0f0fe8f8eb9f4f42f42d4e4e4f42"
	mixin := getMixinKey(imgKey + subKey)

	params := map[string]string{"foo": "bar", "aaa": "b c", "Zeta": "1"}
	sig := wbiSign(params, imgKey, subKey)

	if len(sig) != 32 {
		t.Fatalf("signature length = %d, want 32 hex chars (%q)", len(sig), sig)
	}
	if !isHex(sig) {
		t.Fatalf("signature %q is not a hex md5 string", sig)
	}

	// wbiSign injects wts (current unix seconds) into the map it signs.
	wts, ok := params["wts"]
	if !ok {
		t.Fatalf("wbiSign did not set params[\"wts\"]")
	}
	if len(wts) < 10 {
		t.Fatalf("wts = %q does not look like a unix timestamp", wts)
	}

	want := wbiExpected(t, params, mixin)
	if sig != want {
		t.Fatalf("signature = %q, want %q", sig, want)
	}

	// Signature must be stable for identical inputs (same map contents and same
	// wts): rebuild two maps with the observed wts pinned and compare.
	a := map[string]string{"foo": "bar", "aaa": "b c", "Zeta": "1", "wts": wts}
	b := map[string]string{"Zeta": "1", "wts": wts, "aaa": "b c", "foo": "bar"}
	if sa, sb := wbiSign(a, imgKey, subKey), wbiSign(b, imgKey, subKey); sa != sb || sa != want {
		t.Fatalf("signature not stable: %q vs %q (want %q)", sa, sb, want)
	}

	// A different mixin key must produce a different signature.
	if other := wbiSign(map[string]string{"foo": "bar", "wts": wts}, "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"); other == want {
		t.Fatalf("signature unchanged with different wbi keys")
	}
}

func TestWbiSignFiltersSpecialChars(t *testing.T) {
	imgKey := "e8f8eb9f4f42f42d4e4e4f8b0f0f0f0f"
	subKey := "0f0f0f0fe8f8eb9f4f42f42d4e4e4f42"
	mixin := getMixinKey(imgKey + subKey)

	// ' ( ) ! are stripped from values before escaping, so a value of
	// "a'b(c)!" is signed as "abc" (url-escaped) while the original value is
	// preserved in the caller's map.
	params := map[string]string{"k": "a'b(c)!"}
	sig := wbiSign(params, imgKey, subKey)
	if params["k"] != "a'b(c)!" {
		t.Fatalf("wbiSign mutated input value: %q", params["k"])
	}
	want := wbiExpected(t, params, mixin)
	if sig != want {
		t.Fatalf("signature = %q, want %q", sig, want)
	}

	// Sanity check that the filter really applied: signing the pre-stripped
	// value gives the same digest.
	stripped := map[string]string{"k": "abc", "wts": params["wts"]}
	if got := wbiSign(stripped, imgKey, subKey); got != sig {
		t.Fatalf("special-char filtering not applied: %q vs %q", got, sig)
	}

	// Values are percent-escaped in the signed string: "+" survives the escape,
	// and space becomes "+".
	p2 := map[string]string{"q": "a+b c"}
	s2 := wbiSign(p2, imgKey, subKey)
	if exp := wbiExpected(t, p2, mixin); s2 != exp {
		t.Fatalf("escaped signature = %q, want %q", s2, exp)
	}
}

func isHex(s string) bool {
	for _, r := range s {
		if !((r >= '0' && r <= '9') || (r >= 'a' && r <= 'f')) {
			return false
		}
	}
	return true
}

var buvid3Re = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

func TestGenerateBuvid3(t *testing.T) {
	got := generateBuvid3()
	if !buvid3Re.MatchString(got) {
		t.Fatalf("buvid3 %q does not match xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx", got)
	}
	if second := generateBuvid3(); second == got {
		t.Fatalf("buvid3 not random: two calls returned %q", got)
	}
}

func TestGenerateDmImgList(t *testing.T) {
	// NOTE: the `if x < 0 / y < 0 / timestamp < 0` clamps in generateDmImgList
	// (client.go:92-101) are dead code - the random bases 1245, 1285 and 30 with a
	// +/-5 jitter can never go negative - so they stay uncovered by design.
	raw := generateDmImgList()
	var parsed []struct {
		X         int `json:"x"`
		Y         int `json:"y"`
		Z         int `json:"z"`
		Timestamp int `json:"timestamp"`
		Type      int `json:"type"`
	}
	if err := json.Unmarshal([]byte(raw), &parsed); err != nil {
		t.Fatalf("dm_img_list %q is not valid JSON: %v", raw, err)
	}
	if len(parsed) != 1 {
		t.Fatalf("dm_img_list has %d entries, want 1: %q", len(parsed), raw)
	}
	e := parsed[0]
	if e.Z != 0 {
		t.Fatalf("dm_img_list z = %d, want 0", e.Z)
	}
	if e.Type != 0 {
		t.Fatalf("dm_img_list type = %d, want 0", e.Type)
	}
	// timestamp is 30 +/- 5 -> [25, 35]
	if e.Timestamp < 25 || e.Timestamp > 35 {
		t.Fatalf("dm_img_list timestamp = %d, want within [25,35]", e.Timestamp)
	}
	// x = 3*(1245+[-5,5]) + 2*(1285+[-5,5]) and y = 4*xx - 5*yy
	if e.X < 3*1240+2*1280 || e.X > 3*1250+2*1290 {
		t.Fatalf("dm_img_list x = %d out of expected range", e.X)
	}
	if e.Y < 4*1240-5*1290 || e.Y > 4*1250-5*1280 {
		t.Fatalf("dm_img_list y = %d out of expected range", e.Y)
	}
	if !strings.HasPrefix(raw, `[{"x":`) {
		t.Fatalf("unexpected dm_img_list shape: %q", raw)
	}
}

func TestAddDmVerifyInfo(t *testing.T) {
	got := addDmVerifyInfo("a=1&b=2")
	if !strings.HasPrefix(got, "a=1&b=2&dm_img_list=") {
		t.Fatalf("addDmVerifyInfo did not append dm params to input: %q", got)
	}
	if !strings.HasSuffix(got, "&dm_img_str=bm8gd2ViZ2w&dm_cover_img_str=bm8gd2ViZ2w") {
		t.Fatalf("unexpected dm_img_str / dm_cover_img_str suffix: %q", got)
	}

	parsed, err := url.ParseQuery(got[len("a=1&b=2&"):])
	if err != nil {
		t.Fatalf("result is not a parseable query string: %v", err)
	}
	list := parsed.Get("dm_img_list")
	if list == "" {
		t.Fatalf("missing dm_img_list in %q", got)
	}
	if !strings.HasPrefix(list, "[{") {
		t.Fatalf("dm_img_list was not url-escaped JSON: %q", list)
	}
	if got := parsed.Get("dm_img_str"); got != "bm8gd2ViZ2w" {
		t.Fatalf("dm_img_str = %q, want base64 of \"no webgl\"", got)
	}
	// Empty base params are still handled (leading "&" is kept by the impl).
	empty := addDmVerifyInfo("")
	if !strings.HasPrefix(empty, "&dm_img_list=") {
		t.Fatalf("addDmVerifyInfo(\"\") = %q, want leading &dm_img_list", empty)
	}
}

func TestNewClient(t *testing.T) {
	c := NewClient("my-sessdata")
	if c.SESSDATA != "my-sessdata" {
		t.Fatalf("SESSDATA = %q", c.SESSDATA)
	}
	if !strings.Contains(c.UserAgent, "Mozilla/5.0") {
		t.Fatalf("UserAgent = %q", c.UserAgent)
	}
	if !buvid3Re.MatchString(c.Buvid3) {
		t.Fatalf("Buvid3 = %q is not a buvid3 UUID", c.Buvid3)
	}
	if c.client == nil || c.client.Timeout.Seconds() != 30 {
		t.Fatalf("http client timeout = %v, want 30s", c.client)
	}
	if c.BiliJct != "" || c.DedeUserID != "" || c.ImgKey != "" || c.SubKey != "" {
		t.Fatalf("NewClient should leave csrf/user/wbi keys empty")
	}

	// An empty SESSDATA keeps the client in the "not logged in" state that the
	// header builder and the write endpoints guard on.
	anon := NewClient("")
	if anon.SESSDATA != "" {
		t.Fatalf("anonymous client SESSDATA = %q", anon.SESSDATA)
	}
	if _, ok := anon.getHeaders()["Cookie"]; ok {
		t.Fatalf("anonymous client must not send a Cookie header")
	}
}

func TestNewClientWithConfig(t *testing.T) {
	c := NewClientWithConfig("sd", "jct", "12345")
	if c.SESSDATA != "sd" || c.BiliJct != "jct" || c.DedeUserID != "12345" {
		t.Fatalf("unexpected credentials: %+v", c)
	}
	if c.UserAgent == "" || !buvid3Re.MatchString(c.Buvid3) {
		t.Fatalf("NewClientWithConfig must reuse NewClient defaults")
	}
}

func TestGetHeaders(t *testing.T) {
	c := NewClient("sess-data-value")
	h := c.getHeaders()

	for k, want := range map[string]string{
		"User-Agent":         c.UserAgent,
		"Referer":            "https://www.bilibili.com",
		"Origin":             "https://www.bilibili.com",
		"Accept":             "application/json, text/plain, */*",
		"Accept-Language":    "zh-CN,zh;q=0.9,en;q=0.8",
		"Accept-Encoding":    "gzip, deflate",
		"Connection":         "keep-alive",
		"Sec-Ch-Ua-Mobile":   "?0",
		"Sec-Fetch-Dest":     "empty",
		"Sec-Fetch-Mode":     "cors",
		"Sec-Fetch-Site":     "same-site",
		"Sec-Ch-Ua-Platform": `"Windows"`,
	} {
		if got := h[k]; got != want {
			t.Fatalf("header %s = %q, want %q", k, got, want)
		}
	}

	// Logged-in state: cookie jar contents are flattened into one Cookie header.
	cookie := h["Cookie"]
	if cookie == "" {
		t.Fatalf("expected Cookie header for logged-in client")
	}
	if !strings.Contains(cookie, "SESSDATA=sess-data-value") {
		t.Fatalf("Cookie missing SESSDATA: %q", cookie)
	}
	if !strings.Contains(cookie, "buvid3="+c.Buvid3) {
		t.Fatalf("Cookie missing buvid3: %q", cookie)
	}
	if !regexp.MustCompile(`buvid4=[0-9a-f-]{36}`).MatchString(cookie) {
		t.Fatalf("Cookie missing generated buvid4: %q", cookie)
	}
	if !regexp.MustCompile(`b_lsid=[0-9a-f]{8}_[0-9a-f]{10}`).MatchString(cookie) {
		t.Fatalf("Cookie missing b_lsid in 8hex_10hex form: %q", cookie)
	}
	if !strings.Contains(cookie, "b_nut=1234567890") ||
		!strings.Contains(cookie, "bili_ticket=") ||
		!strings.Contains(cookie, "bili_ticket_mid=") {
		t.Fatalf("Cookie missing static fields: %q", cookie)
	}
	// bili_jct / DedeUserID are only appended when configured.
	if strings.Contains(cookie, "bili_jct=") || strings.Contains(cookie, "DedeUserID=") {
		t.Fatalf("Cookie should not contain csrf/user id when unset: %q", cookie)
	}

	full := NewClientWithConfig("sd", "csrf-token", "987")
	h2 := full.getHeaders()
	cookie2 := h2["Cookie"]
	if !strings.Contains(cookie2, "SESSDATA=sd") ||
		!strings.Contains(cookie2, "bili_jct=csrf-token") ||
		!strings.Contains(cookie2, "DedeUserID=987") {
		t.Fatalf("full client Cookie = %q", cookie2)
	}
	if !strings.Contains(cookie2, "bili_ticket=;") && !strings.HasSuffix(cookie2, "bili_ticket_mid=") {
		// bili_ticket / bili_ticket_mid are always present (empty valued).
		t.Fatalf("full client Cookie lost empty ticket fields: %q", cookie2)
	}
}

func TestBiliResponseUnmarshal(t *testing.T) {
	t.Run("success envelope", func(t *testing.T) {
		var r BiliResponse
		body := `{"code":0,"message":"0","ttl":1,"data":{"max":123,"list":[]}}`
		if err := json.Unmarshal([]byte(body), &r); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if r.Code != 0 || r.Message != "0" {
			t.Fatalf("got %+v", r)
		}
		var data struct {
			Max int64 `json:"max"`
		}
		if err := json.Unmarshal(r.Data, &data); err != nil {
			t.Fatalf("data round trip failed: %v", err)
		}
		if data.Max != 123 {
			t.Fatalf("data.max = %d", data.Max)
		}
	})

	t.Run("non-zero code with message", func(t *testing.T) {
		var r BiliResponse
		if err := json.Unmarshal([]byte(`{"code":-101,"message":"账号未登录","data":null}`), &r); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if r.Code != -101 || r.Message != "账号未登录" {
			t.Fatalf("got %+v", r)
		}
		if string(r.Data) != "null" {
			t.Fatalf("Data = %s, want raw null", r.Data)
		}
	})

	t.Run("risk control code -352", func(t *testing.T) {
		var r BiliResponse
		if err := json.Unmarshal([]byte(`{"code":-352,"message":"-352","data":{}}`), &r); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if r.Code != -352 {
			t.Fatalf("code = %d, want -352 (risk control)", r.Code)
		}
	})

	t.Run("http 200 malformed body", func(t *testing.T) {
		var r BiliResponse
		if err := json.Unmarshal([]byte(`<html>500</html>`), &r); err == nil {
			t.Fatalf("expected json error for non-JSON body")
		}
	})

	t.Run("data shape mismatch", func(t *testing.T) {
		var r BiliResponse
		if err := json.Unmarshal([]byte(`{"code":0,"message":"0","data":"oops"}`), &r); err != nil {
			t.Fatalf("RawMessage must accept any shape: %v", err)
		}
		var data HistoryCursorData
		if err := json.Unmarshal(r.Data, &data); err == nil {
			t.Fatalf("expected decode of string data into struct to fail")
		}
	})

	t.Run("missing fields default to zero values", func(t *testing.T) {
		var r BiliResponse
		if err := json.Unmarshal([]byte(`{}`), &r); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if r.Code != 0 || r.Message != "" || r.Data != nil {
			t.Fatalf("got %+v", r)
		}
	})
}

func TestApiError(t *testing.T) {
	err := &ApiError{Code: -101, Message: "unsafe protected"}
	if got, want := err.Error(), "api error: code=-101, message=unsafe protected"; got != want {
		t.Fatalf("Error() = %q, want %q", got, want)
	}
	if got, want := (&ApiError{Code: 0, Message: ""}).Error(), "api error: code=0, message="; got != want {
		t.Fatalf("Error() = %q, want %q", got, want)
	}
	// Errors surfaced through the error interface stay reachable as *ApiError.
	var target *ApiError
	var asError error = &ApiError{Code: -352, Message: "risk control"}
	if !errors.As(asError, &target) || target.Code != -352 {
		t.Fatalf("errors.As could not recover *ApiError from %v", asError)
	}
}

func TestModelJSONTags(t *testing.T) {
	// History entry / cursor decoding from a recorded fixture shape.
	var data HistoryCursorData
	fixture := `{
		"cursor":{"max":100,"view_at":1700000000,"business":"archive","ps":8},
		"list":[{
			"title":"t","long_title":"lt","cover":"c","uri":"u",
			"history":{"bvid":"BV1xx","page":1,"cid":42,"part":"p","business":"archive","dt":5},
			"view_at":1700000000,"progress":10,"badge":"b","show_title":"st","icon":"i",
			"business":"archive","bvid":"BV1xx","duration":200,"author_name":"an",
			"author_face":"af","author_mid":777
		}]
	}`
	if err := json.Unmarshal([]byte(fixture), &data); err != nil {
		t.Fatalf("fixture decode failed: %v", err)
	}
	if data.Cursor.Max != 100 || data.Cursor.ViewAt != 1700000000 || data.Cursor.Business != "archive" || data.Cursor.Ps != 8 {
		t.Fatalf("cursor = %+v", data.Cursor)
	}
	if len(data.List) != 1 {
		t.Fatalf("list len = %d", len(data.List))
	}
	e := data.List[0]
	if e.Title != "t" || e.LongTitle != "lt" || e.Bvid != "BV1xx" || e.DTotal != 200 || e.AuthorMid != 777 {
		t.Fatalf("entry = %+v", e)
	}
	if e.History.Cid != 42 || e.History.Bvid != "BV1xx" || e.History.Dt != 5 || e.History.Part != "p" {
		t.Fatalf("history = %+v", e.History)
	}

	// Video info, watch later, liked video and favorite models.
	var vi VideoInfo
	if err := json.Unmarshal([]byte(`{"bvid":"BV1","aid":2,"videos":3,"tid":4,"tname":"tn","copyright":1,"pic":"p","title":"T","pubdate":10,"ctime":11,"desc":"d","duration":12,"owner":{"mid":13,"name":"o","face":"f"},"stat":{"view":1,"danmaku":2,"reply":3,"favorite":4,"coin":5,"share":6,"like":7}}`), &vi); err != nil {
		t.Fatalf("video info decode failed: %v", err)
	}
	if vi.Bvid != "BV1" || vi.Aid != 2 || vi.Owner.Mid != 13 || vi.Stat.Like != 7 || vi.Duration != 12 {
		t.Fatalf("video info = %+v", vi)
	}

	var wl WatchLaterData
	if err := json.Unmarshal([]byte(`{"count":1,"list":[{"aid":5,"bvid":"BV5","title":"w","pic":"p","desc":"d","duration":9,"tid":1,"tname":"tn","owner":{"mid":2,"name":"n","face":"f"},"stat":{"view":3},"add_at":100,"pubdate":101}]}`), &wl); err != nil {
		t.Fatalf("watch later decode failed: %v", err)
	}
	if wl.Count != 1 || wl.List[0].Aid != 5 || wl.List[0].AddAt != 100 || wl.List[0].Owner.Name != "n" {
		t.Fatalf("watch later = %+v", wl)
	}

	var lv LikedVideoData
	if err := json.Unmarshal([]byte(`{"list":[{"aid":6,"title":"l","pic":"p","desc":"d","duration":8,"tid":1,"tname":"t2","owner":{"mid":3,"name":"o","face":"f"},"stat":{"like":4},"pubdate":50,"bvid":"BV6"}]}`), &lv); err != nil {
		t.Fatalf("liked video decode failed: %v", err)
	}
	if lv.List[0].Bvid != "BV6" || lv.List[0].Tname != "t2" {
		t.Fatalf("liked video = %+v", lv)
	}

	var fav FavResourceData
	if err := json.Unmarshal([]byte(`{"info":{"id":1,"fid":2,"mid":3,"title":"f","cover":"c","attr":0,"intro":"i","ctime":4,"mtime":5,"state":0,"media_count":6,"fav_state":0,"like_state":0,"type":0,"link":"l","upper":{"mid":7,"name":"u","face":"uf"}},"medias":[{"id":8,"type":2,"title":"m","cover":"mc","intro":"mi","page":1,"duration":60,"upper":{"mid":9,"name":"um","face":"umf"},"ctime":10,"pubtime":11,"fav_time":12,"attr":0,"ugc":{"bvid":"BV7"},"stat":{"view":1},"cid":13}],"page":{"num":1,"size":20,"count":2,"total":2}}`), &fav); err != nil {
		t.Fatalf("fav resource decode failed: %v", err)
	}
	if fav.Info == nil || fav.Info.Fid != 2 || fav.Info.Upper.Name != "u" {
		t.Fatalf("fav info = %+v", fav.Info)
	}
	if len(fav.Media) != 1 {
		t.Fatalf("medias = %+v", fav.Media)
	}
	m := fav.Media[0]
	if m.ID != 8 || m.Cid != 13 || m.UGC == nil || m.UGC.Bvid != "BV7" || m.Stat == nil || m.Stat.View != 1 {
		t.Fatalf("media item = %+v", m)
	}
	if fav.Page.Num != 1 || fav.Page.Count != 2 || fav.Page.Total != 2 {
		t.Fatalf("page = %+v", fav.Page)
	}

	// Optional pointer fields stay nil when absent.
	var favEmpty FavResourceData
	if err := json.Unmarshal([]byte(`{"medias":[{"id":1}]}`), &favEmpty); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if favEmpty.Info != nil || favEmpty.Media[0].UGC != nil || favEmpty.Media[0].Stat != nil {
		t.Fatalf("expected nil optional fields, got %+v", favEmpty)
	}

	var season SeasonData
	if err := json.Unmarshal([]byte(`{"info":{"id":1},"medias":[{"id":2,"title":"s","cover":"c","duration":30,"pubtime":31,"bvid":"BV8","upper":{"mid":5,"name":"n","face":"f"},"cnt_info":{"view":9}}],"page":{"num":1,"size":20,"count":1,"total":1}}`), &season); err != nil {
		t.Fatalf("season decode failed: %v", err)
	}
	if season.Info == nil || season.Info.ID != 1 {
		t.Fatalf("season info = %+v", season.Info)
	}
	if season.Media[0].Bvid != "BV8" || season.Media[0].CntInfo == nil || season.Media[0].CntInfo.View != 9 {
		t.Fatalf("season media = %+v", season.Media[0])
	}

	var folder FavFolderListData
	if err := json.Unmarshal([]byte(`{"count":1,"list":[{"id":11,"fid":12,"media_count":3,"title":"fd"}]}`), &folder); err != nil {
		t.Fatalf("folder list decode failed: %v", err)
	}
	if folder.Count != 1 || folder.List[0].Fid != 12 || folder.List[0].MediaCount != 3 {
		t.Fatalf("folder list = %+v", folder)
	}

	var card UserCardResponse
	if err := json.Unmarshal([]byte(`{"card":{"mid":"123","name":"up","face":"fc","sign":"sg","level":6,"fans":10,"attention":2,"archive":3}}`), &card); err != nil {
		t.Fatalf("user card decode failed: %v", err)
	}
	if card.Card.Mid != "123" || card.Card.Name != "up" || card.Card.Level != 6 || card.Card.Fans != 10 {
		t.Fatalf("user card = %+v", card.Card)
	}

	var dyn DynamicSpaceResponse
	if err := json.Unmarshal([]byte(`{"has_more":true,"offset":"o-1","items":[{"id_str":"1","type":"DRAW","modules":{"a":1}}]}`), &dyn); err != nil {
		t.Fatalf("dynamic decode failed: %v", err)
	}
	if !dyn.HasMore || dyn.Offset != "o-1" || len(dyn.Items) != 1 {
		t.Fatalf("dynamic = %+v", dyn)
	}
	if dyn.Items[0].IDStr != "1" || dyn.Items[0].Type != "DRAW" || string(dyn.Items[0].Modules) == "" {
		t.Fatalf("dynamic item = %+v", dyn.Items[0])
	}
}
