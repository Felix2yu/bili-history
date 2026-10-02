package scheduler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"bilibili-history-go/database"
)

// TestMain: same harness pattern as the sibling database package tests.
// The scheduler shares process-wide singletons (config.LoadConfig, database
// GetSchedulerDB via sync.Once), so we chdir into a fresh temp dir BEFORE any
// config/database access, drop a minimal config/config.yaml there (cwd-relative
// path wins in GetConfigPath) and namespace all task IDs with "sctest_".
func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "scsched-test-")
	if err != nil {
		panic(err)
	}
	orig, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		panic(err)
	}
	if err := os.Mkdir(filepath.Join(dir, "config"), 0755); err != nil {
		panic(err)
	}
	// Port/timeout values below are asserted by TestGetSchedulerSingleton.
	cfgYAML := "db_file: \"bilibili_history.test.db\"\n" +
		"server:\n" +
		"  port: 18899\n" +
		"scheduler:\n" +
		"  task_timeout: 42\n"
	if err := os.WriteFile(filepath.Join(dir, "config", "config.yaml"), []byte(cfgYAML), 0644); err != nil {
		panic(err)
	}
	if database.GetSchedulerDB() == nil {
		panic("scheduler database failed to initialize")
	}
	// Regression: output/database must be created on demand by the DB layer,
	// not pre-made by the caller — a fresh install used to fail here.
	if _, err := os.Stat(filepath.Join(dir, "output", "database")); err != nil {
		panic(fmt.Sprintf("scheduler db dir not created: %v", err))
	}
	if err := database.RecordExecution("sc_bootstrap", "sc_bootstrap", "completed", "", "", time.Now(), time.Now()); err != nil {
		panic(fmt.Sprintf("scheduler db not writable: %v", err))
	}

	code := m.Run()

	_ = os.Chdir(orig)
	_ = os.RemoveAll(dir)
	os.Exit(code)
}

// ---------------------------------------------------------------------------
// helpers (file-unique prefix: sc)
// ---------------------------------------------------------------------------

func newSCScheduler() *Scheduler {
	return &Scheduler{
		tasks:       make(map[string]*ScheduleTask),
		stopCh:      make(chan struct{}),
		serverPort:  9, // discard port; nothing should dial it
		taskTimeout: 3 * time.Second,
	}
}

type scRecorder struct {
	mu          sync.Mutex
	count       int
	lastMethod  string
	lastURI     string
	lastBody    string
	lastCT      string
	pathsSeen   []string
	lastQueries []string
}

func (r *scRecorder) record(req *http.Request, body string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.count++
	r.lastMethod = req.Method
	r.lastURI = req.URL.RequestURI()
	r.lastBody = body
	r.lastCT = req.Header.Get("Content-Type")
	r.pathsSeen = append(r.pathsSeen, req.URL.Path)
	r.lastQueries = append(r.lastQueries, req.URL.RawQuery)
}

// newSCServer returns a Scheduler whose serverPort targets a local httptest
// server plus the request recorder. No traffic ever leaves the machine.
func newSCServer(t *testing.T) (*Scheduler, *scRecorder) {
	t.Helper()
	rec := &scRecorder{}
	long := strings.Repeat("x", 700)
	mux := http.NewServeMux()
	mux.HandleFunc("/ok", func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		rec.record(r, string(body))
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"success","message":"done"}`))
	})
	mux.HandleFunc("/err", func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		rec.record(r, string(body))
		_, _ = w.Write([]byte(`{"status":"error","message":"boom"}`))
	})
	mux.HandleFunc("/fail", func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		rec.record(r, string(body))
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("server exploded"))
	})
	mux.HandleFunc("/long", func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		rec.record(r, string(body))
		_, _ = w.Write([]byte(long))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	u, err := url.Parse(srv.URL)
	if err != nil {
		t.Fatalf("parse httptest URL: %v", err)
	}
	idx := strings.LastIndex(u.Host, ":")
	port, err := strconv.Atoi(u.Host[idx+1:])
	if err != nil {
		t.Fatalf("parse httptest port: %v", err)
	}
	s := newSCScheduler()
	s.serverPort = port
	return s, rec
}

func scTask(id string, mut func(*ScheduleTask)) *ScheduleTask {
	task := &ScheduleTask{
		ID:           id,
		Name:         id + "-name",
		Endpoint:     "/api/x",
		Method:       "GET",
		ScheduleType: "daily",
		Enabled:      true,
		TaskType:     "main",
		LastStatus:   "idle",
	}
	if mut != nil {
		mut(task)
	}
	return task
}

func scSortIDs(tasks []*ScheduleTask) []string {
	ids := make([]string, 0, len(tasks))
	for _, t := range tasks {
		ids = append(ids, t.ID)
	}
	sort.Strings(ids)
	return ids
}

// scMainTask fetches one row from the scheduler DB by task id.
func scMainTask(t *testing.T, id string) (database.MainTask, bool) {
	t.Helper()
	tasks, err := database.GetMainTasks()
	if err != nil {
		t.Fatalf("GetMainTasks: %v", err)
	}
	for _, mt := range tasks {
		if mt.TaskID == id {
			return mt, true
		}
	}
	return database.MainTask{}, false
}

// ---------------------------------------------------------------------------
// convertTask / buildCronExpr
// ---------------------------------------------------------------------------

func TestBuildCronExpr(t *testing.T) {
	cases := []struct {
		name string
		task ScheduleTask
		want string
	}{
		{"daily with time", ScheduleTask{ScheduleType: "daily", ScheduleTime: "05:30"}, "30 05 * * *"},
		{"daily no time defaults to midnight", ScheduleTask{ScheduleType: "daily"}, "0 0 * * *"},
		{"daily empty time defaults", ScheduleTask{ScheduleType: "daily", ScheduleTime: ""}, "0 0 * * *"},
		{"daily malformed no colon", ScheduleTask{ScheduleType: "daily", ScheduleTime: "0830"}, "0 0 * * *"},
		{"daily malformed three parts", ScheduleTask{ScheduleType: "daily", ScheduleTime: "1:2:3"}, "0 0 * * *"},
		{"daily hour out of range", ScheduleTask{ScheduleType: "daily", ScheduleTime: "25:99"}, "0 0 * * *"},
		{"daily minute out of range", ScheduleTask{ScheduleType: "daily", ScheduleTime: "09:60"}, "0 0 * * *"},
		{"daily negative parts", ScheduleTask{ScheduleType: "daily", ScheduleTime: "-1:-1"}, "0 0 * * *"},
		{"daily non-numeric", ScheduleTask{ScheduleType: "daily", ScheduleTime: "ab:cd"}, "0 0 * * *"},
		{"daily upper bound stays valid", ScheduleTask{ScheduleType: "daily", ScheduleTime: "23:59"}, "59 23 * * *"},
		{"interval minutes", ScheduleTask{ScheduleType: "interval", IntervalValue: 15, IntervalUnit: "minutes"}, "*/15 * * * *"},
		{"interval hours", ScheduleTask{ScheduleType: "interval", IntervalValue: 6, IntervalUnit: "hours"}, "0 */6 * * *"},
		{"interval days", ScheduleTask{ScheduleType: "interval", IntervalValue: 2, IntervalUnit: "days"}, "0 0 */2 * *"},
		{"interval months", ScheduleTask{ScheduleType: "interval", IntervalValue: 3, IntervalUnit: "months"}, "0 0 1 */3 *"},
		{"interval years approximated", ScheduleTask{ScheduleType: "interval", IntervalValue: 1, IntervalUnit: "years"}, "0 0 1 1 *"},
		{"interval non-positive value", ScheduleTask{ScheduleType: "interval", IntervalValue: 0, IntervalUnit: "minutes"}, ""},
		{"interval negative value", ScheduleTask{ScheduleType: "interval", IntervalValue: -5, IntervalUnit: "hours"}, ""},
		{"interval unknown unit", ScheduleTask{ScheduleType: "interval", IntervalValue: 5, IntervalUnit: "weeks"}, ""},
		{"chain has no cron", ScheduleTask{ScheduleType: "chain", ScheduleTime: "10:00"}, ""},
		{"once has no cron", ScheduleTask{ScheduleType: "once", ScheduleTime: "10:00"}, ""},
		{"unknown type", ScheduleTask{ScheduleType: "weird"}, ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := buildCronExpr(c.task); got != c.want {
				t.Errorf("buildCronExpr(%+v) = %q, want %q", c.task, got, c.want)
			}
		})
	}
}

func TestConvertTask(t *testing.T) {
	// Defaults are applied for empty method / schedule type / task type.
	task := convertTask(database.MainTask{TaskID: "sctest_convert", Name: "conv"})
	if task.Method != "GET" {
		t.Errorf("default Method = %q, want GET", task.Method)
	}
	if task.ScheduleType != "daily" {
		t.Errorf("default ScheduleType = %q, want daily", task.ScheduleType)
	}
	if task.TaskType != "main" {
		t.Errorf("default TaskType = %q, want main", task.TaskType)
	}
	if task.LastStatus != "idle" {
		t.Errorf("default LastStatus = %q, want idle", task.LastStatus)
	}
	if task.Enabled {
		t.Error("Enabled=0 must convert to false")
	}
	if task.CronExpr != "0 0 * * *" {
		t.Errorf("CronExpr = %q, want \"0 0 * * *\"", task.CronExpr)
	}

	full := convertTask(database.MainTask{
		TaskID: "sctest_convert2", Name: "f", Endpoint: "/api/f", Method: "POST",
		Params: "p=1", ScheduleType: "daily", ScheduleTime: "06:15",
		IntervalValue: 3, IntervalUnit: "hours", Enabled: 1, TaskType: "sub",
		ParentID: "pp", DependsOn: "dd", CreatedAt: "c", LastModified: "m",
	})
	if !full.Enabled || full.Method != "POST" || full.CronExpr != "15 06 * * *" {
		t.Errorf("full conversion wrong: %+v", full)
	}
	if full.ParentID != "pp" || full.DependsOn != "dd" || full.CreatedAt != "c" || full.LastModified != "m" {
		t.Errorf("fields not copied: %+v", full)
	}
}

// ---------------------------------------------------------------------------
// shouldRun / findDependents
// ---------------------------------------------------------------------------

func TestShouldRun(t *testing.T) {
	s := &Scheduler{}
	now := time.Date(2026, 8, 8, 10, 30, 0, 0, time.Local)

	cases := []struct {
		name string
		task *ScheduleTask
		want bool
	}{
		{"empty cron never runs", scTask("a", func(t *ScheduleTask) { t.CronExpr = "" }), false},
		{"invalid cron never runs", scTask("a", func(t *ScheduleTask) { t.CronExpr = "not a cron" }), false},
		{"non-matching cron", scTask("a", func(t *ScheduleTask) { t.CronExpr = "0 0 * * *" }), false},
		{"matching cron", scTask("a", func(t *ScheduleTask) { t.CronExpr = "30 10 * * *" }), true},
		{
			"matching but already ran this minute",
			scTask("a", func(t *ScheduleTask) {
				t.CronExpr = "30 10 * * *"
				t.LastRunTime = now.Format("2006-01-02 15:04:05")
			}), false,
		},
		{
			"matching, last run yesterday",
			scTask("a", func(t *ScheduleTask) {
				t.CronExpr = "30 10 * * *"
				t.LastRunTime = now.AddDate(0, 0, -1).Format("2006-01-02 15:04:05")
			}), true,
		},
		{
			"matching, same minute different day string parse garbage",
			scTask("a", func(t *ScheduleTask) {
				t.CronExpr = "30 10 * * *"
				t.LastRunTime = "garbage" // unparseable -> no dedupe, allowed
			}), true,
		},
		{
			"matching, last run same time but different year",
			scTask("a", func(t *ScheduleTask) {
				t.CronExpr = "30 10 * * *"
				t.LastRunTime = now.AddDate(-1, 0, 0).Format("2006-01-02 15:04:05")
			}), true,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := s.shouldRun(c.task, now); got != c.want {
				t.Errorf("shouldRun = %v, want %v", got, c.want)
			}
		})
	}
}

func TestFindDependents(t *testing.T) {
	s := newSCScheduler()
	s.tasks["parent"] = scTask("parent", nil)
	s.tasks["c1"] = scTask("c1", func(t *ScheduleTask) { t.DependsOn = "parent" })
	s.tasks["c2"] = scTask("c2", func(t *ScheduleTask) { t.DependsOn = "parent" })
	s.tasks["other"] = scTask("other", func(t *ScheduleTask) { t.DependsOn = "someone-else" })

	got := scSortIDs(s.findDependents("parent"))
	want := []string{"c1", "c2"}
	if len(got) != 2 || got[0] != want[0] || got[1] != want[1] {
		t.Errorf("findDependents(parent) = %v, want %v", got, want)
	}
	if deps := s.findDependents("missing"); len(deps) != 0 {
		t.Errorf("findDependents(missing) = %v, want empty", deps)
	}
}

// ---------------------------------------------------------------------------
// string / param helpers
// ---------------------------------------------------------------------------

func TestTruncate(t *testing.T) {
	if got := truncate("abcdef", 10); got != "abcdef" {
		t.Errorf("short string truncated: %q", got)
	}
	if got := truncate("abcdef", 6); got != "abcdef" {
		t.Errorf("exact-length string truncated: %q", got)
	}
	if got := truncate("abcdef", 3); got != "abc..." {
		t.Errorf("truncate(abcdef,3) = %q, want abc...", got)
	}
}

func TestNormalizeEndpointPath(t *testing.T) {
	cases := []struct{ in, want string }{
		{"", "/"},
		{"   ", "/"},
		{"api/x", "/api/x"},
		{"/x", "/x"},
		{"  /y  ", "/y"},
	}
	for _, c := range cases {
		if got := normalizeEndpointPath(c.in); got != c.want {
			t.Errorf("normalizeEndpointPath(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestBuildQueryString(t *testing.T) {
	if got := buildQueryString(""); got != "" {
		t.Errorf("empty params -> %q", got)
	}
	if got := buildQueryString("  {} "); got != "" {
		t.Errorf("empty JSON object -> %q, want empty", got)
	}
	if got := buildQueryString("a=1&b=2"); got != "a=1&b=2" {
		t.Errorf("query-string passthrough -> %q", got)
	}
	if got := buildQueryString(`{"q":"hello world"}`); got != "q=hello+world" {
		t.Errorf("JSON single key -> %q, want q=hello+world", got)
	}
	// JSON numbers decode to float64; fmt %v renders 5.0 as "5".
	if got := buildQueryString(`{"n":5}`); got != "n=5" {
		t.Errorf("JSON numeric -> %q, want n=5", got)
	}
	// Malformed JSON starting with "{" falls through to passthrough.
	if got := buildQueryString("{bad json"); got != "{bad json" {
		t.Errorf("invalid JSON -> %q, want passthrough", got)
	}
	// JSON arrays do not start with "{" -> passthrough.
	if got := buildQueryString(`["a"]`); got != `["a"]` {
		t.Errorf("JSON array -> %q, want passthrough", got)
	}
}

func TestParseParams(t *testing.T) {
	if v, ok := parseParams("").(map[string]interface{}); !ok || len(v) != 0 {
		t.Errorf("empty params should yield empty map, got %#v", parseParams(""))
	}
	v := parseParams(`{"day":"2026-01-01"}`)
	m, ok := v.(map[string]interface{})
	if !ok || m["day"] != "2026-01-01" {
		t.Errorf("JSON object params -> %#v", v)
	}
	arr, ok := parseParams(`[1,2]`).([]interface{})
	if !ok || len(arr) != 2 {
		t.Errorf("JSON array params -> %#v", v)
	}
	if s, ok := parseParams("plain text").(string); !ok || s != "plain text" {
		t.Errorf("invalid JSON should stay raw string, got %#v", parseParams("plain text"))
	}
}

func TestGetString(t *testing.T) {
	m := map[string]interface{}{
		"a": "",
		"b": "hit",
		"c": 42,
		"d": "second",
	}
	if got := getString(m, "a", "b"); got != "hit" {
		t.Errorf("skip empty string fallback -> %q", got)
	}
	if got := getString(m, "missing", "b"); got != "hit" {
		t.Errorf("missing key fallback -> %q", got)
	}
	if got := getString(m, "c"); got != "" {
		t.Errorf("non-string value -> %q, want empty", got)
	}
	if got := getString(m, "x", "y"); got != "" {
		t.Errorf("all missing -> %q, want empty", got)
	}
	if got := getString(m, "c", "d"); got != "second" {
		t.Errorf("wrong-type then valid -> %q", got)
	}
}

func TestGetInt(t *testing.T) {
	m := map[string]interface{}{
		"f":   float64(7),
		"i":   3,
		"s":   "12",
		"bad": "xx",
		"obj": map[string]interface{}{},
	}
	if got := getInt(m, "f"); got != 7 {
		t.Errorf("float64 -> %d", got)
	}
	if got := getInt(m, "i"); got != 3 {
		t.Errorf("int -> %d", got)
	}
	if got := getInt(m, "s"); got != 12 {
		t.Errorf("numeric string -> %d", got)
	}
	if got := getInt(m, "bad", "s"); got != 12 {
		t.Errorf("invalid string should fall through to next key -> %d", got)
	}
	if got := getInt(m, "obj", "missing"); got != 0 {
		t.Errorf("wrong type/missing -> %d, want 0", got)
	}
	if got := getInt(m, "zzz"); got != 0 {
		t.Errorf("missing key -> %d, want 0", got)
	}
}

func TestParamsToStringAndGetParamsString(t *testing.T) {
	if got := paramsToString(nil); got != "" {
		t.Errorf("nil params -> %q", got)
	}
	if got := paramsToString("raw=1"); got != "raw=1" {
		t.Errorf("string params -> %q", got)
	}
	got := paramsToString(map[string]interface{}{"k": "v"})
	var back map[string]interface{}
	if err := json.Unmarshal([]byte(got), &back); err != nil || back["k"] != "v" {
		t.Errorf("map params -> %q (err %v)", got, err)
	}
	// Unmarshalable value (channel) -> empty string.
	if got := paramsToString(make(chan int)); got != "" {
		t.Errorf("unmarshalable params -> %q, want empty", got)
	}
	if got := getParamsString(map[string]interface{}{}); got != "" {
		t.Errorf("missing params key -> %q", got)
	}
	if got := getParamsString(map[string]interface{}{"params": "a=b"}); got != "a=b" {
		t.Errorf("params key -> %q", got)
	}
}

// ---------------------------------------------------------------------------
// frontend formatting: GetTasks / GetTask / GetStatus
// ---------------------------------------------------------------------------

func TestGetTasksFrontendFormat(t *testing.T) {
	s := newSCScheduler()
	s.tasks["new"] = scTask("new", func(t *ScheduleTask) {
		t.Name = "N"
		t.Params = `{"day":"2026"}`
		t.CreatedAt = "2026-01-02 00:00:00"
		t.LastModified = "lm"
		t.NextRunTime = "2026-02-01 00:00:00"
	})
	s.tasks["old"] = scTask("old", func(t *ScheduleTask) {
		t.CreatedAt = "2025-01-01 00:00:00"
		t.DependsOn = "new"
		t.Params = "not json"
	})
	s.tasks["kid"] = scTask("kid", func(t *ScheduleTask) {
		t.TaskType = "sub"
		t.ParentID = "new"
		t.CreatedAt = "2026-01-03 00:00:00"
	})
	s.tasks["kid2"] = scTask("kid2", func(t *ScheduleTask) {
		t.TaskType = "sub"
		t.ParentID = "new"
		t.CreatedAt = "2026-01-01 00:00:00" // older than kid
	})

	tasks := s.GetTasks()
	if len(tasks) != 2 {
		t.Fatalf("GetTasks should exclude sub-tasks, got %d", len(tasks))
	}
	// sorted by created_at: old first
	if tasks[0]["task_id"] != "old" || tasks[1]["task_id"] != "new" {
		t.Errorf("GetTasks order = %v, %v; want old, new", tasks[0]["task_id"], tasks[1]["task_id"])
	}
	if tasks[0]["depends_on"] != "new" {
		t.Errorf("depends_on missing: %v", tasks[0])
	}
	if _, ok := tasks[1]["depends_on"]; ok {
		t.Error("depends_on should be absent when empty")
	}
	if _, ok := tasks[1]["parent_id"]; ok {
		t.Error("parent_id should be absent when empty")
	}

	cfg := tasks[1]["config"].(map[string]interface{})
	if cfg["name"] != "N" || cfg["enabled"] != true || cfg["method"] != "GET" {
		t.Errorf("config map wrong: %#v", cfg)
	}
	if p, ok := cfg["params"].(map[string]interface{}); !ok || p["day"] != "2026" {
		t.Errorf("JSON params not parsed: %#v", cfg["params"])
	}
	if p, ok := tasks[0]["config"].(map[string]interface{})["params"].(string); !ok || p != "not json" {
		t.Errorf("raw params should stay string: %#v", tasks[0])
	}
	ex := tasks[1]["execution"].(map[string]interface{})
	if ex["next_run"] != "2026-02-01 00:00:00" || ex["status"] != "idle" {
		t.Errorf("execution map wrong: %#v", ex)
	}

	subs := tasks[1]["sub_tasks"].([]map[string]interface{})
	if len(subs) != 2 {
		t.Fatalf("sub_tasks len = %d, want 2", len(subs))
	}
	if subs[0]["task_id"] != "kid2" || subs[1]["task_id"] != "kid" {
		t.Errorf("sub tasks not sorted by created_at: %v %v", subs[0]["task_id"], subs[1]["task_id"])
	}
	if subs[0]["parent_id"] != "new" {
		t.Errorf("sub task parent_id missing: %#v", subs[0])
	}
	emptySubs := tasks[0]["sub_tasks"]
	if arr, ok := emptySubs.([]interface{}); !ok || len(arr) != 0 {
		t.Errorf("no-subs task should have empty []interface{}, got %#v", emptySubs)
	}
}

func TestGetTaskAndErrors(t *testing.T) {
	s := newSCScheduler()
	s.tasks["one"] = scTask("one", nil)
	m, err := s.GetTask("one")
	if err != nil {
		t.Fatalf("GetTask(one): %v", err)
	}
	if m["task_id"] != "one" || m["task_type"] != "main" {
		t.Errorf("GetTask payload wrong: %#v", m)
	}
	if _, err := s.GetTask("missing"); err == nil {
		t.Error("GetTask(missing) should error")
	}
}

func TestGetStatus(t *testing.T) {
	s := newSCScheduler()
	if got := s.GetStatus()["total_tasks"]; got != 0 {
		t.Errorf("empty total_tasks = %v", got)
	}
	s.tasks["a"] = scTask("a", func(t *ScheduleTask) { t.Enabled = true; t.Running = true })
	s.tasks["b"] = scTask("b", func(t *ScheduleTask) { t.Enabled = true; t.Running = false })
	s.tasks["c"] = scTask("c", func(t *ScheduleTask) { t.Enabled = false; t.Running = true })
	s.tasks["d"] = scTask("d", func(t *ScheduleTask) { t.Enabled = false })
	st := s.GetStatus()
	if st["total_tasks"] != 4 || st["running_tasks"] != 2 || st["enabled_tasks"] != 2 {
		t.Errorf("GetStatus = %#v", st)
	}
	if st["running"] != false {
		t.Errorf("scheduler daemon flag should be false, got %#v", st["running"])
	}
}

// ---------------------------------------------------------------------------
// loadTasks
// ---------------------------------------------------------------------------

func TestLoadTasks(t *testing.T) {
	// Seed deterministic rows with namespaced IDs, then load. The test DB is
	// process-shared so we assert membership, never exact counts.
	a := database.MainTask{TaskID: "sctest_load_a", Name: "A", Endpoint: "/api/a", Method: "GET", ScheduleType: "daily", ScheduleTime: "01:02", Enabled: 1, TaskType: "main"}
	b := database.MainTask{TaskID: "sctest_load_b", Name: "B", Endpoint: "/api/b", Method: "POST", ScheduleType: "interval", IntervalValue: 10, IntervalUnit: "minutes", Enabled: 0, TaskType: "main"}
	if err := database.UpsertMainTask(a); err != nil {
		t.Fatalf("upsert a: %v", err)
	}
	if err := database.UpsertMainTask(b); err != nil {
		t.Fatalf("upsert b: %v", err)
	}
	if err := database.UpdateTaskStatus("sctest_load_a", "running", "", 2.5, true); err != nil {
		t.Fatalf("status a: %v", err)
	}
	if err := database.UpdateTaskStatus("sctest_load_b", "failed", "oops", 1.0, false); err != nil {
		t.Fatalf("status b: %v", err)
	}

	s := newSCScheduler()
	s.loadTasks()

	ta, ok := s.tasks["sctest_load_a"]
	if !ok {
		t.Fatal("task a not loaded")
	}
	if !ta.Running {
		t.Error("last_status running should set Running flag")
	}
	if ta.TotalRuns != 1 || ta.SuccessRuns != 1 || ta.SuccessRate != 100 {
		t.Errorf("runtime stats not merged: %+v", ta)
	}
	if ta.CronExpr != "02 01 * * *" {
		t.Errorf("loaded task cron = %q", ta.CronExpr)
	}
	tb, ok := s.tasks["sctest_load_b"]
	if !ok {
		t.Fatal("task b not loaded")
	}
	if tb.Running || tb.LastError != "oops" || tb.FailRuns != 1 {
		t.Errorf("task b merge wrong: %+v", tb)
	}
	if tb.LastRunTime == "" {
		t.Error("LastRunTime should be merged from status map")
	}
}

// ---------------------------------------------------------------------------
// GetScheduler singleton (config plumbing)
// ---------------------------------------------------------------------------

func TestGetSchedulerSingleton(t *testing.T) {
	s := GetScheduler()
	if s == nil {
		t.Fatal("GetScheduler returned nil")
	}
	if again := GetScheduler(); again != s {
		t.Error("GetScheduler must return the same singleton")
	}
	// Values asserted here come from the config/config.yaml written in TestMain.
	if s.serverPort != 18899 {
		t.Errorf("serverPort = %d, want 18899 from config", s.serverPort)
	}
	if s.taskTimeout != 42*time.Second {
		t.Errorf("taskTimeout = %v, want 42s from config", s.taskTimeout)
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, id := range []string{"sessdata_health_check", "sync_likes", "fetch_history", "import_data", "analyze_data", "send_daily_report"} {
		if _, ok := s.tasks[id]; !ok {
			t.Errorf("default task %s missing after GetScheduler init", id)
		}
	}
	if imp := s.tasks["import_data"]; imp.DependsOn != "fetch_history" {
		t.Errorf("import_data DependsOn = %q, want fetch_history", imp.DependsOn)
	}
}

// ---------------------------------------------------------------------------
// initDefaultTasks on an isolated scheduler: seeding + migrations
// ---------------------------------------------------------------------------

func TestInitDefaultTasksSeedAndMigrations(t *testing.T) {
	s := newSCScheduler()

	oldFetch := database.MainTask{
		TaskID: "fetch_history", Name: "旧版获取", Endpoint: "/fetch/bili-history",
		Method: "GET", ScheduleType: "daily", ScheduleTime: "00:00", Enabled: 0, TaskType: "main",
	}
	oldImport := database.MainTask{
		TaskID: "import_data", Name: "旧版导入", Endpoint: "/api/importSqlite/import_data_sqlite",
		Method: "POST", ScheduleType: "chain", DependsOn: "wrong_old_parent", Enabled: 0, TaskType: "main",
	}
	custom := database.MainTask{
		TaskID: "sctest_legacy_ep", Name: "旧路径", Endpoint: "/viewing/stats",
		Method: "GET", ScheduleType: "daily", Enabled: 0, TaskType: "main",
	}
	urlTask := database.MainTask{
		TaskID: "sctest_url_ep", Name: "外部URL", Endpoint: "http://example.invalid/x",
		Method: "GET", ScheduleType: "daily", Enabled: 0, TaskType: "main",
	}
	slashless := database.MainTask{
		TaskID: "sctest_slashless_ep", Name: "无斜杠", Endpoint: "raw_path",
		Method: "GET", ScheduleType: "daily", Enabled: 0, TaskType: "main",
	}
	empty := database.MainTask{
		TaskID: "sctest_empty_ep", Name: "空", Endpoint: "",
		Method: "GET", ScheduleType: "daily", Enabled: 0, TaskType: "main",
	}
	for _, mt := range []database.MainTask{oldFetch, oldImport, custom, urlTask, slashless, empty} {
		if err := database.UpsertMainTask(mt); err != nil {
			t.Fatalf("seed %s: %v", mt.TaskID, err)
		}
		s.tasks[mt.TaskID] = convertTask(mt)
	}

	s.initDefaultTasks()

	// Endpoint migration from the defaults table (fetch_history lacks /api).
	if got := s.tasks["fetch_history"].Endpoint; got != "/api/fetch/bili-history" {
		t.Errorf("fetch_history endpoint = %q, want /api/fetch/bili-history", got)
	}
	if mt, ok := scMainTask(t, "fetch_history"); !ok || mt.Endpoint != "/api/fetch/bili-history" {
		t.Errorf("fetch_history DB endpoint = %+v", mt)
	}
	// DependsOn migration from the defaults table.
	if got := s.tasks["import_data"].DependsOn; got != "fetch_history" {
		t.Errorf("import_data DependsOn = %q, want fetch_history", got)
	}
	if mt, ok := scMainTask(t, "import_data"); !ok || mt.DependsOn != "fetch_history" {
		t.Errorf("import_data DB DependsOn = %+v", mt)
	}
	// Generic /api prefix sweep for user-defined tasks.
	if got := s.tasks["sctest_legacy_ep"].Endpoint; got != "/api/viewing/stats" {
		t.Errorf("legacy endpoint = %q, want /api/viewing/stats", got)
	}
	if mt, ok := scMainTask(t, "sctest_legacy_ep"); !ok || mt.Endpoint != "/api/viewing/stats" {
		t.Errorf("legacy DB endpoint = %+v", mt)
	}
	// Absolute URLs, slash-less and empty endpoints must be left alone.
	if got := s.tasks["sctest_url_ep"].Endpoint; !strings.HasPrefix(got, "http://example.invalid") {
		t.Errorf("absolute URL endpoint changed: %q", got)
	}
	if got := s.tasks["sctest_slashless_ep"].Endpoint; got != "raw_path" {
		t.Errorf("slash-less endpoint changed: %q", got)
	}
	if got := s.tasks["sctest_empty_ep"].Endpoint; got != "" {
		t.Errorf("empty endpoint changed: %q", got)
	}
	// The remaining defaults were absent from memory and get seeded.
	for _, id := range []string{"sessdata_health_check", "sync_likes", "analyze_data", "send_daily_report"} {
		if _, ok := s.tasks[id]; !ok {
			t.Errorf("default task %s not seeded", id)
		}
	}
	// Idempotent second call: nothing changes, "all present" log branch.
	before := s.tasks["fetch_history"].Endpoint
	s.initDefaultTasks()
	if s.tasks["fetch_history"].Endpoint != before {
		t.Error("second initDefaultTasks must not re-modify endpoints")
	}
}

// ---------------------------------------------------------------------------
// Start / Stop lifecycle (no ticker firing: 60s period, immediate stop)
// ---------------------------------------------------------------------------

func TestSchedulerStartStop(t *testing.T) {
	s := newSCScheduler()
	s.Start()
	s.mu.RLock()
	running := s.running
	s.mu.RUnlock()
	if !running {
		t.Fatal("scheduler should be running after Start")
	}
	// Second Start must hit the already-running guard.
	s.Start()
	// Stop flips the flag and waits for the run goroutine to see stopCh.
	s.Stop()
	s.mu.RLock()
	running = s.running
	s.mu.RUnlock()
	if running {
		t.Error("scheduler must not be running after Stop")
	}
	// Second Stop hits the not-running guard and returns.
	s.Stop()
}

// ---------------------------------------------------------------------------
// callEndpoint / doRequest against a local httptest server
// ---------------------------------------------------------------------------

func TestCallEndpointGet(t *testing.T) {
	s, rec := newSCServer(t)

	res, err := s.callEndpoint("GET", "/ok", "a=1&b=2")
	if err != nil {
		t.Fatalf("GET /ok: %v", err)
	}
	if !strings.Contains(res, "success") {
		t.Errorf("GET result = %q", res)
	}
	rec.mu.Lock()
	if rec.lastMethod != "GET" || rec.lastURI != "/ok?a=1&b=2" {
		t.Errorf("recorded request = %s %s", rec.lastMethod, rec.lastURI)
	}
	rec.mu.Unlock()

	// Empty method behaves like GET.
	if _, err := s.callEndpoint("", "/ok", ""); err != nil {
		t.Fatalf("empty method should GET: %v", err)
	}

	// Endpoint that already contains a query gets "&" appended.
	if _, err := s.callEndpoint("GET", "/ok?x=1", "y=2"); err != nil {
		t.Fatalf("GET with existing query: %v", err)
	}
	rec.mu.Lock()
	defer rec.mu.Unlock()
	if rec.lastURI != "/ok?x=1&y=2" {
		t.Errorf("existing-query URI = %q, want /ok?x=1&y=2", rec.lastURI)
	}
}

func TestCallEndpointPost(t *testing.T) {
	s, rec := newSCServer(t)

	if _, err := s.callEndpoint("post", "/ok", `{"k":"v"}`); err != nil {
		t.Fatalf("POST: %v", err)
	}
	rec.mu.Lock()
	if rec.lastMethod != "POST" || rec.lastBody != `{"k":"v"}` || rec.lastCT != "application/json" {
		t.Errorf("POST record: method=%q body=%q ct=%q", rec.lastMethod, rec.lastBody, rec.lastCT)
	}
	rec.mu.Unlock()

	// Empty POST body is replaced by "{}".
	if _, err := s.callEndpoint("PUT", "/ok", ""); err != nil {
		t.Fatalf("PUT: %v", err)
	}
	rec.mu.Lock()
	defer rec.mu.Unlock()
	if rec.lastMethod != "PUT" || rec.lastBody != "{}" {
		t.Errorf("PUT record: method=%q body=%q", rec.lastMethod, rec.lastBody)
	}
}

func TestCallEndpointErrors(t *testing.T) {
	s, _ := newSCServer(t)

	if _, err := s.callEndpoint("GET", "", ""); err == nil || !strings.Contains(err.Error(), "endpoint") {
		t.Errorf("empty endpoint error = %v", err)
	}
	// Error envelope in a 200 response is a task failure.
	res, err := s.callEndpoint("GET", "/err", "")
	if err == nil || !strings.Contains(err.Error(), "boom") {
		t.Errorf("error envelope err = %v", err)
	}
	if !strings.Contains(res, "error") {
		t.Errorf("error envelope body should still be returned, got %q", res)
	}
	// HTTP >= 400 is a failure with the truncated body in the message.
	res, err = s.callEndpoint("GET", "/fail", "")
	if err == nil || !strings.Contains(err.Error(), "HTTP 500") {
		t.Errorf("500 err = %v", err)
	}
	if res != "server exploded" {
		t.Errorf("500 body = %q", res)
	}
	// An invalid method fails http.NewRequest validation.
	if _, err := s.callEndpoint("P OST", "/ok", ""); err == nil {
		t.Error("invalid method should error")
	}
	// An invalid URL escape fails http.NewRequest in the GET branch.
	if _, err := s.callEndpoint("GET", "/%zz", ""); err == nil {
		t.Error("invalid URL should error")
	}
}

func TestCallEndpointNetworkError(t *testing.T) {
	// Spin up a local httptest server, close the listener, then dial its port:
	// the connection is refused immediately on loopback, no external traffic.
	probe, _ := newSCServer(t)
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	u, err := url.Parse(srv.URL)
	if err != nil {
		t.Fatalf("parse URL: %v", err)
	}
	idx := strings.LastIndex(u.Host, ":")
	closedPort, err := strconv.Atoi(u.Host[idx+1:])
	if err != nil {
		t.Fatalf("parse port: %v", err)
	}
	srv.Close()

	s := newSCScheduler()
	s.serverPort = closedPort
	if _, err := s.callEndpoint("GET", "/ok", ""); err == nil {
		t.Errorf("expected connection error against closed port %d", closedPort)
	}
	// sanity: probe scheduler still points at a live server
	if _, err := probe.callEndpoint("GET", "/ok", ""); err != nil {
		t.Errorf("live server unexpectedly failed: %v", err)
	}
}

// ---------------------------------------------------------------------------
// executeTask (synchronous) incl. chain triggering and persistence
// ---------------------------------------------------------------------------

func TestExecuteTaskSuccessWithChain(t *testing.T) {
	s, rec := newSCServer(t)
	s.tasks["sctest_exec_parent"] = scTask("sctest_exec_parent", func(t *ScheduleTask) {
		t.Endpoint = "/ok"
	})
	s.tasks["sctest_exec_child"] = scTask("sctest_exec_child", func(t *ScheduleTask) {
		t.Endpoint = "/ok"
		t.ScheduleType = "chain"
		t.DependsOn = "sctest_exec_parent"
	})

	s.executeTask("sctest_exec_parent", true)

	p := s.tasks["sctest_exec_parent"]
	if p.LastStatus != "completed" || p.Running || p.LastError != "" {
		t.Errorf("parent after run: %+v", p)
	}
	if p.TotalRuns != 1 || p.SuccessRuns != 1 || p.SuccessRate != 100 {
		t.Errorf("parent stats not reloaded from DB: %+v", p)
	}
	c := s.tasks["sctest_exec_child"]
	if c.LastStatus != "completed" {
		t.Errorf("chain child not triggered: %+v", c)
	}
	rec.mu.Lock()
	n := rec.count
	rec.mu.Unlock()
	if n != 2 {
		t.Errorf("endpoint called %d times, want 2 (parent + child)", n)
	}

	// DB persistence: status row + execution history with stored result.
	st, err := database.GetTaskStatusMap()
	if err != nil {
		t.Fatalf("status map: %v", err)
	}
	if st["sctest_exec_parent"].LastStatus != "completed" {
		t.Errorf("DB status row: %+v", st["sctest_exec_parent"])
	}
	hist := s.GetTaskExecutions("sctest_exec_parent", 10)
	if len(hist) == 0 {
		t.Fatal("execution history not recorded")
	}
	if hist[0]["status"] != "completed" {
		t.Errorf("history row: %#v", hist[0])
	}

	// triggerChain=false must run the parent but not the child.
	s.tasks["sctest_exec_child"].LastStatus = "idle"
	s.executeTask("sctest_exec_parent", false)
	if s.tasks["sctest_exec_child"].LastStatus != "idle" {
		t.Error("child must not run when triggerChain=false")
	}

	// Unknown task id and already-running guard are no-ops.
	s.executeTask("sctest_exec_missing", true)
	s.tasks["sctest_exec_child"].Running = true
	s.executeTask("sctest_exec_child", true)
	if s.tasks["sctest_exec_child"].LastStatus != "idle" {
		t.Error("running task must not execute again")
	}
	s.tasks["sctest_exec_child"].Running = false
}

func TestExecuteTaskFailureAndTruncation(t *testing.T) {
	s, _ := newSCServer(t)
	s.tasks["sctest_exec_fail"] = scTask("sctest_exec_fail", func(t *ScheduleTask) {
		t.Endpoint = "/fail"
	})
	s.tasks["sctest_exec_failchild"] = scTask("sctest_exec_failchild", func(t *ScheduleTask) {
		t.Endpoint = "/ok"
		t.DependsOn = "sctest_exec_fail"
	})
	s.tasks["sctest_exec_long"] = scTask("sctest_exec_long", func(t *ScheduleTask) {
		t.Endpoint = "/long"
	})

	s.executeTask("sctest_exec_fail", true)
	f := s.tasks["sctest_exec_fail"]
	if f.LastStatus != "failed" || !strings.Contains(f.LastError, "HTTP 500") {
		t.Errorf("failed task state: %+v", f)
	}
	if f.TotalRuns != 1 || f.FailRuns != 1 || f.SuccessRate != 0 {
		t.Errorf("failed task stats: %+v", f)
	}
	if s.tasks["sctest_exec_failchild"].LastStatus != "idle" {
		t.Error("dependents must not run after failure")
	}

	// Long responses are truncated to 500 bytes in execution history.
	s.executeTask("sctest_exec_long", true)
	hist := s.GetTaskExecutions("sctest_exec_long", 5)
	if len(hist) == 0 {
		t.Fatal("no history for long task")
	}
	got, _ := hist[0]["result"].(string)
	if len(got) != 500 {
		t.Errorf("stored result length = %d, want 500", len(got))
	}
}

// TestExecuteTaskConcurrentReaders runs a task while other goroutines read task
// state through the exported accessors. It exists for `go test -race`: executeTask
// used to read task.Name and task.LastStatus after releasing s.mu, which races
// with those readers. Without the race detector it only asserts the run completes
// and the task ends in a sane state.
func TestExecuteTaskConcurrentReaders(t *testing.T) {
	s, _ := newSCServer(t)
	s.tasks["sctest_exec_race"] = scTask("sctest_exec_race", func(t *ScheduleTask) {
		t.Endpoint = "/ok"
	})

	stop := make(chan struct{})
	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-stop:
					return
				default:
					_ = s.GetTasks()
					_, _ = s.GetTask("sctest_exec_race")
				}
			}
		}()
	}

	s.executeTask("sctest_exec_race", false)
	close(stop)
	wg.Wait()

	if st := s.tasks["sctest_exec_race"]; st.LastStatus != "completed" || st.Running {
		t.Errorf("task after concurrent readers: %+v", st)
	}
}

// ---------------------------------------------------------------------------
// checkAndRunTasks / RunTask (async paths, local httptest only)
// ---------------------------------------------------------------------------

// scWaitFor polls cond every 20ms, failing after timeout.
func scWaitFor(t *testing.T, timeout time.Duration, cond func() bool) bool {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if cond() {
			return true
		}
		time.Sleep(20 * time.Millisecond)
	}
	return false
}

func TestCheckAndRunTasks(t *testing.T) {
	s, _ := newSCServer(t)

	s.tasks["disabled"] = scTask("disabled", func(t *ScheduleTask) { t.Enabled = false })
	s.tasks["busy"] = scTask("busy", func(t *ScheduleTask) { t.Running = true })
	s.tasks["chain"] = scTask("chain", func(t *ScheduleTask) { t.ScheduleType = "chain" })
	s.tasks["nocron"] = scTask("nocron", func(t *ScheduleTask) { t.CronExpr = "" })
	target := scTask("match", func(t *ScheduleTask) { t.Endpoint = "/ok" })
	s.tasks["match"] = target

	// The cron expression must match the current minute when the picker runs;
	// rebuild it each attempt in case the minute rolls over mid-test. The
	// budget is deliberately generous: with -race on a loaded machine the
	// executeTask HTTP round trip can take longer than a couple of seconds,
	// and a tight deadline made this test flake.
	deadline := time.Now().Add(20 * time.Second)
	for time.Now().Before(deadline) {
		now := time.Now()
		target.CronExpr = fmt.Sprintf("%d %d * * *", now.Minute(), now.Hour())
		target.LastRunTime = ""
		s.checkAndRunTasks()
		if scWaitFor(t, 2*time.Second, func() bool {
			s.mu.RLock()
			defer s.mu.RUnlock()
			return target.LastStatus != "idle"
		}) {
			break
		}
	}

	s.mu.RLock()
	status := target.LastStatus
	s.mu.RUnlock()
	if status != "completed" {
		t.Errorf("matching task LastStatus = %q, want completed", status)
	}
	for _, id := range []string{"disabled", "busy", "chain", "nocron"} {
		s.mu.RLock()
		st := s.tasks[id].LastStatus
		s.mu.RUnlock()
		if st != "idle" {
			t.Errorf("task %s should have been skipped, status=%q", id, st)
		}
	}
}

func TestRunTask(t *testing.T) {
	s, _ := newSCServer(t)
	if err := s.RunTask("sctest_rt_missing"); err == nil {
		t.Error("RunTask(missing) should error")
	}
	s.tasks["sctest_rt"] = scTask("sctest_rt", func(t *ScheduleTask) { t.Endpoint = "/ok" })
	busy := scTask("sctest_rt_busy", func(t *ScheduleTask) { t.Endpoint = "/ok"; t.Running = true })
	s.tasks["sctest_rt_busy"] = busy
	if err := s.RunTask("sctest_rt_busy"); err == nil || !strings.Contains(err.Error(), "running") {
		t.Errorf("RunTask(running task) = %v, want already-running error", err)
	}
	if err := s.RunTask("sctest_rt"); err != nil {
		t.Fatalf("RunTask: %v", err)
	}
	if !scWaitFor(t, 3*time.Second, func() bool {
		s.mu.RLock()
		defer s.mu.RUnlock()
		return s.tasks["sctest_rt"].LastStatus == "completed"
	}) {
		t.Error("RunTask did not complete in time")
	}
}

// ---------------------------------------------------------------------------
// CRUD API
// ---------------------------------------------------------------------------

func TestCreateTaskFromConfig(t *testing.T) {
	s := newSCScheduler()

	if _, err := s.CreateTaskFromConfig(map[string]interface{}{}); err == nil {
		t.Error("missing task_id must error")
	}

	nested := map[string]interface{}{
		"task_id":    "sctest_create_nested",
		"task_type":  "sub",
		"parent_id":  "sctest_parent_x",
		"depends_on": "dep1",
		"config": map[string]interface{}{
			"name":          "Nested",
			"endpoint":      "api/custom", // missing leading slash
			"method":        "POST",
			"params":        map[string]interface{}{"day": 3},
			"schedule_type": "interval",
			"interval":      float64(5), // alias key for interval_value
			"interval_unit": "minutes",
			"enabled":       true,
		},
	}
	out, err := s.CreateTaskFromConfig(nested)
	if err != nil {
		t.Fatalf("CreateTaskFromConfig(nested): %v", err)
	}
	if out["task_id"] != "sctest_create_nested" || out["task_type"] != "sub" {
		t.Errorf("nested output: %#v", out)
	}
	task := s.tasks["sctest_create_nested"]
	if task.Endpoint != "/api/custom" {
		t.Errorf("endpoint not normalized: %q", task.Endpoint)
	}
	if task.IntervalValue != 5 || task.CronExpr != "*/5 * * * *" {
		t.Errorf("interval alias/cron: value=%d cron=%q", task.IntervalValue, task.CronExpr)
	}
	if !task.Enabled || task.Method != "POST" || task.ParentID != "sctest_parent_x" || task.DependsOn != "dep1" {
		t.Errorf("nested fields: %+v", task)
	}
	if task.Params != `{"day":3}` {
		t.Errorf("params marshaled = %q", task.Params)
	}
	if mt, ok := scMainTask(t, "sctest_create_nested"); !ok || mt.Name != "Nested" {
		t.Errorf("row not persisted: %+v", mt)
	}

	flat := map[string]interface{}{
		"task_id":       "sctest_create_flat",
		"name":          "Flat",
		"endpoint":      "/api/flat",
		"schedule_type": "daily",
		"schedule_time": "07:45",
		"params":        "k=v",
		"depends_on":    []interface{}{"d1", "d2"},
	}
	if _, err := s.CreateTaskFromConfig(flat); err != nil {
		t.Fatalf("CreateTaskFromConfig(flat): %v", err)
	}
	ft := s.tasks["sctest_create_flat"]
	if ft.DependsOn != "d1,d2" {
		t.Errorf("depends_on list join = %q", ft.DependsOn)
	}
	if ft.CronExpr != "45 07 * * *" || ft.Method != "GET" || ft.TaskType != "main" {
		t.Errorf("flat defaults: %+v", ft)
	}
	if ft.Enabled {
		t.Error("missing enabled key must stay disabled")
	}
}

func TestUpdateTaskFromConfig(t *testing.T) {
	s := newSCScheduler()
	base := database.MainTask{
		TaskID: "sctest_update", Name: "Old", Endpoint: "/api/old", Method: "GET",
		ScheduleType: "daily", ScheduleTime: "08:00", Enabled: 0, TaskType: "main",
	}
	if err := database.UpsertMainTask(base); err != nil {
		t.Fatalf("seed: %v", err)
	}
	s.tasks["sctest_update"] = convertTask(base)

	if _, err := s.UpdateTaskFromConfig("sctest_no_such", map[string]interface{}{}); err == nil {
		t.Error("update of missing task must error")
	}

	payload := map[string]interface{}{
		"enabled": true, // top-level override wins over config.enabled=false
		"config": map[string]interface{}{
			"name":           "Renamed",
			"endpoint":       "api/new",
			"method":         "POST",
			"params":         map[string]interface{}{"p": 1},
			"schedule_type":  "interval",
			"interval_value": float64(30),
			"interval_unit":  "minutes",
			"enabled":        false,
			"unknown_key":    "ignored",
		},
	}
	out, err := s.UpdateTaskFromConfig("sctest_update", payload)
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	task := s.tasks["sctest_update"]
	if task.Name != "Renamed" || task.Endpoint != "/api/new" || task.Method != "POST" {
		t.Errorf("update not applied: %+v", task)
	}
	if task.ScheduleType != "interval" || task.CronExpr != "*/30 * * * *" {
		t.Errorf("schedule change not reflected: %+v", task)
	}
	if !task.Enabled {
		t.Error("top-level enabled=true must override config.enabled=false")
	}
	if task.Params != `{"p":1}` {
		t.Errorf("params = %q", task.Params)
	}
	if out["task_id"] != "sctest_update" {
		t.Errorf("output: %#v", out)
	}

	// Flat payload, zero interval_value must keep the previous value.
	_, err = s.UpdateTaskFromConfig("sctest_update", map[string]interface{}{
		"schedule_time":  "09:10",
		"interval_value": 0,
	})
	if err != nil {
		t.Fatalf("flat update: %v", err)
	}
	task = s.tasks["sctest_update"]
	if task.ScheduleTime != "09:10" {
		t.Errorf("flat schedule_time = %q", task.ScheduleTime)
	}
	if task.IntervalValue != 30 {
		t.Errorf("zero interval_value should be ignored, got %d", task.IntervalValue)
	}
	// Enabled is untouched when no enabled key is present.
	if !task.Enabled {
		t.Error("enabled flag lost on partial update")
	}

	// params explicitly set to null clears them.
	if _, err := s.UpdateTaskFromConfig("sctest_update", map[string]interface{}{"params": nil}); err != nil {
		t.Fatalf("null params update: %v", err)
	}
	if s.tasks["sctest_update"].Params != "" {
		t.Errorf("null params should clear, got %q", s.tasks["sctest_update"].Params)
	}
}

func TestDeleteTask(t *testing.T) {
	s := newSCScheduler()
	parent := database.MainTask{TaskID: "sctest_del_p", Name: "P", Endpoint: "/api/p", ScheduleType: "daily", Enabled: 1, TaskType: "main"}
	child := database.MainTask{TaskID: "sctest_del_c", Name: "C", Endpoint: "/api/c", ScheduleType: "daily", Enabled: 1, TaskType: "sub", ParentID: "sctest_del_p"}
	for _, mt := range []database.MainTask{parent, child} {
		if err := database.UpsertMainTask(mt); err != nil {
			t.Fatalf("seed %s: %v", mt.TaskID, err)
		}
		s.tasks[mt.TaskID] = convertTask(mt)
	}

	if err := s.DeleteTask("sctest_del_missing"); err == nil {
		t.Error("deleting missing task must error")
	}
	if err := s.DeleteTask("sctest_del_p"); err != nil {
		t.Fatalf("DeleteTask: %v", err)
	}
	if _, ok := s.tasks["sctest_del_p"]; ok {
		t.Error("parent still in memory")
	}
	if _, ok := s.tasks["sctest_del_c"]; ok {
		t.Error("sub-task should be removed from memory with parent")
	}
	if _, ok := scMainTask(t, "sctest_del_p"); ok {
		t.Error("parent still in DB")
	}
	if _, ok := scMainTask(t, "sctest_del_c"); ok {
		t.Error("sub-task still in DB")
	}
}

func TestSetTaskEnabled(t *testing.T) {
	s := newSCScheduler()
	mt := database.MainTask{TaskID: "sctest_toggle", Name: "T", Endpoint: "/api/t", ScheduleType: "daily", Enabled: 0, TaskType: "main"}
	if err := database.UpsertMainTask(mt); err != nil {
		t.Fatalf("seed: %v", err)
	}
	s.tasks["sctest_toggle"] = convertTask(mt)

	if err := s.SetTaskEnabled("sctest_toggle", true); err != nil {
		t.Fatalf("SetTaskEnabled: %v", err)
	}
	if !s.tasks["sctest_toggle"].Enabled {
		t.Error("in-memory enabled not set")
	}
	if db, ok := scMainTask(t, "sctest_toggle"); !ok || db.Enabled != 1 {
		t.Errorf("DB enabled = %v", db.Enabled)
	}
	// Unknown id: no memory change, DB update is a silent no-op.
	if err := s.SetTaskEnabled("sctest_toggle_missing", true); err != nil {
		t.Errorf("SetTaskEnabled(missing) = %v, want nil", err)
	}
}

func TestGetTaskExecutionsEmpty(t *testing.T) {
	s := newSCScheduler()
	// No rows for this namespaced id.
	if got := s.GetTaskExecutions("sctest_never_ran", 10); len(got) != 0 {
		t.Errorf("expected empty history, got %#v", got)
	}
	// Invalid limit values are normalized inside database.GetExecutionHistory.
	start := time.Now().Add(-time.Hour)
	end := start.Add(time.Minute)
	if err := database.RecordExecution("sctest_hist_1", "sctest_hist_task", "completed", "ok", "", start, end); err != nil {
		t.Fatalf("record: %v", err)
	}
	got := s.GetTaskExecutions("sctest_hist_task", 0)
	if len(got) != 1 || got[0]["status"] != "completed" {
		t.Errorf("history: %#v", got)
	}
}

// ---------------------------------------------------------------------------
// async task registry
// ---------------------------------------------------------------------------

func TestAsyncTasks(t *testing.T) {
	id1, start := StartAsyncTask("测试异步任务")
	if !strings.HasPrefix(id1, "async_") || start.IsZero() {
		t.Fatalf("StartAsyncTask = %q, %v", id1, start)
	}
	id2, _ := StartAsyncTask("失败任务")

	st := GetAsyncTaskStatus(id1)
	if st["status"] != "running" || st["name"] != "测试异步任务" || st["task_id"] != id1 {
		t.Errorf("running status: %#v", st)
	}
	if _, ok := st["end_time"]; ok {
		t.Error("running task must not expose end_time")
	}

	CompleteAsyncTask(id1, true, strings.Repeat("r", 700), "")
	st = GetAsyncTaskStatus(id1)
	if st["status"] != "completed" {
		t.Fatalf("completed status: %#v", st)
	}
	if _, ok := st["end_time"]; !ok {
		t.Error("completed task must expose end_time")
	}
	if d, ok := st["duration"].(float64); !ok || d < 0 {
		t.Errorf("duration = %#v", st["duration"])
	}
	if r, _ := st["result"].(string); len(r) != 700 {
		t.Errorf("in-memory result should keep full payload, len=%d", len(r))
	}
	// Regression: the completion must land in the execution history that
	// StartAsyncTask opened as "running" — RecordExecution would collide with
	// that row's PRIMARY KEY id, so CompleteAsyncTask closes it out instead.
	hist, _ := database.GetExecutionHistory(id1, 5)
	if len(hist) != 1 {
		t.Fatalf("async history = %#v, want one row", hist)
	}
	if hist[0]["status"] != "completed" {
		t.Errorf("async history status = %#v, want completed", hist[0]["status"])
	}
	if r, _ := hist[0]["result"].(string); len(r) != 500 {
		t.Errorf("persisted result should be truncated to 500, len=%d", len(r))
	}

	CompleteAsyncTask(id2, false, "", "boom")
	st = GetAsyncTaskStatus(id2)
	if st["status"] != "failed" || st["error"] != "boom" {
		t.Errorf("failed status: %#v", st)
	}
	hist2, _ := database.GetExecutionHistory(id2, 5)
	if len(hist2) != 1 || hist2[0]["status"] != "failed" {
		t.Errorf("failed task history = %#v, want single failed row", hist2)
	}

	// Completing an unknown id must be a silent no-op.
	CompleteAsyncTask("sctest_unknown_async", true, "x", "")

	// Unknown in-memory id falls back to the DB history.
	if err := database.RecordExecution("async_sctest_dbonly", "async_sctest_dbonly", "completed", "from-db", "", time.Now(), time.Now()); err != nil {
		t.Fatalf("record: %v", err)
	}
	st = GetAsyncTaskStatus("async_sctest_dbonly")
	if st == nil || st["status"] != "completed" {
		t.Errorf("DB fallback status: %#v", st)
	}

	// Totally unknown id -> nil.
	if st := GetAsyncTaskStatus("sctest_no_such_async"); st != nil {
		t.Errorf("unknown async id = %#v, want nil", st)
	}
}
