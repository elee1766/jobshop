package cli

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
)

const maxStepOutputBytes = 1 << 20

func readStepOutput(path string) (map[string]any, error) {
	if path == "" {
		return nil, nil
	}
	reader := stdin
	if path != "-" {
		file, err := os.Open(path)
		if err != nil {
			return nil, fmt.Errorf("step output: %w", err)
		}
		defer func() { _ = file.Close() }()
		reader = file
	}
	body, err := io.ReadAll(io.LimitReader(reader, maxStepOutputBytes+1))
	if err != nil {
		return nil, fmt.Errorf("step output: %w", err)
	}
	if len(body) > maxStepOutputBytes {
		return nil, errors.New("step output exceeds 1 MiB")
	}
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.UseNumber()
	var output map[string]any
	if err := decoder.Decode(&output); err != nil || output == nil {
		return nil, errors.New("step output must be a JSON object")
	}
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		return nil, errors.New("step output must contain exactly one JSON object")
	}
	return output, nil
}
