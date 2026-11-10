package hooks

import (
	"encoding/json"
	"fmt"

	packcue "github.com/gfx-labs/jobshop/pkg/pack/cue"
)

// DefaultOutputBytes bounds a step's structured output when
// automation.output_bytes is unset.
const DefaultOutputBytes = 1 << 20

// CheckStepOutput validates a finished step's structured output: its size
// against limit (0 is DefaultOutputBytes) and, when the step declared one,
// its JSON Schema. Schema evaluation is pack/cue's.
func CheckStepOutput(limit int, schema, output map[string]any) error {
	data, err := json.Marshal(output)
	if err != nil {
		return err
	}
	if limit == 0 {
		limit = DefaultOutputBytes
	}
	if len(data) > limit {
		return fmt.Errorf("step output exceeds %d bytes", limit)
	}
	return packcue.CheckSchema(schema, output, true)
}
