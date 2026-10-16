package routers

import (
	"net/http"
	"testing"
)

// ---------------------------------------------------------------------------
// Coverage tests for backend/routers/login.go.
//
// login.go reaches Bilibili through hard-coded literal URLs (passport.bilibili
// .com and api.bilibili.com) inside ad-hoc &http.Client values, so the biliapi
// package URL vars cannot be repointed for it. Instead msUseBiliTransport swaps
// http.DefaultTransport for a host-selective stub that only answers requests to
// those two hosts and delegates everything else, keeping the suite fully
// offline. Nothing is ever written to disk except config.yaml, which
// msCfgSnapshot restores.
// ---------------------------------------------------------------------------

func msLoginAPI() regFunc { return RegisterLoginRoutes }

func TestMSLoginGenerateQRCodeSuccess(t *testing.T) {
	msCfgSnapshot(t)
	msUseBiliTransport(t, func(r *http.Request) (*http.Response, error) {
		return msHTTPJSON(200, `{"code":0,"message":"0","data":{"url":"https://passport.bilibili.com/x/ms?q=1","qrcode_key":"msqrkey1"}}`), nil
	})
	e := newAPI(t, msLoginAPI())
	resp := expectCode(t, e, http.MethodGet, "/api/login/qrcode/generate", "", 200)
	if resp.Status != "success" {
		t.Fatalf("status = %q", resp.Status)
	}
	m := resp.dataMap(t)
	if m["qrcode_key"] != "msqrkey1" {
		t.Fatalf("qrcode_key = %v, want msqrkey1", m["qrcode_key"])
	}
	if m["url"] == "" {
		t.Fatalf("url missing in %v", m)
	}
}

func TestMSLoginGenerateQRCodeAPIError(t *testing.T) {
	msCfgSnapshot(t)
	msUseBiliTransport(t, func(r *http.Request) (*http.Response, error) {
		return msHTTPJSON(200, `{"code":-100,"message":" maximize"}`), nil
	})
	e := newAPI(t, msLoginAPI())
	resp := expectStatus(t, e, http.MethodGet, "/api/login/qrcode/generate", "", 400, "error")
	expectMessageContains(t, resp, "B站API返回错误")
}

func TestMSLoginGenerateQRCodeBadJSON(t *testing.T) {
	msCfgSnapshot(t)
	msUseBiliTransport(t, func(r *http.Request) (*http.Response, error) {
		return msHTTPJSON(200, "not-json"), nil
	})
	e := newAPI(t, msLoginAPI())
	resp := expectStatus(t, e, http.MethodGet, "/api/login/qrcode/generate", "", 500, "error")
	expectMessageContains(t, resp, "解析响应失败")
}

func TestMSLoginGenerateQRCodeNetworkError(t *testing.T) {
	msCfgSnapshot(t)
	msUseBiliTransport(t, msHTTPNetworkError("dial refused"))
	e := newAPI(t, msLoginAPI())
	resp := expectStatus(t, e, http.MethodGet, "/api/login/qrcode/generate", "", 500, "error")
	expectMessageContains(t, resp, "网络请求失败")
}

func TestMSLoginQRCodeImageNotImplemented(t *testing.T) {
	msCfgSnapshot(t)
	e := newAPI(t, msLoginAPI())
	resp := expectStatus(t, e, http.MethodGet, "/api/login/qrcode/image", "", 404, "error")
	expectMessageContains(t, resp, "二维码图片接口暂未实现")
}

func TestMSLoginPollMissingKey(t *testing.T) {
	msCfgSnapshot(t)
	e := newAPI(t, msLoginAPI())
	resp := expectStatus(t, e, http.MethodGet, "/api/login/qrcode/poll", "", 400, "error")
	expectMessageContains(t, resp, "缺少必要的qrcode_key参数")
}

func TestMSLoginPollOuterError(t *testing.T) {
	msCfgSnapshot(t)
	msUseBiliTransport(t, func(r *http.Request) (*http.Response, error) {
		return msHTTPJSON(200, `{"code":-100,"message":"expired outer"}`), nil
	})
	e := newAPI(t, msLoginAPI())
	w := doRaw(t, e, http.MethodGet, "/api/login/qrcode/poll?qrcode_key=mskey1", "")
	m := msRawMap(t, w)
	if m["status"] != "error" {
		t.Fatalf("status = %v, want error", m["status"])
	}
	data, _ := m["data"].(map[string]interface{})
	if data == nil || data["message"] != "expired outer" {
		t.Fatalf("data = %v", m["data"])
	}
}

func TestMSLoginPollNotScanned(t *testing.T) {
	msCfgSnapshot(t)
	msUseBiliTransport(t, func(r *http.Request) (*http.Response, error) {
		// outer code 0, inner data.code 86101 = 未扫码
		return msHTTPJSON(200, `{"code":0,"data":{"code":86101,"message":"未扫码"}}`), nil
	})
	e := newAPI(t, msLoginAPI())
	w := doRaw(t, e, http.MethodGet, "/api/login/qrcode/poll?qrcode_key=mskey1", "")
	m := msRawMap(t, w)
	if m["status"] != "success" {
		t.Fatalf("status = %v, want success", m["status"])
	}
	data, _ := m["data"].(map[string]interface{})
	if data["code"].(float64) != 86101 {
		t.Fatalf("data.code = %v, want 86101", data["code"])
	}
}

func TestMSLoginPollSuccessSavesCookiesURLOverride(t *testing.T) {
	msSetCreds(t, "old-sess", "old-jct", "old-dede")
	// URL query values must override the cookie_info values (parseURLCookies
	// writes into the same map after the cookie loop).
	msUseBiliTransport(t, func(r *http.Request) (*http.Response, error) {
		return msHTTPJSON(200, `{
			"code":0,
			"data":{
				"code":0,
				"message":"成功",
				"url":"https://www.bilibili.com?SESSDATA=url-sess&bili_jct=url-jct&DedeUserID=url-dede&junk=x",
				"cookie_info":{"cookies":[
					{"name":"SESSDATA","value":"cookie-sess"},
					{"name":"bili_jct","value":"cookie-jct"},
					{"name":"DedeUserID","value":"cookie-dede"},
					{"name":"DedeUserID__ckMd5","value":"cookie-md5"}
				]}
			}
		}`), nil
	})
	e := newAPI(t, msLoginAPI())
	w := doRaw(t, e, http.MethodGet, "/api/login/qrcode/poll?qrcode_key=mskey1", "")
	m := msRawMap(t, w)
	if m["status"] != "success" {
		t.Fatalf("status = %v body=%s", m["status"], w.Body.String())
	}
	cfg := loadCfg(t)
	if cfg.SESSDATA != "url-sess" {
		t.Fatalf("SESSDATA = %q, want url-sess", cfg.SESSDATA)
	}
	if cfg.BiliJct != "url-jct" {
		t.Fatalf("BiliJct = %q, want url-jct", cfg.BiliJct)
	}
	if cfg.DedeUserID != "url-dede" {
		t.Fatalf("DedeUserID = %q, want url-dede", cfg.DedeUserID)
	}
	if cfg.DedeUserIDCkMd5 != "cookie-md5" {
		t.Fatalf("DedeUserIDCkMd5 = %q, want cookie-md5", cfg.DedeUserIDCkMd5)
	}
}

func TestMSLoginPollCookieOnlyNoURL(t *testing.T) {
	msSetCreds(t, "", "", "")
	msUseBiliTransport(t, func(r *http.Request) (*http.Response, error) {
		return msHTTPJSON(200, `{
			"code":0,
			"data":{
				"code":0,
				"message":"成功",
				"cookie_info":{"cookies":[{"name":"SESSDATA","value":"only-cookie"}]}
			}
		}`), nil
	})
	e := newAPI(t, msLoginAPI())
	w := doRaw(t, e, http.MethodGet, "/api/login/qrcode/poll?qrcode_key=mskey1", "")
	if msRawMap(t, w)["status"] != "success" {
		t.Fatalf("want success")
	}
	if got := loadCfg(t).SESSDATA; got != "only-cookie" {
		t.Fatalf("SESSDATA = %q, want only-cookie", got)
	}
}

func TestMSLoginPollBadJSON(t *testing.T) {
	msCfgSnapshot(t)
	msUseBiliTransport(t, func(r *http.Request) (*http.Response, error) {
		return msHTTPJSON(200, "{broken"), nil
	})
	e := newAPI(t, msLoginAPI())
	resp := expectStatus(t, e, http.MethodGet, "/api/login/qrcode/poll?qrcode_key=mskey1", "", 500, "error")
	expectMessageContains(t, resp, "解析响应失败")
}

func TestMSLoginPollNetworkError(t *testing.T) {
	msCfgSnapshot(t)
	msUseBiliTransport(t, msHTTPNetworkError("dial refused"))
	e := newAPI(t, msLoginAPI())
	resp := expectStatus(t, e, http.MethodGet, "/api/login/qrcode/poll?qrcode_key=mskey1", "", 500, "error")
	expectMessageContains(t, resp, "网络请求失败")
}

func TestMSLoginLogout(t *testing.T) {
	msSetCreds(t, "sess-to-clear", "jct", "dede")
	e := newAPI(t, msLoginAPI())
	w := doRaw(t, e, http.MethodPost, "/api/login/logout", "")
	m := msRawMap(t, w)
	if m["status"] != "success" || m["message"] != "已成功退出登录" {
		t.Fatalf("logout resp = %v", m)
	}
	if got := loadCfg(t).SESSDATA; got != "" {
		t.Fatalf("SESSDATA after logout = %q, want empty", got)
	}
}

func TestMSLoginCheckLoggedOut(t *testing.T) {
	msSetCreds(t, "", "", "")
	e := newAPI(t, msLoginAPI())
	w := doRaw(t, e, http.MethodGet, "/api/login/check", "")
	m := msRawMap(t, w)
	if m["code"].(float64) != -101 || m["message"] != "未登录" {
		t.Fatalf("check logged-out resp = %v", m)
	}
}

func TestMSLoginCheckLoggedInPassthrough(t *testing.T) {
	msSetCreds(t, "sess-live", "", "")
	msUseBiliTransport(t, func(r *http.Request) (*http.Response, error) {
		return msHTTPJSON(200, `{"code":0,"message":"OK","data":{"isLogin":true,"uname":"msuser"}}`), nil
	})
	e := newAPI(t, msLoginAPI())
	w := doRaw(t, e, http.MethodGet, "/api/login/check", "")
	m := msRawMap(t, w)
	if m["code"].(float64) != 0 {
		t.Fatalf("nav passthrough code = %v, want 0 (body=%s)", m["code"], w.Body.String())
	}
	data, _ := m["data"].(map[string]interface{})
	if data["uname"] != "msuser" {
		t.Fatalf("uname = %v, want msuser", data["uname"])
	}
}

func TestMSLoginCheckNetworkError(t *testing.T) {
	msSetCreds(t, "sess-live", "", "")
	msUseBiliTransport(t, msHTTPNetworkError("nav dial refused"))
	e := newAPI(t, msLoginAPI())
	resp := expectStatus(t, e, http.MethodGet, "/api/login/check", "", 500, "error")
	expectMessageContains(t, resp, "网络请求失败")
}

func TestMSLoginCheckAndNotifyNotConfigured(t *testing.T) {
	msSetCreds(t, "", "", "")
	e := newAPI(t, msLoginAPI())
	w := doRaw(t, e, http.MethodGet, "/api/login/check-and-notify", "")
	m := msRawMap(t, w)
	if m["message"] != "SESSDATA 未配置" {
		t.Fatalf("message = %v, want SESSDATA 未配置", m["message"])
	}
	data, _ := m["data"].(map[string]interface{})
	if data["valid"] != false || data["notified"] != false {
		t.Fatalf("data = %v, want valid=false notified=false", data)
	}
}

func TestMSLoginCheckAndNotifyValid(t *testing.T) {
	msSetCreds(t, "sess-live", "", "")
	msUseBiliTransport(t, func(r *http.Request) (*http.Response, error) {
		return msHTTPJSON(200, `{"code":0,"data":{"isLogin":true,"uname":"msuser"}}`), nil
	})
	e := newAPI(t, msLoginAPI())
	w := doRaw(t, e, http.MethodGet, "/api/login/check-and-notify", "")
	m := msRawMap(t, w)
	if m["message"] != "SESSDATA 有效" {
		t.Fatalf("message = %v, want SESSDATA 有效", m["message"])
	}
	data, _ := m["data"].(map[string]interface{})
	if data["valid"] != true || data["username"] != "msuser" {
		t.Fatalf("data = %v", data)
	}
}

func TestMSLoginCheckAndNotifyExpiredNoNotify(t *testing.T) {
	msSetCreds(t, "sess-live", "", "")
	msCfgSnapshot(t) // ensure notify stays disabled
	cfg := loadCfg(t)
	cfg.Notify.Enabled = false
	cfg.Notify.URLs = nil
	msUseBiliTransport(t, func(r *http.Request) (*http.Response, error) {
		return msHTTPJSON(200, `{"code":-101,"message":"账号未登录","data":null}`), nil
	})
	e := newAPI(t, msLoginAPI())
	w := doRaw(t, e, http.MethodGet, "/api/login/check-and-notify", "")
	m := msRawMap(t, w)
	if m["message"] != "SESSDATA 已失效" {
		t.Fatalf("message = %v, want SESSDATA 已失效", m["message"])
	}
	data, _ := m["data"].(map[string]interface{})
	if data["valid"] != false || data["notified"] != false {
		t.Fatalf("data = %v, want valid=false notified=false", data)
	}
}

func TestMSLoginCheckAndNotifyExpiredWithNotify(t *testing.T) {
	msSetCreds(t, "sess-live", "", "")
	titles := msEnableNotifyForTest(t)
	msUseBiliTransport(t, func(r *http.Request) (*http.Response, error) {
		return msHTTPJSON(200, `{"code":-101,"message":"账号未登录","data":{"uname":"msuser"}}`), nil
	})
	e := newAPI(t, msLoginAPI())
	w := doRaw(t, e, http.MethodGet, "/api/login/check-and-notify", "")
	m := msRawMap(t, w)
	data, _ := m["data"].(map[string]interface{})
	if data["notified"] != true {
		t.Fatalf("notified = %v, want true (body=%s)", data["notified"], w.Body.String())
	}
	if data["username"] != "msuser" {
		t.Fatalf("username = %v, want msuser", data["username"])
	}
	got := titles()
	if len(got) == 0 {
		t.Fatalf("expected a SESSDATA expiry notification to be delivered")
	}
}

func TestMSLoginCheckAndNotifyNetworkError(t *testing.T) {
	msSetCreds(t, "sess-live", "", "")
	msUseBiliTransport(t, msHTTPNetworkError("nav dial refused"))
	e := newAPI(t, msLoginAPI())
	resp := expectStatus(t, e, http.MethodGet, "/api/login/check-and-notify", "", 500, "error")
	expectMessageContains(t, resp, "网络请求失败")
}

// 回归 login.go checkAndNotify 的缺陷修复：nav.Data 是指针，{"code":0,
// "data":null} 时它为 nil，此前 code==0 分支直接读取 nav.Data.Username 会
// nil 指针 panic（测试引擎没有 Recovery 中间件，panic 会冲出 ServeHTTP）。
// 现在应照常返回成功信封，username 为空。
func TestMSLoginCheckAndNotifyCodeZeroNullData(t *testing.T) {
	msSetCreds(t, "sess-live", "", "")
	msUseBiliTransport(t, func(r *http.Request) (*http.Response, error) {
		return msHTTPJSON(200, `{"code":0,"message":"OK","data":null}`), nil
	})
	e := newAPI(t, msLoginAPI())
	w := doRaw(t, e, http.MethodGet, "/api/login/check-and-notify", "")
	if w.Code != http.StatusOK {
		t.Fatalf("code = %d, want 200 (body %s)", w.Code, w.Body.String())
	}
	m := msRawMap(t, w)
	if m["message"] != "SESSDATA 有效" {
		t.Fatalf("message = %v, want SESSDATA 有效", m["message"])
	}
	data, _ := m["data"].(map[string]interface{})
	if data["valid"] != true || data["username"] != "" {
		t.Fatalf("data = %v, want valid=true username=\"\"", data)
	}
}
