package config

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/goccy/go-yaml"
)

// loadRootPack keeps the composition file unprefixed; its children are
// ordinary pack instances with their own vars and prefixes.
func loadRootPack(path, dir string, rewrite func(string) string) ([]packLayer, []PackRef, error) {
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
		layers, err := loadCUEPack(path)
		if err != nil {
			return nil, nil, packErr(err)
		}
		return layers, nil, nil
	}
	layers, err := loadPackFile(path, nil, "", 0, true)
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
			ref.Path = filepath.Join(fileDir(path), ref.Path)
		}
	}
	return layers, refs, nil
}

func packValidationError(err error, refs []PackRef, layers []packLayer) error {
	if len(refs) > 0 && strings.Contains(err.Error(), PackFileName) {
		return packErr(err)
	}
	for _, layer := range layers {
		if strings.Contains(err.Error(), layer.file) {
			return packErr(err)
		}
		for _, files := range layer.entrySources {
			for _, file := range files {
				if strings.Contains(err.Error(), file) {
					return packErr(err)
				}
			}
		}
	}
	return err
}

func rootPackRefs(m, vars map[string]any, head *packHead) error {
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
		Packs []PackRef `yaml:"packs"`
	}
	if err := yaml.UnmarshalWithOptions(b, &refs, yaml.Strict()); err != nil {
		return err
	}
	head.Packs = refs.Packs
	return nil
}
