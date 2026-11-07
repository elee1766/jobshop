package hooks

import (
	"bytes"
	"encoding/json"
)

// DecodeJSON keeps structured numeric inputs exact across durable recovery.
func DecodeJSON(data []byte, out any) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	return decoder.Decode(out)
}
