package automation

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

// Conversion budgets also reject cyclic input/output without walking indefinitely.
type dataBudget struct{ remaining int }

func (b *dataBudget) take(size, depth int) error {
	b.remaining -= size + 1
	if b.remaining < 0 || depth > 64 {
		return errors.New("structured data byte/depth limit exceeded")
	}
	return nil
}

func toValue(value any, limit int) (starlark.Value, error) {
	return (&dataBudget{remaining: limit}).encode(reflect.ValueOf(value), 0)
}

func (b *dataBudget) encode(v reflect.Value, depth int) (starlark.Value, error) {
	if err := b.take(0, depth); err != nil {
		return nil, err
	}
	if !v.IsValid() {
		return starlark.None, nil
	}
	if v.Kind() == reflect.Interface || v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return starlark.None, nil
		}
		return b.encode(v.Elem(), depth+1)
	}
	if v.Type() == reflect.TypeFor[json.Number]() {
		n := v.Interface().(json.Number)
		if i, err := n.Int64(); err == nil {
			return starlark.MakeInt64(i), nil
		}
		if !strings.ContainsAny(string(n), ".eE") {
			u, err := strconv.ParseUint(string(n), 10, 64)
			if err != nil {
				return nil, errors.New("JSON integer exceeds 64-bit structured data range")
			}
			return starlark.MakeUint64(u), nil
		}
		f, err := n.Float64()
		if err != nil || math.IsInf(f, 0) || math.IsNaN(f) {
			return nil, errors.New("invalid JSON number")
		}
		return starlark.Float(f), nil
	}
	switch v.Kind() {
	case reflect.Bool:
		return starlark.Bool(v.Bool()), nil
	case reflect.String:
		if err := b.take(len(v.String()), depth); err != nil {
			return nil, err
		}
		return starlark.String(v.String()), nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return starlark.MakeInt64(v.Int()), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return starlark.MakeUint64(v.Uint()), nil
	case reflect.Float32, reflect.Float64:
		f := v.Float()
		if math.IsNaN(f) || math.IsInf(f, 0) {
			return nil, errors.New("non-finite number")
		}
		return starlark.Float(f), nil
	case reflect.Slice, reflect.Array:
		values := make([]starlark.Value, 0, min(v.Len(), b.remaining))
		for i := 0; i < v.Len(); i++ {
			value, err := b.encode(v.Index(i), depth+1)
			if err != nil {
				return nil, err
			}
			values = append(values, value)
		}
		return starlark.NewList(values), nil
	case reflect.Map:
		if v.Type().Key().Kind() != reflect.String {
			return nil, errors.New("structured map keys must be strings")
		}
		keys := v.MapKeys()
		sort.Slice(keys, func(i, j int) bool { return keys[i].String() < keys[j].String() })
		dict := starlark.NewDict(0)
		for _, key := range keys {
			if err := b.take(len(key.String()), depth); err != nil {
				return nil, err
			}
			value, err := b.encode(v.MapIndex(key), depth+1)
			if err != nil {
				return nil, err
			}
			if err := dict.SetKey(starlark.String(key.String()), value); err != nil {
				return nil, err
			}
		}
		return dict, nil
	case reflect.Struct:
		attrs := make(starlark.StringDict)
		for i := 0; i < v.NumField(); i++ {
			field := v.Type().Field(i)
			if !field.IsExported() {
				continue
			}
			name := strings.Split(field.Tag.Get("json"), ",")[0]
			if name == "-" {
				continue
			}
			if name == "" {
				name = strings.ToLower(field.Name[:1]) + field.Name[1:]
			}
			if err := b.take(len(name), depth); err != nil {
				return nil, err
			}
			value, err := b.encode(v.Field(i), depth+1)
			if err != nil {
				return nil, err
			}
			attrs[name] = value
		}
		return starlarkstruct.FromStringDict(starlark.String(v.Type().Name()), attrs), nil
	default:
		return nil, fmt.Errorf("unsupported structured data type %s", v.Type())
	}
}

func fromValue(value starlark.Value, limit int) (any, error) {
	decoded, err := (&dataBudget{remaining: limit}).decode(value, 0)
	if err != nil {
		return nil, err
	}
	encoded, err := json.Marshal(decoded)
	if err != nil {
		return nil, err
	}
	if len(encoded) > limit {
		return nil, errors.New("structured data byte limit exceeded")
	}
	return decoded, nil
}

func (b *dataBudget) decode(value starlark.Value, depth int) (any, error) {
	if err := b.take(0, depth); err != nil {
		return nil, err
	}
	switch v := value.(type) {
	case starlark.NoneType:
		return nil, nil
	case starlark.Bool:
		return bool(v), nil
	case starlark.String:
		if err := b.take(len(v), depth); err != nil {
			return nil, err
		}
		return string(v), nil
	case starlark.Int:
		if i, ok := v.Int64(); ok {
			return i, nil
		}
		if u, ok := v.Uint64(); ok {
			return u, nil
		}
		return nil, errors.New("integer exceeds 64-bit structured data range")
	case starlark.Float:
		f := float64(v)
		if math.IsNaN(f) || math.IsInf(f, 0) {
			return nil, errors.New("non-finite result number")
		}
		return f, nil
	case *starlark.Dict:
		out := make(map[string]any, v.Len())
		for _, pair := range v.Items() {
			key, ok := starlark.AsString(pair[0])
			if !ok {
				return nil, errors.New("result map keys must be strings")
			}
			if err := b.take(len(key), depth); err != nil {
				return nil, err
			}
			item, err := b.decode(pair[1], depth+1)
			if err != nil {
				return nil, err
			}
			out[key] = item
		}
		return out, nil
	case *starlark.List:
		return b.sequence(v, depth)
	case starlark.Tuple:
		return b.sequence(v, depth)
	case *starlarkstruct.Struct:
		out := make(map[string]any)
		for _, name := range v.AttrNames() {
			if err := b.take(len(name), depth); err != nil {
				return nil, err
			}
			attr, err := v.Attr(name)
			if err != nil {
				return nil, err
			}
			item, err := b.decode(attr, depth+1)
			if err != nil {
				return nil, err
			}
			out[name] = item
		}
		return out, nil
	default:
		return nil, fmt.Errorf("unsupported result type %s", value.Type())
	}
}

func (b *dataBudget) sequence(v starlark.Indexable, depth int) (any, error) {
	out := make([]any, 0, min(v.Len(), b.remaining))
	for i := 0; i < v.Len(); i++ {
		item, err := b.decode(v.Index(i), depth+1)
		if err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, nil
}
