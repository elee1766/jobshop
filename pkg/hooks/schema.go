package hooks

import (
	"encoding/json"
	"fmt"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/encoding/jsonschema"
)

// DefaultOutputBytes bounds a step's structured output when
// automation.output_bytes is unset.
const DefaultOutputBytes = 1 << 20

// CheckStepOutput validates a finished step's structured output: its size
// against limit (0 is DefaultOutputBytes) and, when the step declared one,
// its JSON Schema.
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
	return CheckSchema(schema, output, true)
}

// CheckSchema compiles a step's output_schema and, when validate is set,
// checks output against it. With validate unset it only proves the schema
// is acceptable, which start_step does before the step exists. A nil schema
// accepts anything.
func CheckSchema(schema map[string]any, output map[string]any, validate bool) error {
	if schema == nil {
		return nil
	}
	data, err := json.Marshal(schema)
	if err != nil {
		return err
	}
	ctx := cuecontext.New()
	file, err := jsonschema.Extract(ctx.CompileBytes(data), &jsonschema.Config{Strict: true})
	if err != nil {
		return fmt.Errorf("output_schema: %w", err)
	}
	value := ctx.BuildFile(file)
	if err := value.Err(); err != nil {
		return fmt.Errorf("output_schema: %w", err)
	}
	if !validate {
		return nil
	}
	data, err = json.Marshal(output)
	if err != nil {
		return err
	}
	if err := value.Unify(ctx.CompileBytes(data)).Validate(cue.Concrete(true)); err != nil {
		return fmt.Errorf("step output: %w", err)
	}
	return nil
}
