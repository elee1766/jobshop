package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"strconv"
	"time"

	"go.starlark.net/starlark"
)

// CallContext bounds a capability timeout by the host's invocation ceiling.
func CallContext(host Host, timeout string) (context.Context, context.CancelFunc, error) {
	duration := host.Options().Limits.Duration
	if timeout != "" {
		parsed, err := time.ParseDuration(timeout)
		if err != nil || parsed <= 0 {
			return nil, nil, errors.New("timeout must be a positive duration string")
		}
		duration = min(duration, parsed)
	}
	ctx, cancel := context.WithTimeout(host.Context(), duration)
	return ctx, cancel, nil
}

// ReadBounded refuses output beyond the host byte ceiling.
func ReadBounded(reader io.Reader, limit int) ([]byte, error) {
	data, err := io.ReadAll(io.LimitReader(reader, int64(limit)+1))
	if err != nil {
		return nil, errors.New("could not read bounded output")
	}
	if len(data) > limit {
		return nil, errors.New("output byte limit exceeded")
	}
	return data, nil
}

// DecodeJSON retains exact numbers and rejects trailing JSON values.
func DecodeJSON(data []byte) (any, error) {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	var value any
	if err := decoder.Decode(&value); err != nil {
		return nil, err
	}
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		return nil, errors.New("unexpected trailing JSON data")
	}
	return value, nil
}

// Argument finds an unpacked argument without converting its structured data.
func Argument(args starlark.Tuple, kwargs []starlark.Tuple, name string, index int) starlark.Value {
	for _, pair := range kwargs {
		if key, _ := starlark.AsString(pair[0]); key == name {
			return pair[1]
		}
	}
	if index < len(args) {
		return args[index]
	}
	return starlark.None
}

// RecordedInt accepts only an exact JSON integer in the host int range.
func RecordedInt(value any) (int, bool) {
	number, ok := value.(json.Number)
	if !ok {
		return 0, false
	}
	integer, err := strconv.Atoi(string(number))
	return integer, err == nil
}
