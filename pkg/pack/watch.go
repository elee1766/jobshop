package pack

import (
	"path/filepath"
	"slices"
	"strings"
)

// WatchPaths are what a config watcher follows: the config file and its
// includes as files, and every resolved pack dir (extends included)
// recursively. Recorded by the loader, so a pack added to the config is
// watched after the reload that loads it.
type WatchPaths struct {
	Files []string
	Dirs  []string
	// CUEDirs watch direct CUE siblings, without walking entire checkouts.
	CUEDirs []string
}

// AddFile watches one file.
func (w *WatchPaths) AddFile(p string) {
	if p = filepath.Clean(p); !slices.Contains(w.Files, p) {
		w.Files = append(w.Files, p)
	}
}

func (w *WatchPaths) addDir(p string) {
	if p = filepath.Clean(p); !slices.Contains(w.Dirs, p) {
		w.Dirs = append(w.Dirs, p)
	}
}

func (w *WatchPaths) addCUEDir(p string) {
	if p = filepath.Clean(p); !slices.Contains(w.CUEDirs, p) {
		w.CUEDirs = append(w.CUEDirs, p)
	}
}

// IgnoredChange reports editor temp and swap files, which never trigger
// a reload: vim's .x.swp and 4913 probe, emacs' x~ and .#x lock.
func IgnoredChange(path string) bool {
	b := filepath.Base(path)
	return b == "4913" || filepath.Ext(b) == ".swp" || filepath.Ext(b) == ".swx" ||
		len(b) > 0 && b[len(b)-1] == '~' || len(b) > 1 && b[:2] == ".#"
}

// FileDir is the directory relative pack and prompt file paths resolve
// against: the config file's own directory after following links (#532),
// so a jobshop.yaml linked from a git checkout finds the packs beside it.
// A link into the Nix store keeps the link's directory: a generated file
// there has no siblings, and relative paths in it meant the root.
func FileDir(path string) string {
	target, err := filepath.EvalSymlinks(path)
	if err != nil || strings.HasPrefix(target, "/nix/store/") {
		return filepath.Dir(path)
	}
	return filepath.Dir(target)
}
