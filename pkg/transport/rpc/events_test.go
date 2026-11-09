package rpc_test

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"image"
	"image/png"
	"testing"

	"github.com/elee1766/jobshop/pkg/transport/rpc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTranscriptImages(t *testing.T) {
	var buf bytes.Buffer
	require.NoError(t, png.Encode(&buf, image.NewRGBA(image.Rect(0, 0, 12, 8))))
	data := base64.StdEncoding.EncodeToString(buf.Bytes())
	block := map[string]any{"type": "image", "mimeType": "image/png", "data": data}
	for _, tc := range []struct {
		name, method, kind string
		params             any
	}{
		{"sent", "session/prompt", rpc.EventTurnStart, map[string]any{"prompt": []any{block}}},
		{"received", "session/update", rpc.EventImage, map[string]any{"update": map[string]any{"sessionUpdate": "agent_message_chunk", "content": block}}},
		{"tool", "session/update", rpc.EventTool, map[string]any{"update": map[string]any{"sessionUpdate": "tool_call_update", "toolCallId": "read-image", "status": "completed", "content": []any{map[string]any{"type": "content", "content": block}}}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			raw, err := json.Marshal(map[string]any{"method": tc.method, "params": tc.params})
			require.NoError(t, err)
			var dec rpc.EventDecoder
			event, ok := dec.Decode(7, 0, "agent", raw)
			require.True(t, ok)
			assert.Equal(t, tc.kind, event.Type)
			require.Len(t, event.Images, 1)
			assert.Equal(t, rpc.SessionImage{MimeType: "image/png", Data: data}, event.Images[0])
		})
	}
	for _, tc := range []struct{ name, mime, data string }{
		{"svg", "image/svg+xml", base64.StdEncoding.EncodeToString([]byte(`<svg onload="alert(1)"/>`))},
		{"mismatched mime", "image/jpeg", data},
		{"invalid base64", "image/png", "not base64"},
		{"too large", "image/png", base64.StdEncoding.EncodeToString(make([]byte, rpc.MaxAttachmentBytes+1))},
	} {
		t.Run(tc.name, func(t *testing.T) {
			raw, err := json.Marshal(map[string]any{"method": "session/update", "params": map[string]any{"update": map[string]any{"sessionUpdate": "agent_message_chunk", "content": map[string]any{"type": "image", "mimeType": tc.mime, "data": tc.data}}}})
			require.NoError(t, err)
			var dec rpc.EventDecoder
			event, ok := dec.Decode(8, 0, "agent", raw)
			assert.False(t, ok)
			assert.Empty(t, event.Images)
		})
	}
}
