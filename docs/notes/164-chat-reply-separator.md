# Chat replies split at tool calls (#164)

2026-11-02. The gateway joined every `text` event of a turn into one
reply, so text written before a tool call ran into the text after it
("Now cancel and close.I cancelled run ..."). Text events are streaming
chunks of one message and stay joined as-is; a `tool` event between two
non-empty runs of text now inserts a blank line (`pkg/gateway/gateway.go`,
`follow`). The reply is still one post per turn.
