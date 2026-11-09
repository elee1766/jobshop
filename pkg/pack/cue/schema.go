package cue

import (
	"encoding/json"
	"fmt"

	cuelang "cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/encoding/jsonschema"
)

// CheckSchema compiles a JSON Schema given as decoded JSON and, when
// validate is set, checks output against it. With validate unset it only
// proves the schema is acceptable. A nil schema accepts anything. Hooks
// and command steps use it for a step's output_schema, so this package
// stays the only one that imports the CUE evaluator.
func CheckSchema(schema, output map[string]any, validate bool) error {
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
	if err := value.Unify(ctx.CompileBytes(data)).Validate(cuelang.Concrete(true)); err != nil {
		return fmt.Errorf("step output: %w", err)
	}
	return nil
}
