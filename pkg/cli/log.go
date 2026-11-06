package cli

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"

	"github.com/elee1766/jobshop/pkg/logctx"
)

// LogFlags set the level and format of the log handler shared by serve,
// desk, ingest and gateway (JOB-49).
type LogFlags struct {
	LogLevel  string `help:"Log level: debug, info, warn or error." env:"JOBSHOP_LOG_LEVEL" default:"info"`
	LogFormat string `help:"Log format: text or json." env:"JOBSHOP_LOG_FORMAT" default:"text" enum:"text,json"`
}

// logger builds the process logger on stderr.
func (f LogFlags) logger() (*slog.Logger, error) { return f.loggerTo(os.Stderr) }

func (f LogFlags) loggerTo(w io.Writer) (*slog.Logger, error) {
	var lvl slog.Level
	if err := lvl.UnmarshalText([]byte(strings.TrimSpace(f.LogLevel))); err != nil {
		return nil, fmt.Errorf("--log-level %q: want debug, info, warn or error", f.LogLevel)
	}
	opts := &slog.HandlerOptions{Level: lvl}
	var h slog.Handler
	switch f.LogFormat {
	case "", "text":
		h = slog.NewTextHandler(w, opts)
	case "json":
		h = slog.NewJSONHandler(w, opts)
	default:
		return nil, fmt.Errorf("--log-format %q: want text or json", f.LogFormat)
	}
	return slog.New(logctx.NewHandler(h)), nil
}
