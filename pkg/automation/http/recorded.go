package http

import (
	"encoding/json"
	"errors"

	"github.com/elee1766/jobshop/pkg/automation/api"
	"go.starlark.net/starlark"
)

// Restore validates a recorded response and reconstructs its live .json method.
// Fields must come from bounded JSON decoded with api.DecodeJSON.
func Restore(fields map[string]any, options api.Options) (starlark.Value, error) {
	if !options.AllowHTTP {
		return nil, errors.New("HTTP capability is disabled by host")
	}
	invalid := errors.New("recorded capability result does not match its capability contract")
	status, ok := api.RecordedInt(fields["status_code"])
	if !ok || status < 100 || status > 999 {
		return nil, invalid
	}
	body, ok := fields["body"].(string)
	if !ok || len(body) > options.Limits.OutputBytes {
		return nil, invalid
	}
	if text, exists := fields["text"]; exists {
		if text != body {
			return nil, invalid
		}
	}
	headers, ok := fields["headers"].(map[string]any)
	if !ok {
		return nil, invalid
	}
	for _, value := range headers {
		if _, ok := value.(string); !ok {
			return nil, invalid
		}
	}
	metadata := map[string]any{"status_code": fields["status_code"], "headers": headers}
	if _, err := api.ToValue(metadata, options.Limits.DataBytes); err != nil {
		return nil, invalid
	}
	encodedMetadata, err := json.Marshal(metadata)
	if err != nil || len(encodedMetadata) > options.Limits.DataBytes {
		return nil, invalid
	}
	return responseValue(status, []byte(body), headers, options)
}
