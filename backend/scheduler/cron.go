package scheduler

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// cronMatch evaluates a standard 5-field cron expression ("minute hour day month weekday")
// against the given time. Each field supports:
//   - "*"        (any value)
//   - "n"        (exact value)
//   - "a-b"      (range, inclusive)
//   - "*/n"      (every n units from the field's minimum)
//   - "a,b,c"    (list of values/ranges)
//   - "a-b/n"    (stepped range)
//
// Weekday uses 0-6 where 0 and 7 both mean Sunday (cron convention).
// The match is purely on the calendar fields — second-level precision is not supported.
func cronMatch(expr string, t time.Time) (bool, error) {
	fields := strings.Fields(strings.TrimSpace(expr))
	if len(fields) != 5 {
		return false, fmt.Errorf("cron 表达式必须是 5 个字段: %q", expr)
	}

	minute, hour, day, month, weekday := fields[0], fields[1], fields[2], fields[3], fields[4]

	ok, err := matchField(minute, 0, 59, t.Minute())
	if err != nil || !ok {
		return false, err
	}
	ok, err = matchField(hour, 0, 23, t.Hour())
	if err != nil || !ok {
		return false, err
	}
	ok, err = matchField(day, 1, 31, t.Day())
	if err != nil || !ok {
		return false, err
	}
	ok, err = matchField(month, 1, 12, int(t.Month()))
	if err != nil || !ok {
		return false, err
	}
	// Go's Weekday: Sunday=0, Monday=1, ... Saturday=6 — same as cron.
	cronWD := int(t.Weekday())
	// Normalize 7 -> 0 for Sunday (some cron users write 7 for Sunday).
	wd := weekday
	if wd == "7" {
		wd = "0"
	}
	ok, err = matchField(wd, 0, 7, cronWD)
	if err != nil || !ok {
		return false, err
	}
	return true, nil
}

// matchField checks whether a single cron field matches the given value.
func matchField(field string, min, max, value int) (bool, error) {
	if field == "*" {
		return true, nil
	}
	// Handle comma-separated lists: "1,5,10" or "1-5,30"
	for _, part := range strings.Split(field, ",") {
		ok, err := matchPart(part, min, max, value)
		if err != nil {
			return false, err
		}
		if ok {
			return true, nil
		}
	}
	return false, nil
}

// matchPart evaluates one component of a cron field (which may be a range, step,
// or single value).
func matchPart(part string, min, max, value int) (bool, error) {
	part = strings.TrimSpace(part)
	if part == "" {
		return false, nil
	}
	if part == "*" {
		return true, nil
	}

	// Stepped form: "*/n" or "a-b/n"
	if idx := strings.Index(part, "/"); idx >= 0 {
		rangePart := part[:idx]
		stepStr := part[idx+1:]
		step, err := strconv.Atoi(stepStr)
		if err != nil || step <= 0 {
			return false, fmt.Errorf("无效的步长 %q", stepStr)
		}
		var lo, hi int
		if rangePart == "*" {
			lo, hi = min, max
		} else if dashIdx := strings.Index(rangePart, "-"); dashIdx >= 0 {
			lo, err = strconv.Atoi(rangePart[:dashIdx])
			if err != nil {
				return false, fmt.Errorf("无效的范围起始 %q", rangePart[:dashIdx])
			}
			hi, err = strconv.Atoi(rangePart[dashIdx+1:])
			if err != nil {
				return false, fmt.Errorf("无效的范围结束 %q", rangePart[dashIdx+1:])
			}
		} else {
			// "n/m" means starting at n, every m units up to max.
			n, err := strconv.Atoi(rangePart)
			if err != nil {
				return false, fmt.Errorf("无效的步长起始 %q", rangePart)
			}
			lo = n
			hi = max
		}
		if lo < min {
			lo = min
		}
		if hi > max {
			hi = max
		}
		if value < lo || value > hi {
			return false, nil
		}
		return (value-lo)%step == 0, nil
	}

	// Range form: "a-b"
	if dashIdx := strings.Index(part, "-"); dashIdx >= 0 {
		lo, err := strconv.Atoi(part[:dashIdx])
		if err != nil {
			return false, fmt.Errorf("无效的范围起始 %q", part[:dashIdx])
		}
		hi, err := strconv.Atoi(part[dashIdx+1:])
		if err != nil {
			return false, fmt.Errorf("无效的范围结束 %q", part[dashIdx+1:])
		}
		return value >= lo && value <= hi, nil
	}

	// Single value
	n, err := strconv.Atoi(part)
	if err != nil {
		return false, fmt.Errorf("无效的 cron 值 %q", part)
	}
	return value == n, nil
}
