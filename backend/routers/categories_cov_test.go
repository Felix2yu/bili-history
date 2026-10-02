package routers

import (
	"net/url"
	"sort"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

// TestFTCategoryRoutes drives every handler installed by RegisterCategoryRoutes
// (categories.go:12) against the shared sqlite database. The category tree is
// written by database.InitCategories, so the assertions read back the built-in
// rows rather than fixtures of this file.
func TestFTCategoryRoutes(t *testing.T) {
	e := newAPI(t, RegisterCategoryRoutes)

	// POST /api/categories/init -> initCategories (categories.go:22).
	initResp := expectStatus(t, e, "POST", "/api/categories/init", "", 200, "success")
	expectMessage(t, initResp, "视频分类表初始化成功")

	// The insert loop is documented as idempotent ("INSERT OR IGNORE"), but the
	// table declares no unique constraint over (main_category, sub_category), so
	// IGNORE can never trigger: every init appends a full copy of the tree.
	// PRODUCTION BUG: database/categories.go:194 re-seeds duplicates instead of
	// ignoring them; a UNIQUE(main_category, sub_category) index (or an upsert)
	// would make the second init a no-op.
	countBefore := countRows(t, "SELECT COUNT(*) FROM video_categories")
	second := expectStatus(t, e, "POST", "/api/categories/init", "", 200, "success")
	expectMessage(t, second, "视频分类表初始化成功")
	afterSecond := countRows(t, "SELECT COUNT(*) FROM video_categories")
	if delta := afterSecond - countBefore; delta < 100 {
		t.Fatalf("second init inserted %d rows, want a full copy of the built-in tree (>100)", delta)
	}
	if countBefore < 100 {
		t.Fatalf("video_categories holds %d rows after the first init, want the built-in tree (>100)", countBefore)
	}

	// GET /api/categories/categories -> getCategories (categories.go:35).
	cats := expectStatus(t, e, "GET", "/api/categories/categories", "", 200, "success").dataArray(t)
	if len(cats) < 20 {
		t.Fatalf("categories returned %d nodes, want the full main-category tree", len(cats))
	}
	tech := ftCategoryNode(t, cats, "科技")
	subs := dataArray(t, tech["sub_categories"])
	if len(subs) == 0 {
		t.Fatalf("科技 has no sub categories: %v", tech)
	}
	if got := ftSubCategory(t, subs, "数码"); got["tid"] != float64(95) || got["alias"] != "数码" {
		t.Fatalf("数码 sub category = %v, want tid 95 and alias 数码", got)
	}
	// The node image column is empty for every built-in row.
	if tech["image"] != "" && tech["image"] != nil {
		t.Fatalf("科技 image = %v, want the empty string", tech["image"])
	}
	// GetCategories drops sub categories whose name equals the main category, so
	// the self-titled "VLOG" partition ends up with an empty child list.
	vlog := ftCategoryNode(t, cats, "VLOG")
	if kids := dataArray(t, vlog["sub_categories"]); len(kids) != 0 {
		t.Fatalf("VLOG sub_categories = %v, want the self reference filtered out", kids)
	}

	// GET /api/categories/main-categories -> getMainCategories (categories.go:45).
	mains := expectStatus(t, e, "GET", "/api/categories/main-categories", "", 200, "success").dataArray(t)
	if len(mains) != len(cats) {
		t.Fatalf("main-categories = %d entries, want one per category node (%d)", len(mains), len(cats))
	}
	names := make([]string, 0, len(mains))
	for _, m := range mains {
		entry, ok := m.(map[string]interface{})
		if !ok {
			t.Fatalf("main-categories entry = %T, want an object", m)
		}
		name, _ := entry["name"].(string)
		if name == "" {
			t.Fatalf("main-categories entry without a name: %v", entry)
		}
		if _, hasImage := entry["image"]; !hasImage {
			t.Fatalf("main-categories entry %q lost the image key: %v", name, entry)
		}
		names = append(names, name)
	}
	sorted := append([]string(nil), names...)
	sort.Strings(sorted)
	for i := range names {
		if names[i] != sorted[i] {
			t.Fatalf("main-categories are not ordered by name (SQL ORDER BY main_category): %v", names)
		}
	}
	for _, want := range []string{"动画", "科技", "鬼畜"} {
		if !contains(names, want) {
			t.Fatalf("main-categories %v miss %q", names, want)
		}
	}

	// GET /api/categories/sub-categories/:main_category -> getSubCategories
	// (categories.go:55). The rows are ordered by sub_category and deduplicated.
	games := ftSubCategoryList(t, e, "游戏")
	if len(games) < 6 {
		t.Fatalf("游戏 has %d sub categories, want the built-in list", len(games))
	}
	gameNames := make([]string, 0, len(games))
	for _, s := range games {
		gameNames = append(gameNames, s["name"].(string))
	}
	sortedGames := append([]string(nil), gameNames...)
	sort.Strings(sortedGames)
	for i := range gameNames {
		if gameNames[i] != sortedGames[i] {
			t.Fatalf("sub-categories not ordered by name: %v", gameNames)
		}
	}
	if !contains(gameNames, "单机游戏") || !contains(gameNames, "电子竞技") {
		t.Fatalf("sub-categories of 游戏 = %v, want the built-in entries", gameNames)
	}
	if got := ftSubCategory(t, toAnySlice(games), "单机游戏"); got["tid"] != float64(17) {
		t.Fatalf("单机游戏 = %v, want tid 17", got)
	}

	// A main category that exists only as a sub category yields no children. The
	// handler still answers "success", and the empty slice is omitted from the
	// envelope by models.SuccessResponse's omitempty on Data.
	missing := expectStatus(t, e, "GET", "/api/categories/sub-categories/"+url.PathEscape("不存在的分区FT"), "", 200, "success")
	if got := ftRawData(missing); got != "" && got != "null" {
		t.Fatalf("unknown category data = %q, want it omitted or null", got)
	}

	// The self-titled row is skipped here too (sub_category != main_category).
	vlogSubs := expectStatus(t, e, "GET", "/api/categories/sub-categories/"+url.PathEscape("VLOG"), "", 200, "success")
	if got := ftRawData(vlogSubs); got != "" && got != "null" {
		t.Fatalf("VLOG sub-categories = %q, want the self reference filtered out", got)
	}

	// An empty path parameter cannot be routed at all (gin 404s with a plain-text
	// body before the handler runs), which is what keeps getSubCategories free of
	// a validation branch.
	if w := doRaw(t, e, "GET", "/api/categories/sub-categories/", ""); w.Code != 404 {
		t.Fatalf("empty main_category code = %d, want 404", w.Code)
	}
}

// ftRawData reports the raw JSON of the envelope's data field, if any.
func ftRawData(resp apiResp) string {
	return strings.TrimSpace(string(resp.Data))
}

func ftCategoryNode(t *testing.T, cats []interface{}, name string) map[string]interface{} {
	t.Helper()
	for _, c := range cats {
		node, ok := c.(map[string]interface{})
		if !ok {
			continue
		}
		if node["name"] == name {
			return node
		}
	}
	t.Fatalf("category node %q missing from %d nodes", name, len(cats))
	return nil
}

func ftSubCategory(t *testing.T, subs []interface{}, name string) map[string]interface{} {
	t.Helper()
	for _, s := range subs {
		entry, ok := s.(map[string]interface{})
		if !ok {
			continue
		}
		if name == "" || entry["name"] == name {
			return entry
		}
	}
	t.Fatalf("sub category %q missing from %v", name, subs)
	return nil
}

func ftSubCategoryList(t *testing.T, e *gin.Engine, name string) []map[string]interface{} {
	t.Helper()
	raw := expectStatus(t, e, "GET", "/api/categories/sub-categories/"+url.PathEscape(name), "", 200, "success").dataArray(t)
	out := make([]map[string]interface{}, 0, len(raw))
	for _, s := range raw {
		entry, ok := s.(map[string]interface{})
		if !ok {
			t.Fatalf("sub-category entry = %T, want an object", s)
		}
		out = append(out, entry)
	}
	return out
}

func toAnySlice(list []map[string]interface{}) []interface{} {
	out := make([]interface{}, 0, len(list))
	for _, m := range list {
		out = append(out, m)
	}
	return out
}
