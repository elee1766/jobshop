package config

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"

	"github.com/elee1766/jobshop/pkg/pack"
)

// IncludeKey lists files merged under the root config before its own keys
// (#531). The NixOS module writes the keys it owns (identities, gateway,
// sandbox defaults, telemetry, live packs) to jobshop.module.yaml, so a
// user-managed jobshop.yaml holds only the user's keys and names the
// module file here. An explicit key, not a jobshop.d/ lookup, so the
// file says where every key comes from.
const IncludeKey = "include"

// appendedKeys are top-level lists an include adds to instead of
// replacing: both files contribute packs and shared tool dirs.
var appendedKeys = []string{"packs", "shared_tools", "prompt_rules", "hooks"}

// applyIncludes merges the files named by top's include list (paths
// relative to dir) and then top itself. Maps merge per key, the
// including file winning; appendedKeys concatenate, includes first.
// Included files cannot include others.
func applyIncludes(top map[string]any, dir, name string) (map[string]any, error) {
	raw, ok := top[IncludeKey]
	if !ok {
		return top, nil
	}
	delete(top, IncludeKey)
	list, ok := raw.([]any)
	if !ok {
		return nil, fmt.Errorf("%s: include must be a list of file paths", name)
	}
	merged := map[string]any{}
	for i, e := range list {
		p, ok := e.(string)
		if !ok || p == "" {
			return nil, fmt.Errorf("%s: include[%d] must be a file path", name, i)
		}
		if !filepath.IsAbs(p) {
			p = filepath.Join(dir, p)
		}
		b, err := os.ReadFile(p)
		if err != nil {
			return nil, fmt.Errorf("%s: include[%d]: %w", name, i, err)
		}
		m, err := pack.DecodeMap(b)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", filepath.Base(p), err)
		}
		if _, nested := m[IncludeKey]; nested {
			return nil, fmt.Errorf("%s: included files cannot include others", filepath.Base(p))
		}
		mergeTop(merged, m)
	}
	mergeTop(merged, top)
	return merged, nil
}

func mergeTop(dst, src map[string]any) {
	for k, v := range src {
		if slices.Contains(appendedKeys, k) {
			old, _ := dst[k].([]any)
			add, ok := v.([]any)
			if ok {
				dst[k] = append(slices.Clone(old), add...)
				continue
			}
		}
		dst[k] = pack.MergeValue(dst[k], v)
	}
}
