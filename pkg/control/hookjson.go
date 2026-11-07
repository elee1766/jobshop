package control

import "encoding/json"

// decodeHookJSON keeps structured numeric inputs exact across durable recovery.
func decodeHookJSON(data []byte, out any) error {
	decoder := json.NewDecoder(bytesReader(data))
	decoder.UseNumber()
	return decoder.Decode(out)
}
