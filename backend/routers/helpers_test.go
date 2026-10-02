package routers

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseURLCookies(t *testing.T) {
	tests := []struct {
		name string
		url  string
		want map[string]string
	}{
		{
			name: "all four cookies",
			url:  "https://www.bilibili.com/?DedeUserID=123&DedeUserID__ckMd5=abc&SESSDATA=xyz%2C1%2C&bili_jct=token",
			want: map[string]string{"DedeUserID": "123", "DedeUserID__ckMd5": "abc", "SESSDATA": "xyz%2C1%2C", "bili_jct": "token"},
		},
		{
			name: "unknown keys are ignored",
			url:  "https://www.bilibili.com/?foo=bar&SESSDATA=s",
			want: map[string]string{"SESSDATA": "s"},
		},
		{
			name: "no query string at all",
			url:  "https://www.bilibili.com/video/BV1",
			want: map[string]string{},
		},
		{
			name: "parameter without equals sign is skipped",
			url:  "https://x/?flag&SESSDATA=s",
			want: map[string]string{"SESSDATA": "s"},
		},
		{
			name: "empty value is kept verbatim",
			url:  "https://x/?SESSDATA=",
			want: map[string]string{"SESSDATA": ""},
		},
		{
			name: "value containing equals is not split further",
			url:  "https://x/?SESSDATA=a=b=c",
			want: map[string]string{"SESSDATA": "a=b=c"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := map[string]string{}
			parseURLCookies(tc.url, got)
			if len(got) != len(tc.want) {
				t.Fatalf("parseURLCookies(%q) = %v, want %v", tc.url, got, tc.want)
			}
			for k, v := range tc.want {
				if got[k] != v {
					t.Fatalf("parseURLCookies(%q)[%q] = %q, want %q", tc.url, k, got[k], v)
				}
			}
		})
	}
}

func TestParseURLCookiesNilMapPanicsGuard(t *testing.T) {
	// Documented behaviour: the helper writes straight into the caller's map, so
	// a nil map is a programming error. The production callers always pass a
	// non-nil map; assert the happy path only and keep the note for readers.
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic when writing to a nil map")
		}
	}()
	var m map[string]string
	parseURLCookies("https://x/?SESSDATA=s", m)
}

func TestExtractOidFromKid(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"archive_123", "123"},
		{"live_21483240", "21483240"},
		{"article_123456", "123456"},
		{"archive_12_extra", "12_extra"}, // only the first separator is consumed
		{"noseparator", ""},
		{"", ""},
		{"prefix_", ""},
	}
	for _, tc := range tests {
		if got := extractOidFromKid(tc.in); got != tc.want {
			t.Fatalf("extractOidFromKid(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestTitleClassificationHelpers(t *testing.T) {
	cases := []struct {
		in             string
		digit, alpha   bool
		allAlpha, allD bool
	}{
		{"1024", true, false, false, true},
		{"Hello World", false, true, true, false},
		{"你好世界", false, false, false, false},
		{"BV1xx411c7mu", true, true, false, false},
		{"", false, false, false, false},
		{"abc1", true, true, false, false},
		// A single space is neither empty nor contains a disallowed rune, so
		// isAllAlpha reports true; the "pure letter" bucket therefore also
		// captures whitespace-only titles.
		{" ", false, false, true, false},
	}
	for _, c := range cases {
		if got := containsDigit(c.in); got != c.digit {
			t.Fatalf("containsDigit(%q) = %v, want %v", c.in, got, c.digit)
		}
		if got := containsAlpha(c.in); got != c.alpha {
			t.Fatalf("containsAlpha(%q) = %v, want %v", c.in, got, c.alpha)
		}
		if got := isAllAlpha(c.in); got != c.allAlpha {
			t.Fatalf("isAllAlpha(%q) = %v, want %v", c.in, got, c.allAlpha)
		}
		if got := isAllDigit(c.in); got != c.allD {
			t.Fatalf("isAllDigit(%q) = %v, want %v", c.in, got, c.allD)
		}
	}
}

func TestFindImageByHash(t *testing.T) {
	dir := t.TempDir()
	// The helper stats "<dir>/<hash><ext>" for each known extension and returns
	// the first hit, so the extension ordering (.jpg before .png) decides ties.
	if err := os.WriteFile(filepath.Join(dir, "abcdef.png"), []byte("png"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "fedcba.jpg"), []byte("jpg"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "abcdef_extra.txt"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	if got := findImageByHash(dir, "abcdef"); got != filepath.Join(dir, "abcdef.png") {
		t.Fatalf("findImageByHash = %q, want %q", got, filepath.Join(dir, "abcdef.png"))
	}
	if got := findImageByHash(dir, "fedcba"); got != filepath.Join(dir, "fedcba.jpg") {
		t.Fatalf("findImageByHash = %q, want %q", got, filepath.Join(dir, "fedcba.jpg"))
	}
	// Only the exact file name matches; a longer name with the same prefix does not.
	if got := findImageByHash(dir, "abcdef_extra"); got != "" {
		t.Fatalf("findImageByHash for non-image name = %q, want empty", got)
	}
	if got := findImageByHash(dir, "zzzzzz"); got != "" {
		t.Fatalf("findImageByHash for missing hash = %q, want empty", got)
	}
	if got := findImageByHash(dir, "a"); got != "" {
		t.Fatalf("findImageByHash for short hash = %q, want empty", got)
	}
	// A non-existent directory must not panic.
	if got := findImageByHash(filepath.Join(dir, "nope"), "abcdef"); got != "" {
		t.Fatalf("findImageByHash on missing dir = %q, want empty", got)
	}

	// .jpg wins over .png when both exist, matching exts ordering.
	if err := os.WriteFile(filepath.Join(dir, "abcdef.jpg"), []byte("jpg"), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := findImageByHash(dir, "abcdef"); got != filepath.Join(dir, "abcdef.jpg") {
		t.Fatalf("extension priority = %q, want the .jpg variant", got)
	}
}

func TestCountLocalImages(t *testing.T) {
	dir := t.TempDir()
	sub := filepath.Join(dir, "2016")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"a.jpg", "b.png", "c.webp", "d.gif", "e.jpeg", "ignore.txt", "f.JPG"} {
		if err := os.WriteFile(filepath.Join(sub, name), []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	total, downloaded := countLocalImages(dir)
	// The extension is lower-cased before matching, so "f.JPG" is counted too;
	// "ignore.txt" is the only skipped file.
	if downloaded != 6 {
		t.Fatalf("downloaded = %d, want 6 (jpg/jpeg/png/webp/gif under %s)", downloaded, sub)
	}
	if total != downloaded {
		t.Fatalf("total = %d, want %d (no pending entries expected)", total, downloaded)
	}

	// A missing directory is reported as zero rather than an error.
	total, downloaded = countLocalImages(filepath.Join(dir, "missing"))
	if total != 0 || downloaded != 0 {
		t.Fatalf("missing dir -> (%d, %d), want (0, 0)", total, downloaded)
	}
}

func TestCountAllDynamicImages(t *testing.T) {
	dir := t.TempDir()
	// Dynamic images are stored per host as <dynamicDir>/<host>/images/*.
	for _, host := range []string{"100001", "100002"} {
		inner := filepath.Join(dir, host, "images")
		if err := os.MkdirAll(inner, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(inner, "x.png"), []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	// A stray file at the top level is not a host directory and is ignored.
	if err := os.WriteFile(filepath.Join(dir, "stray.png"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	total, downloaded := countAllDynamicImages(dir)
	if total != 2 || downloaded != 2 {
		t.Fatalf("countAllDynamicImages = (%d, %d), want (2, 2)", total, downloaded)
	}

	total, downloaded = countAllDynamicImages(filepath.Join(dir, "missing"))
	if total != 0 || downloaded != 0 {
		t.Fatalf("missing dir -> (%d, %d), want (0, 0)", total, downloaded)
	}
}
