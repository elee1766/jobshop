package pack

import (
	"fmt"
	"path/filepath"

	"github.com/goccy/go-yaml"
)

// loadRootPack keeps the composition file unprefixed; its children are
// ordinary pack instances with their own vars and prefixes.
func (ld loader) loadRoot(path, dir string, rewrite func(string) string) ([]layer, []Ref, error) {
	if path == "" {
		return nil, nil, nil
	}
	if !filepath.IsAbs(path) {
		path = filepath.Join(dir, path)
	}
	path, err := filepath.Abs(path)
	if err != nil {
		return nil, nil, packErr(err)
	}
	if rewrite != nil {
		path = rewrite(path)
	}
	if filepath.Ext(path) == ".cue" {
		layers, err := ld.loadCUE(path)
		if err != nil {
			return nil, nil, packErr(err)
		}
		return layers, nil, nil
	}
	layers, err := ld.loadPackFile(path, nil, "", 0, true)
	if err != nil {
		return nil, nil, packErr(err)
	}
	refs := layers[len(layers)-1].packs
	for i := range refs {
		ref := &refs[i]
		if ref.Path == "" {
			return nil, nil, packErr(fmt.Errorf("%s: packs[%d]: path is required", path, i))
		}
		if ref.Prefix != "" && !prefixPattern.MatchString(ref.Prefix) {
			return nil, nil, packErr(fmt.Errorf("%s: packs[%d]: prefix %q must match %s", path, i, ref.Prefix, prefixPattern))
		}
		if !filepath.IsAbs(ref.Path) {
			ref.Path = filepath.Join(FileDir(path), ref.Path)
		}
	}
	return layers, refs, nil
}

func rootPackRefs(m, vars map[string]any, h *head) error {
	raw, ok := m["packs"]
	if !ok {
		return nil
	}
	delete(m, "packs")
	value, err := substitute(raw, vars, "packs")
	if err != nil {
		return err
	}
	b, err := yaml.Marshal(map[string]any{"packs": value})
	if err != nil {
		return err
	}
	var refs struct {
		Packs []Ref `yaml:"packs"`
	}
	if err := yaml.UnmarshalWithOptions(b, &refs, yaml.Strict()); err != nil {
		return err
	}
	h.Packs = refs.Packs
	return nil
}
