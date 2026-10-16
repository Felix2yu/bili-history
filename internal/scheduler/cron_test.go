package scheduler

import (
	"strings"
	"testing"
	"time"
)

// at builds a deterministic local time for cron matching tests.
func at(y, mo, d, hh, mm int) time.Time {
	return time.Date(y, time.Month(mo), d, hh, mm, 0, 0, time.Local)
}

func TestCronMatchFieldCount(t *testing.T) {
	cases := []struct {
		name    string
		expr    string
		wantErr bool
	}{
		{"empty", "", true},
		{"one field", "*", true},
		{"four fields", "0 0 1 1", true},
		{"six fields", "0 0 1 1 * *", true},
		{"five fields ok", "0 0 1 1 *", false},
		{"surrounding spaces trimmed", "   0 0 1 1 *   ", false},
		{"tabs and multiple spaces", "0\t0  1\t\t1 *", false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := cronMatch(c.expr, at(2026, 3, 1, 0, 0))
			if (err != nil) != c.wantErr {
				t.Fatalf("cronMatch(%q) err = %v, wantErr = %v", c.expr, err, c.wantErr)
			}
			if c.wantErr && !strings.Contains(err.Error(), "5 个字段") {
				t.Errorf("error message = %q, want field-count error", err.Error())
			}
		})
	}
}

func TestCronMatchExact(t *testing.T) {
	// 2026-03-05 14:30 is a Thursday (weekday 4).
	tm := at(2026, 3, 5, 14, 30)
	cases := []struct {
		expr string
		want bool
	}{
		{"30 14 5 3 4", true},
		{"30 14 5 3 *", true},
		{"* * * * *", true},
		{"29 14 5 3 4", false},
		{"30 13 5 3 4", false},
		{"30 14 6 3 4", false},
		{"30 14 5 4 4", false},
		{"30 14 5 3 5", false},
	}
	for _, c := range cases {
		ok, err := cronMatch(c.expr, tm)
		if err != nil {
			t.Fatalf("cronMatch(%q): unexpected error %v", c.expr, err)
		}
		if ok != c.want {
			t.Errorf("cronMatch(%q) = %v, want %v", c.expr, ok, c.want)
		}
	}
}

func TestCronMatchRanges(t *testing.T) {
	tm := at(2026, 12, 24, 9, 45)
	cases := []struct {
		expr string
		want bool
	}{
		{"30-50 8-10 20-28 11-12 1-5", true},
		{"0-45 * * * *", true},   // inclusive bounds
		{"45-45 * * * *", true},  // degenerate range
		{"46-50 * * * *", false}, // value below range
		{"0-44 * * * *", false},  // value above range
		{"* * 25-12 * *", false}, // reversed range simply never matches (no error)
	}
	for _, c := range cases {
		ok, err := cronMatch(c.expr, tm)
		if err != nil {
			t.Fatalf("cronMatch(%q): unexpected error %v", c.expr, err)
		}
		if ok != c.want {
			t.Errorf("cronMatch(%q) = %v, want %v", c.expr, ok, c.want)
		}
	}
}

func TestCronMatchLists(t *testing.T) {
	tm := at(2026, 7, 4, 12, 15)
	cases := []struct {
		expr string
		want bool
	}{
		{"5,15,25 * * * *", true},
		{"1-10,15 * * * *", true}, // list of ranges
		{"5,25 * * * *", false},
		{"15, * * * *", true},  // empty part after trailing comma matches nothing itself
		{"16, * * * *", false}, // so "16," must not match minute 15
		{"1,,2 * * * *", false},
		{"15,,16 * * * *", true},
	}
	for _, c := range cases {
		ok, err := cronMatch(c.expr, tm)
		if err != nil {
			t.Fatalf("cronMatch(%q): unexpected error %v", c.expr, err)
		}
		if ok != c.want {
			t.Errorf("cronMatch(%q) = %v, want %v", c.expr, ok, c.want)
		}
	}
}

func TestCronMatchStarInsideList(t *testing.T) {
	tm := at(2026, 1, 1, 0, 7)
	ok, err := cronMatch("1,*,3 * * * *", tm)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Error("expected \"1,*,3\" to match any minute via the * entry")
	}
}

func TestCronMatchSteps(t *testing.T) {
	cases := []struct {
		name string
		expr string
		t    time.Time
		want bool
	}{
		{"star-slash matches zero offset", "*/15 * * * *", at(2026, 1, 1, 5, 30), true},
		{"star-slash no match", "*/15 * * * *", at(2026, 1, 1, 5, 31), false},
		{"stepped range", "0-20/5 * * * *", at(2026, 1, 1, 5, 15), true},
		{"stepped range out of hi", "0-10/5 * * * *", at(2026, 1, 1, 5, 15), false},
		// Note: day-of-month minimum is 1; */5 means 1,6,11,... (value-1)%5==0.
		{"stepped day field clamps to min", "0 0 */5 * *", at(2026, 1, 6, 0, 0), true},
		{"stepped day field clamps miss", "0 0 */5 * *", at(2026, 1, 5, 0, 0), false},
		// "n/m" form: start at n, step m up to field max.
		{"value-slash start", "10/7 * * * *", at(2026, 1, 1, 0, 24), true}, // 10,17,24,31,38,45,52
		{"value-slash miss", "10/7 * * * *", at(2026, 1, 1, 0, 25), false},
		{"value-slash at start", "10/7 * * * *", at(2026, 1, 1, 0, 10), true},
		{"stepped range below lo", "10-20/3 * * * *", at(2026, 1, 1, 0, 9), false},
		{"stepped hour month", "0 */6 * * *", at(2026, 1, 1, 18, 0), true},
		{"stepped hour miss", "0 */6 * * *", at(2026, 1, 1, 17, 0), false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ok, err := cronMatch(c.expr, c.t)
			if err != nil {
				t.Fatalf("cronMatch(%q): unexpected error %v", c.expr, err)
			}
			if ok != c.want {
				t.Errorf("cronMatch(%q, %v) = %v, want %v", c.expr, c.t, ok, c.want)
			}
		})
	}
}

func TestCronMatchWeekday(t *testing.T) {
	// 2026-01-04 is a Sunday; 2026-01-05 a Monday; 2026-01-07 a Wednesday;
	// 2026-01-09 a Friday.
	sun := at(2026, 1, 4, 0, 0)
	mon := at(2026, 1, 5, 0, 0)
	wed := at(2026, 1, 7, 0, 0)
	fri := at(2026, 1, 9, 0, 0)

	cases := []struct {
		expr string
		t    time.Time
		want bool
	}{
		{"* * * * 0", sun, true}, // cron 0 = Sunday == Go time.Sunday
		{"* * * * 1", mon, true},
		{"* * * * 7", sun, true},   // bare 7 normalized to Sunday
		{"* * * * 1-5", fri, true}, // workdays
		{"* * * * 1-5", sun, false},
	}
	for _, c := range cases {
		ok, err := cronMatch(c.expr, c.t)
		if err != nil {
			t.Fatalf("cronMatch(%q): unexpected error %v", c.expr, err)
		}
		if ok != c.want {
			t.Errorf("cronMatch(%q, %v) = %v, want %v", c.expr, c.t, ok, c.want)
		}
	}

	// Regression: "7" is cron's alternate spelling of Sunday and must work
	// inside ranges and lists too, not only as the bare field "7".
	for _, c := range []struct {
		expr string
		t    time.Time
		want bool
	}{
		{"* * * * 5-7", sun, true},   // Fri-Sun covers Sunday
		{"* * * * 5-7", fri, true},
		{"* * * * 5-7", wed, false},
		{"* * * * 1,7", sun, true},
		{"* * * * 0-6", sun, true},
		{"* * * * 0-7", sun, true}, // every day of week
		{"* * * * 0-7", wed, true},
		{"* * * * 0-3", fri, false}, // Sunday+7 must not leak into other fields
		{"* * * * 7", sun, true},
		{"* * * * 7", mon, false},
	} {
		ok, err := cronMatch(c.expr, c.t)
		if err != nil {
			t.Fatalf("cronMatch(%q): %v", c.expr, err)
		}
		if ok != c.want {
			t.Errorf("cronMatch(%q, %v) = %v, want %v", c.expr, c.t.Weekday(), ok, c.want)
		}
	}
}

func TestCronMatchErrors(t *testing.T) {
	exprs := []string{
		"abc * * * *",   // invalid minute value
		"*/0 * * * *",   // zero step
		"*/x * * * *",   // non-numeric step
		"*/-1 * * * *",  // negative step
		"*/ * * * *",    // empty step
		"5-x * * * *",   // bad range end
		"x-5 * * * *",   // bad range start
		"5-x/2 * * * *", // bad stepped range end
		"x-5/2 * * * *", // bad stepped range start
		"x/2 * * * *",   // bad step start (value-slash form)
		"* x * * *",     // invalid hour
		"* * x * *",     // invalid day
		"* * * x *",     // invalid month
		"* * * * x",     // invalid weekday
		"1,bad * * * *", // error propagates from list member
	}
	for _, e := range exprs {
		t.Run(e, func(t *testing.T) {
			matched, err := cronMatch(e, at(2026, 1, 1, 0, 0))
			if err == nil {
				t.Fatalf("cronMatch(%q): expected error, got matched=%v", e, matched)
			}
			if matched {
				t.Errorf("cronMatch(%q) matched with error %v", e, err)
			}
		})
	}
}

func TestMatchFieldStar(t *testing.T) {
	ok, err := matchField("*", 0, 59, 42)
	if err != nil || !ok {
		t.Errorf("matchField(\"*\") = %v, %v; want true, nil", ok, err)
	}
}

func TestMatchFieldErrorPropagation(t *testing.T) {
	// First list item doesn't match, second is invalid: error must surface even
	// though an earlier part parsed fine.
	ok, err := matchField("1,zz", 0, 59, 5)
	if err == nil || ok {
		t.Errorf("matchField(\"1,zz\") = %v, %v; want false, error", ok, err)
	}
	// All valid but none match.
	ok, err = matchField("1,2", 0, 59, 5)
	if err != nil || ok {
		t.Errorf("matchField(\"1,2\", 5) = %v, %v; want false, nil", ok, err)
	}
}

func TestMatchPartWhitespaceOnly(t *testing.T) {
	ok, err := matchPart("   ", 0, 59, 5)
	if err != nil || ok {
		t.Errorf("matchPart(blank) = %v, %v; want false, nil", ok, err)
	}
}

func TestMatchPartStepClamping(t *testing.T) {
	// "0-31/10" on the day field (min 1, max 31): lo clamped 0->1, values 1,11,21,31.
	for _, day := range []int{1, 11, 21, 31} {
		ok, err := matchPart("0-31/10", 1, 31, day)
		if err != nil {
			t.Fatalf("day %d: %v", day, err)
		}
		if !ok {
			t.Errorf("day %d should match 0-31/10 after clamp", day)
		}
	}
	ok, _ := matchPart("0-31/10", 1, 31, 2)
	if ok {
		t.Error("day 2 should not match 0-31/10")
	}
	// Range above max: hi clamped to max; value-slash form "25/3" runs 25,28,... up to 31.
	ok, err := matchPart("20-100/5", 1, 31, 30)
	if err != nil || !ok {
		t.Errorf("20-100/5 at 30 = %v, %v; want true, nil", ok, err)
	}
	// value beyond clamped hi never matches.
	ok, err = matchPart("20-25/1", 1, 31, 30)
	if err != nil || ok {
		t.Errorf("20-25/1 at 30 = %v, %v; want false, nil", ok, err)
	}
}

func TestMatchPartSingleValues(t *testing.T) {
	ok, err := matchPart("7", 0, 59, 7)
	if err != nil || !ok {
		t.Errorf("single 7 vs 7 = %v, %v; want true, nil", ok, err)
	}
	ok, err = matchPart("7", 0, 59, 8)
	if err != nil || ok {
		t.Errorf("single 7 vs 8 = %v, %v; want false, nil", ok, err)
	}
	if _, err := matchPart("7x", 0, 59, 7); err == nil {
		t.Error("expected error for invalid single value")
	}
}

func TestCronMatchSecondPrecisionIgnored(t *testing.T) {
	// Matching is purely on calendar fields; seconds are not considered.
	base := at(2026, 6, 6, 6, 6)
	for _, secs := range []int{0, 17, 59} {
		ok, err := cronMatch("6 6 6 6 *", base.Add(time.Duration(secs)*time.Second))
		if err != nil || !ok {
			t.Errorf("seconds=%d: cronMatch = %v, %v; want true, nil", secs, ok, err)
		}
	}
}
