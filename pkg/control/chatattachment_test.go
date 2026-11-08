package control_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"net/url"
	"os"
	"strings"
	"testing"

	acp "github.com/coder/acp-go-sdk"
	"github.com/elee1766/jobshop/pkg/control"
	"github.com/elee1766/jobshop/pkg/control/controltest"
	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestChatAttachments(t *testing.T) {
	t.Parallel()
	root := controltest.NewRoot(t, fmt.Sprintf(`default_project: js
projects:
  js: {}
agents:
  fake:
    command: %q
    args: [fake-agent]
pools:
  operator:
    project: js
    agent: fake
    type: chat
    dir: $ROOT
`, os.Args[0]))
	c := controltest.Start(t, root, control.Options{})
	var imageBytes bytes.Buffer
	require.NoError(t, png.Encode(&imageBytes, image.NewRGBA(image.Rect(0, 0, 8, 8))))
	document := strings.Repeat("document stays outside prompt\n", 1500) + "終わり"
	files := []rpc.AttachmentContent{{Name: "screenshot.png", MimeType: "image/png", Data: imageBytes.Bytes()}, {Name: "notes.txt", MimeType: "text/plain", Data: []byte(document)}}
	var chat rpc.SessionInfo
	c.Call(rpc.MethodChatManage, rpc.ChatManageParams{Action: "new", Pool: "operator", Name: "attachments", Attachments: files}, &chat)
	var journal rpc.SessionJournalResult
	var documentPath string
	c.PassUntil(func() bool {
		c.Call(rpc.MethodSessionJournal, rpc.SessionJournalParams{ID: chat.ID}, &journal)
		for _, frame := range journal.Frames {
			var request struct {
				Method string            `json:"method"`
				Params acp.PromptRequest `json:"params"`
			}
			if json.Unmarshal(frame.Msg, &request) != nil || request.Method != "session/prompt" {
				continue
			}
			require.Len(t, request.Params.Prompt, 3)
			assert.NotContains(t, request.Params.Prompt[0].Text.Text, document[:200])
			require.NotNil(t, request.Params.Prompt[1].Image)
			assert.Equal(t, "image/png", request.Params.Prompt[1].Image.MimeType)
			assert.NotEmpty(t, request.Params.Prompt[1].Image.Data)
			link := request.Params.Prompt[2].ResourceLink
			require.NotNil(t, link)
			assert.Equal(t, "notes.txt", link.Name)
			u, err := url.Parse(link.Uri)
			require.NoError(t, err)
			assert.Equal(t, "file", u.Scheme)
			documentPath = u.Path
			assert.True(t, strings.HasPrefix(documentPath, root.SessionDir(chat.ID)+string(os.PathSeparator)))
			data, err := os.ReadFile(documentPath)
			require.NoError(t, err)
			assert.Equal(t, document, string(data))
			return true
		}
		return false
	}, "attachments never reached ACP prompt")
	for _, invalid := range []rpc.AttachmentContent{
		{Name: "../escape.txt", MimeType: "text/plain", Data: []byte("text")},
		{Name: "wrong.png", MimeType: "image/png", Data: []byte("not an image")},
		{Name: "large.txt", MimeType: "text/plain", Data: bytes.Repeat([]byte("x"), (2<<20)+1)},
	} {
		require.ErrorIs(t, c.Admin.Call(t.Context(), rpc.MethodSessionSend, rpc.SessionSendParams{ID: chat.ID, Attachments: []rpc.AttachmentContent{invalid}}, nil), store.ErrInvalid)
	}
	c.Stop()
	controltest.Start(t, root, control.Options{})
	data, err := os.ReadFile(documentPath)
	require.NoError(t, err)
	assert.Equal(t, document, string(data), "linked documents survive controller restart")
}
