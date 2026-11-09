// Package cron parses standard 5-field cron expressions (minute hour
// day-of-month month day-of-week) and finds the next matching minute.
// Supported per field: *, numbers, a-b ranges, lists with commas, and /step
// on * or ranges. Day of week is 0-7 with 0 and 7 both Sunday. As in Vixie
// cron, when both day fields are restricted a day matches if either does.
// Names (JAN, MON) and macros (@daily) are not supported.
package cron

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Schedule is a parsed expression.
type Schedule struct {
	minute, hour, dom, month, dow uint64 // bit i set: value i matches
	domStar, dowStar              bool
}

type field struct {
	name     string
	min, max int
}

var fields = []field{
	{"minute", 0, 59},
	{"hour", 0, 23},
	{"day of month", 1, 31},
	{"month", 1, 12},
	{"day of week", 0, 7},
}

// Parse parses a 5-field expression.
func Parse(expr string) (*Schedule, error) {
	parts := strings.Fields(expr)
	if len(parts) != 5 {
		return nil, fmt.Errorf("cron %q: want 5 fields, got %d", expr, len(parts))
	}
	var bits [5]uint64
	for i, p := range parts {
		b, err := parseField(p, fields[i])
		if err != nil {
			return nil, fmt.Errorf("cron %q: %w", expr, err)
		}
		bits[i] = b
	}
	if bits[4]&(1<<7) != 0 { // 7 is Sunday too
		bits[4] |= 1
	}
	return &Schedule{
		minute: bits[0], hour: bits[1], dom: bits[2], month: bits[3], dow: bits[4],
		domStar: parts[2] == "*", dowStar: parts[4] == "*",
	}, nil
}

func parseField(s string, f field) (uint64, error) {
	var bits uint64
	for _, item := range strings.Split(s, ",") {
		rng, stepStr, hasStep := strings.Cut(item, "/")
		lo, hi := f.min, f.max
		switch {
		case rng == "*":
		case strings.Contains(rng, "-"):
			a, b, _ := strings.Cut(rng, "-")
			var err error
			if lo, err = num(a, f); err != nil {
				return 0, err
			}
			if hi, err = num(b, f); err != nil {
				return 0, err
			}
			if lo > hi {
				return 0, fmt.Errorf("%s: range %q is backwards", f.name, rng)
			}
		default:
			n, err := num(rng, f)
			if err != nil {
				return 0, err
			}
			lo, hi = n, n
			if hasStep { // n/step means n through max
				hi = f.max
			}
		}
		step := 1
		if hasStep {
			n, err := strconv.Atoi(stepStr)
			if err != nil || n <= 0 {
				return 0, fmt.Errorf("%s: bad step %q", f.name, stepStr)
			}
			step = n
		}
		for v := lo; v <= hi; v += step {
			bits |= 1 << v
		}
	}
	return bits, nil
}

func num(s string, f field) (int, error) {
	n, err := strconv.Atoi(s)
	if err != nil || n < f.min || n > f.max {
		return 0, fmt.Errorf("%s: %q is not in %d-%d", f.name, s, f.min, f.max)
	}
	return n, nil
}

// Next returns the first matching minute strictly after t, in t's location.
// It returns the zero time if nothing matches within five years (e.g.
// "0 0 30 2 *").
func (s *Schedule) Next(t time.Time) time.Time {
	loc := t.Location()
	t = t.Truncate(time.Minute).Add(time.Minute)
	end := t.AddDate(5, 0, 0)
	for t.Before(end) {
		if !has(s.month, int(t.Month())) {
			t = time.Date(t.Year(), t.Month()+1, 1, 0, 0, 0, 0, loc)
			continue
		}
		if !s.dayMatches(t) {
			t = time.Date(t.Year(), t.Month(), t.Day()+1, 0, 0, 0, 0, loc)
			continue
		}
		if !has(s.hour, t.Hour()) {
			t = time.Date(t.Year(), t.Month(), t.Day(), t.Hour()+1, 0, 0, 0, loc)
			continue
		}
		if !has(s.minute, t.Minute()) {
			t = t.Add(time.Minute)
			continue
		}
		return t
	}
	return time.Time{}
}

func (s *Schedule) dayMatches(t time.Time) bool {
	dom, dow := has(s.dom, t.Day()), has(s.dow, int(t.Weekday()))
	switch {
	case s.domStar && s.dowStar:
		return true
	case s.domStar:
		return dow
	case s.dowStar:
		return dom
	}
	return dom || dow
}

func has(bits uint64, v int) bool { return bits&(1<<v) != 0 }
