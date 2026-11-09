package cli

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/support/logctx"
)

func TestLogFlags(t *testing.T) {
	var buf bytes.Buffer
	log, err := LogFlags{LogLevel: "debug", LogFormat: "json"}.loggerTo(&buf)
	require.NoError(t, err)
	log.DebugContext(logctx.With(t.Context(), "run", "r1"), "hi")
	var rec map[string]any
	require.NoError(t, json.Unmarshal(buf.Bytes(), &rec))
	assert.Equal(t, "DEBUG", rec["level"])
	assert.Equal(t, "r1", rec["run"])

	buf.Reset()
	log, err = LogFlags{LogLevel: "warn", LogFormat: "text"}.loggerTo(&buf)
	require.NoError(t, err)
	log.InfoContext(t.Context(), "dropped")
	assert.Empty(t, buf.String())

	_, err = LogFlags{LogLevel: "loud"}.loggerTo(&buf)
	require.Error(t, err)
	_, err = LogFlags{LogLevel: "info", LogFormat: "xml"}.loggerTo(&buf)
	require.Error(t, err)
}
