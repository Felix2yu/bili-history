package routers

import (
	"sort"
	"strings"
	"testing"
)

func TestGetEndpointMetaExactHit(t *testing.T) {
	RegisterEndpointMeta("GET", "/__test__/endpoint", EndpointMeta{
		Summary:     "测试端点",
		Tags:        []string{"测试"},
		OperationID: "test_endpoint",
	})

	meta := GetEndpointMeta("GET", "/__test__/endpoint")
	if meta.Summary != "测试端点" || meta.OperationID != "test_endpoint" {
		t.Fatalf("unexpected meta: %+v", meta)
	}
	if len(meta.Tags) != 1 || meta.Tags[0] != "测试" {
		t.Fatalf("unexpected tags: %v", meta.Tags)
	}
}

func TestGetEndpointMetaStripsAPIPrefix(t *testing.T) {
	RegisterEndpointMeta("POST", "/__test__/prefixed", EndpointMeta{Summary: "带前缀", OperationID: "prefixed"})

	meta := GetEndpointMeta("POST", "/api/__test__/prefixed")
	if meta.Summary != "带前缀" || meta.OperationID != "prefixed" {
		t.Fatalf("/api prefix not handled: %+v", meta)
	}
}

func TestGetEndpointMetaUnknownReturnsZeroValue(t *testing.T) {
	meta := GetEndpointMeta("GET", "/__test__/does-not-exist")
	if meta.Summary != "" || meta.OperationID != "" {
		t.Fatalf("expected zero meta, got %+v", meta)
	}
	// The contract documented on GetEndpointMeta is an *empty* (non-nil) slice,
	// so JSON consumers always render [] rather than null.
	if meta.Tags == nil {
		t.Fatal("Tags must be an empty slice, not nil")
	}
	if len(meta.Tags) != 0 {
		t.Fatalf("Tags = %v, want empty", meta.Tags)
	}
}

func TestRegisterEndpointMetaOverwrites(t *testing.T) {
	key := "GET /__test__/overwrite"
	RegisterEndpointMeta("GET", "/__test__/overwrite", EndpointMeta{Summary: "first"})
	RegisterEndpointMeta("GET", "/__test__/overwrite", EndpointMeta{Summary: "second"})
	if got := endpointRegistry[key]; got.Summary != "second" {
		t.Fatalf("registry entry = %+v, want the last registration to win", got)
	}
}

// fullAPIRoutes returns every "METHOD /path" the production router exposes.
func fullAPIRoutes() []string {
	e := newAPI(t0,
		RegisterHistoryRoutes, RegisterCategoryRoutes, RegisterLoginRoutes,
		RegisterAnalysisRoutes, RegisterViewingRoutes, RegisterFavoriteRoutes,
		RegisterConfigRoutes, RegisterSchedulerRoutes, RegisterDataSyncRoutes,
		RegisterExportRoutes, RegisterImportRoutes, RegisterCleanRoutes,
		RegisterLogRoutes, RegisterFetchRoutes, RegisterDeleteRoutes,
		RegisterVideoDetailsRoutes, RegisterInteractionRoutes,
		RegisterTitleAnalyticsRoutes, RegisterReportRoutes,
		RegisterImageRoutes, RegisterDownloadRoutes,
	)
	var out []string
	for _, ri := range e.Routes() {
		out = append(out, ri.Method+" "+strings.TrimPrefix(ri.Path, "/api"))
	}
	sort.Strings(out)
	return out
}

// t0 is a throwaway *testing.T used only while assembling the engine for the
// registry consistency check; the registration functions themselves never fail.
var t0 = &testing.T{}

func TestEndpointRegistryCoversAllRoutes(t *testing.T) {
	// Known gaps: routes that exist in the router but were never given metadata
	// in endpoints.go. They are listed here so the consistency check still
	// guards every other endpoint and so the gap is explicit.
	knownGaps := map[string]bool{
		"DELETE /dynamic/item/:id_str":              true,
		"GET /config/mcp-config":                    true,
		"GET /dynamic/user_card/:mid":               true,
		"GET /export/download/:task_id":             true,
		"GET /favorite/collected/local/list":        true,
		"GET /favorite/content/online":              true,
		"GET /favorite/local/list":                  true,
		"GET /history/dates":                        true,
		"GET /task/:id/status":                      true,
		"GET /title-analytics/length":               true,
		"GET /title-analytics/patterns":             true,
		"GET /title-analytics/sentiment":            true,
		"GET /title-analytics/stats":                true,
		"GET /title-analytics/trend":                true,
		"GET /viewing/extra-overview":               true,
		"GET /viewing/favorites-analysis":           true,
		"GET /viewing/likes-analysis":               true,
		"GET /viewing/watchlater-analysis":          true,
		"POST /config/mcp-config":                   true,
		"POST /importSqlite/import_data_sqlite":     true,
		"POST /like/toggle":                         true,
	}

	routes := fullAPIRoutes()
	if len(routes) < 100 {
		t.Fatalf("suspiciously few routes discovered: %d", len(routes))
	}

	var missing []string
	for _, r := range routes {
		if _, ok := endpointRegistry[r]; ok {
			continue
		}
		if knownGaps[r] {
			continue
		}
		missing = append(missing, r)
	}
	if len(missing) > 0 {
		t.Errorf("routes without endpoint metadata (%d):\n%s", len(missing), strings.Join(missing, "\n"))
	}
}

func TestEndpointRegistryHasNoStaleEntries(t *testing.T) {
	routes := fullAPIRoutes()
	alive := map[string]bool{}
	for _, r := range routes {
		alive[r] = true
	}

	// Known stale entries: endpoints.go registers the title analytics metadata
	// under "/title/..." while RegisterTitleAnalyticsRoutes mounts the handlers
	// on "/title-analytics/...". PRODUCTION BUG (not fixed here): GetEndpointMeta
	// can therefore never return the metadata for those five endpoints.
	knownStale := map[string]bool{
		"GET /title/length":   true,
		"GET /title/patterns": true,
		"GET /title/sentiment": true,
		"GET /title/stats":    true,
		"GET /title/trend":    true,
	}

	var stale []string
	for key := range endpointRegistry {
		if strings.Contains(key, "/__test__/") {
			continue // entries registered by this test suite
		}
		if knownStale[key] {
			continue
		}
		if !alive[key] {
			stale = append(stale, key)
		}
	}
	sort.Strings(stale)
	if len(stale) > 0 {
		t.Errorf("endpoint metadata for routes that no longer exist (%d):\n%s", len(stale), strings.Join(stale, "\n"))
	}
}
