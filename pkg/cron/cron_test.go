package cron

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNext(t *testing.T) {
	at := func(s string) time.Time {
		v, err := time.Parse("2006-01-02 15:04", s)
		require.NoError(t, err)
		return v
	}
	tests := []struct{ expr, from, want string }{
		{"0 6 * * *", "2026-11-01 05:59", "2026-11-01 06:00"},
		{"0 6 * * *", "2026-11-01 06:00", "2026-11-02 06:00"},
		{"*/15 * * * *", "2026-11-01 06:01", "2026-11-01 06:15"},
		{"30 9 * * 1-5", "2026-11-01 10:00", "2026-11-02 09:30"}, // Sunday -> Monday
		{"0 0 * * 7", "2026-11-02 00:00", "2026-11-08 00:00"},    // 7 is Sunday
		{"0 0 1 * *", "2026-11-15 00:00", "2026-12-01 00:00"},
		{"0 0 1 1 *", "2026-11-15 00:00", "2027-01-01 00:00"},
		{"0 0 13 * 5", "2026-11-01 00:00", "2026-11-06 00:00"}, // dom or dow
		{"5,10 1-2 * * *", "2026-11-01 01:07", "2026-11-01 01:10"},
		{"0 12/6 * * *", "2026-11-01 13:00", "2026-11-01 18:00"},
	}
	for _, tc := range tests {
		s, err := Parse(tc.expr)
		require.NoError(t, err, tc.expr)
		assert.Equal(t, at(tc.want), s.Next(at(tc.from)), "%s from %s", tc.expr, tc.from)
	}
	s, err := Parse("0 0 30 2 *")
	require.NoError(t, err)
	assert.True(t, s.Next(at("2026-11-01 00:00")).IsZero(), "never matches")
}

func TestParseErrors(t *testing.T) {
	for _, expr := range []string{"", "* * * *", "60 * * * *", "* 24 * * *", "* * 0 * *", "* * * 13 *", "* * * * 8", "5-1 * * * *", "*/0 * * * *", "a * * * *", "@daily"} {
		_, err := Parse(expr)
		assert.Error(t, err, expr)
	}
}
