package control

import (
	"encoding/json"
	"fmt"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/encoding/jsonschema"

	"github.com/elee1766/jobshop/pkg/config"
)

func checkStepOutput(cfg *config.Config, schema, output map[string]any) error {
	data, err := json.Marshal(output)
	if err != nil {
		return err
	}
	limit := cfg.Automation.OutputBytes
	if limit == 0 {
		limit = 1 << 20
	}
	if len(data) > limit {
		return fmt.Errorf("step output exceeds %d bytes", limit)
	}
	return checkHookSchema(schema, output, true)
}

func checkHookSchema(schema map[string]any, output map[string]any, validate bool) error {
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
