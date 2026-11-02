package control

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestGroupLogs(t *testing.T) {
	t0 := time.Date(2026, 11, 2, 0, 0, 0, 0, time.UTC)
	recs := []logRecord{
		{At: t0, Level: slog.LevelWarn, Msg: "crash release", Attrs: map[string]string{"ticket": "a"}},
		{At: t0.Add(time.Minute), Level: slog.LevelError, Msg: "close session ticket"},
		{At: t0.Add(2 * time.Minute), Level: slog.LevelWarn, Msg: "crash release", Attrs: map[string]string{"ticket": "b"}},
		{At: t0.Add(3 * time.Minute), Level: slog.LevelError, Msg: "crash release"},
	}
	g := groupLogs(recs)
	require.Len(t, g, 3)
	require.Equal(t, "WARN", g[0].Level)
	require.Equal(t, "crash release", g[0].Message)
	require.Equal(t, 2, g[0].Count)
	require.Equal(t, t0, g[0].First)
	require.Equal(t, t0.Add(2*time.Minute), g[0].Last)
	require.Equal(t, "b", g[0].Sample["ticket"])
	// Ties order by most recent last.
	require.Equal(t, "ERROR", g[1].Level)
	require.Equal(t, "crash release", g[1].Message)
	require.Equal(t, "close session ticket", g[2].Message)
	require.Empty(t, groupLogs(nil))
}

func TestIncidentLogSince(t *testing.T) {
	il := &incidentLog{}
	log := slog.New(incidentHandler{Handler: slog.DiscardHandler, log: il}).With("pool", "dev")
	t0 := time.Date(2026, 11, 2, 0, 0, 0, 0, time.UTC)
	for i, lvl := range []slog.Level{slog.LevelInfo, slog.LevelWarn, slog.LevelError, slog.LevelWarn} {
		r := slog.NewRecord(t0.Add(time.Duration(i)*time.Hour), lvl, "m", 0)
		r.AddAttrs(slog.Int("i", i))
		require.NoError(t, log.Handler().Handle(context.Background(), r))
	}
	all := il.since(time.Time{})
	require.Len(t, all, 3, "info is not kept")
	require.Equal(t, "dev", all[0].Attrs["pool"])
	require.Equal(t, "1", all[0].Attrs["i"])
	got := il.since(t0.Add(2 * time.Hour))
	require.Len(t, got, 2, "boundary is inclusive")
	require.Equal(t, "2", got[0].Attrs["i"])
	require.Empty(t, il.since(t0.Add(4*time.Hour)))
}

func TestIncidentLogCap(t *testing.T) {
	il := &incidentLog{}
	t0 := time.Now()
	for i := range incidentCap + 10 {
		il.add(logRecord{At: t0.Add(time.Duration(i)), Msg: "m"})
	}
	all := il.since(time.Time{})
	require.Len(t, all, incidentCap)
	require.Equal(t, t0.Add(10), all[0].At)
}
