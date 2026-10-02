package routers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

// apiGroup registers one or more route groups under the same "/api" prefix the
// production router uses, so handler paths and the endpoint registry line up.
type regFunc func(*gin.RouterGroup)

func newAPI(t *testing.T, regs ...regFunc) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	e := gin.New()
	e.UseRawPath = true
	e.UnescapePathValues = true
	api := e.Group("/api")
	for _, reg := range regs {
		reg(api)
	}
	return e
}

type apiResp struct {
	Status  string          `json:"status"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

func (a apiResp) dataMap(t *testing.T) map[string]interface{} {
	t.Helper()
	var m map[string]interface{}
	if len(a.Data) == 0 {
		return nil
	}
	if err := json.Unmarshal(a.Data, &m); err != nil {
		t.Fatalf("decode data as object: %v (raw=%s)", err, a.Data)
	}
	return m
}

func (a apiResp) dataArray(t *testing.T) []interface{} {
	t.Helper()
	if len(a.Data) == 0 {
		return nil
	}
	var v []interface{}
	if err := json.Unmarshal(a.Data, &v); err != nil {
		t.Fatalf("decode data as array: %v (raw=%s)", err, a.Data)
	}
	return v
}

func (a apiResp) rawString(t *testing.T) string {
	t.Helper()
	var s string
	if len(a.Data) == 0 {
		return ""
	}
	if err := json.Unmarshal(a.Data, &s); err != nil {
		t.Fatalf("decode data as string: %v (raw=%s)", err, a.Data)
	}
	return s
}

// do performs a request against the engine and decodes the standard envelope.
// It fails the test when the body is not valid JSON.
func do(t *testing.T, e *gin.Engine, method, path string, body string) (int, apiResp) {
	t.Helper()
	var req *http.Request
	if body == "" {
		req = httptest.NewRequest(method, path, nil)
	} else {
		req = httptest.NewRequest(method, path, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
	}
	w := httptest.NewRecorder()
	e.ServeHTTP(w, req)
	if w.Code == http.StatusNoContent {
		return w.Code, apiResp{}
	}
	var resp apiResp
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("%s %s -> code %d, body %q: %v", method, path, w.Code, w.Body.String(), err)
	}
	return w.Code, resp
}

// doRaw performs a request and returns the raw response (used for file
// serving, SSE and stub endpoints that do not use the envelope).
func doRaw(t *testing.T, e *gin.Engine, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	var req *http.Request
	if body == "" {
		req = httptest.NewRequest(method, path, nil)
	} else {
		req = httptest.NewRequest(method, path, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
	}
	w := httptest.NewRecorder()
	e.ServeHTTP(w, req)
	return w
}

func expectCode(t *testing.T, e *gin.Engine, method, path, body string, want int) apiResp {
	t.Helper()
	code, resp := do(t, e, method, path, body)
	if code != want {
		t.Fatalf("%s %s: http code = %d, want %d (envelope status=%q message=%q)", method, path, code, want, resp.Status, resp.Message)
	}
	return resp
}

// expectStatus checks both the HTTP code and the envelope "status" field, which
// is "success" / "error" as set by models.SuccessResponse / models.ErrorResponse.
func expectStatus(t *testing.T, e *gin.Engine, method, path, body string, wantHTTP int, wantStatus string) apiResp {
	t.Helper()
	resp := expectCode(t, e, method, path, body, wantHTTP)
	if resp.Status != wantStatus {
		t.Fatalf("%s %s: envelope status = %q, want %q (message=%q)", method, path, resp.Status, wantStatus, resp.Message)
	}
	return resp
}

// dataArray converts a decoded envelope field into a slice, failing the test when
// the value is not a JSON array.
func dataArray(t *testing.T, v interface{}) []interface{} {
	t.Helper()
	list, ok := v.([]interface{})
	if !ok {
		t.Fatalf("value = %T, want a JSON array", v)
	}
	return list
}

func expectMessage(t *testing.T, resp apiResp, want string) {
	t.Helper()
	if resp.Message != want {
		t.Fatalf("message = %q, want %q", resp.Message, want)
	}
}

func expectMessageContains(t *testing.T, resp apiResp, want string) {
	t.Helper()
	if !strings.Contains(resp.Message, want) {
		t.Fatalf("message = %q, want substring %q", resp.Message, want)
	}
}
