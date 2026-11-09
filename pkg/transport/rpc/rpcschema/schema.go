// Package rpcschema describes the admin RPC types for Desk clients
// (#542): a JSON schema per method for GET /api/rpc, and TypeScript
// declarations for rpc.d.ts. Both walk the Go types in rpc.Specs
// by reflection, following encoding/json's field rules.
package rpcschema

import (
	"encoding"
	"encoding/json"
	"reflect"
	"strings"
	"time"
)

var (
	rawType      = reflect.TypeFor[json.RawMessage]()
	timeType     = reflect.TypeFor[time.Time]()
	durationType = reflect.TypeFor[time.Duration]()
	textType     = reflect.TypeFor[encoding.TextMarshaler]()
)

// field is one JSON object member of a struct.
type field struct {
	name     string
	typ      reflect.Type
	optional bool
}

// fields lists t's JSON members in order, flattening embedded structs
// without a tag the way encoding/json does.
func fields(t reflect.Type) []field {
	var out []field
	for i := range t.NumField() {
		f := t.Field(i)
		tag := f.Tag.Get("json")
		if tag == "-" {
			continue
		}
		name, opts, _ := strings.Cut(tag, ",")
		ft := f.Type
		if f.Anonymous && name == "" {
			if ft.Kind() == reflect.Pointer {
				ft = ft.Elem()
			}
			if ft.Kind() == reflect.Struct {
				out = append(out, fields(ft)...)
				continue
			}
		}
		if !f.IsExported() {
			continue
		}
		if name == "" {
			name = f.Name
		}
		opt := strings.Contains(opts, "omitempty") || strings.Contains(opts, "omitzero") || ft.Kind() == reflect.Pointer
		out = append(out, field{name: name, typ: ft, optional: opt})
	}
	return out
}

// Schema returns a JSON schema (draft 2020-12 subset) for v's type.
// Named struct types go in defs and are referenced, so recursive types
// (TraceNode) terminate.
func Schema(v any, defs map[string]any) map[string]any {
	if v == nil {
		return map[string]any{"type": "object"}
	}
	return schemaOf(reflect.TypeOf(v), defs)
}

func defName(t reflect.Type) string {
	pkg := t.PkgPath()
	if i := strings.LastIndex(pkg, "/"); i >= 0 {
		pkg = pkg[i+1:]
	}
	return pkg + "." + t.Name()
}

func schemaOf(t reflect.Type, defs map[string]any) map[string]any {
	switch {
	case t == rawType:
		return map[string]any{}
	case t == timeType:
		return map[string]any{"type": "string", "format": "date-time"}
	case t == durationType:
		return map[string]any{"type": "integer", "description": "nanoseconds"}
	case t.Kind() != reflect.Struct && t.Implements(textType):
		return map[string]any{"type": "string"}
	}
	switch t.Kind() {
	case reflect.Pointer:
		return schemaOf(t.Elem(), defs)
	case reflect.String:
		return map[string]any{"type": "string"}
	case reflect.Bool:
		return map[string]any{"type": "boolean"}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return map[string]any{"type": "integer"}
	case reflect.Float32, reflect.Float64:
		return map[string]any{"type": "number"}
	case reflect.Slice, reflect.Array:
		if t.Elem().Kind() == reflect.Uint8 {
			return map[string]any{"type": "string", "contentEncoding": "base64"}
		}
		return map[string]any{"type": "array", "items": schemaOf(t.Elem(), defs)}
	case reflect.Map:
		return map[string]any{"type": "object", "additionalProperties": schemaOf(t.Elem(), defs)}
	case reflect.Struct:
		if t.Name() == "" {
			return structSchema(t, defs)
		}
		n := defName(t)
		if _, ok := defs[n]; !ok {
			defs[n] = true // placeholder against recursion
			defs[n] = structSchema(t, defs)
		}
		return map[string]any{"$ref": "#/$defs/" + n}
	}
	return map[string]any{}
}

func structSchema(t reflect.Type, defs map[string]any) map[string]any {
	props := map[string]any{}
	req := []string{}
	for _, f := range fields(t) {
		props[f.name] = schemaOf(f.typ, defs)
		if !f.optional {
			req = append(req, f.name)
		}
	}
	s := map[string]any{"type": "object", "properties": props}
	if len(req) > 0 {
		s["required"] = req
	}
	return s
}
