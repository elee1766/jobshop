package control

import (
	"io/fs"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"

	"github.com/elee1766/jobshop/pkg/config"
)

// configWatcher reloads the config after jobshop.yaml, an include or a
// file in a pack dir changes, when reload.watch is on (#533). Changes are
// debounced on the controller clock, and the watch set is recomputed after
// every reload, since a reload can add or drop packs.
type configWatcher struct {
	c       *Controller
	w       *fsnotify.Watcher
	paths   config.WatchPaths
	watched map[string]bool // dirs given to fsnotify
	changed []string        // files changed since the last reload
	timer   <-chan time.Time
	// reloaded, if set, gets each watch reload's changed files and error.
	// Tests wait on it.
	reloaded func(files []string, err error)
}

func newConfigWatcher(c *Controller) *configWatcher {
	return &configWatcher{c: c, watched: map[string]bool{}}
}

// sync starts, stops or updates the fsnotify watches to match the active
// config.
func (cw *configWatcher) sync() {
	cfg := cw.c.config()
	if !cfg.Reload.Watch {
		cw.close()
		return
	}
	if cw.w == nil {
		w, err := fsnotify.NewWatcher()
		if err != nil {
			cw.c.log.Error("config watch: start", "err", err)
			return
		}
		cw.w = w
	}
	cw.paths = cfg.Watch
	want := map[string]bool{}
	for _, f := range cw.paths.Files {
		// The parent dir, so an editor's rename-over still shows up.
		want[filepath.Dir(f)] = true
	}
	for _, d := range cw.paths.Dirs {
		_ = filepath.WalkDir(d, func(p string, e fs.DirEntry, err error) error {
			if err == nil && e.IsDir() {
				want[p] = true
			}
			return nil
		})
	}
	for _, d := range cw.paths.CUEDirs {
		want[d] = true
	}
	for d := range cw.watched {
		if !want[d] {
			_ = cw.w.Remove(d)
			delete(cw.watched, d)
		}
	}
	for d := range want {
		if cw.watched[d] {
			continue
		}
		if err := cw.w.Add(d); err != nil {
			cw.c.log.Warn("config watch: add", "dir", d, "err", err)
			continue
		}
		cw.watched[d] = true
	}
}

func (cw *configWatcher) close() {
	if cw.w != nil {
		_ = cw.w.Close()
		cw.w = nil
	}
	clear(cw.watched)
	cw.changed, cw.timer = nil, nil
}

// relevant reports whether a change at p is in the watch set.
func (cw *configWatcher) relevant(p string) bool {
	if config.IgnoredChange(p) {
		return false
	}
	p = filepath.Clean(p)
	if slices.Contains(cw.paths.Files, p) {
		return true
	}
	if filepath.Ext(p) == ".cue" && slices.Contains(cw.paths.CUEDirs, filepath.Dir(p)) {
		return true
	}
	return slices.ContainsFunc(cw.paths.Dirs, func(d string) bool {
		return p == d || strings.HasPrefix(p, d+string(filepath.Separator))
	})
}

// run serves until the controller stops. Every reload, by the watcher or
// not, signals c.configSwapped so the watch set follows the config.
func (cw *configWatcher) run() {
	defer cw.close()
	cw.sync()
	for {
		var events <-chan fsnotify.Event
		var errs <-chan error
		if cw.w != nil {
			events, errs = cw.w.Events, cw.w.Errors
		}
		select {
		case <-cw.c.ctx.Done():
			return
		case <-cw.c.configSwapped:
			cw.sync()
		case ev, ok := <-events:
			if !ok {
				cw.w = nil
				continue
			}
			cw.event(ev)
		case err, ok := <-errs:
			if ok {
				cw.c.log.Warn("config watch", "err", err)
			}
		case <-cw.timer:
			cw.fire()
		}
	}
}

// event records a change and restarts the debounce.
func (cw *configWatcher) event(ev fsnotify.Event) {
	if ev.Op == fsnotify.Chmod || !cw.relevant(ev.Name) {
		return
	}
	if ev.Op.Has(fsnotify.Create) {
		// A new subdir in a pack dir needs its own watch.
		cw.sync()
	}
	if !slices.Contains(cw.changed, ev.Name) {
		cw.changed = append(cw.changed, ev.Name)
	}
	cw.timer = cw.c.after(cw.c.config().Reload.DebounceDuration())
}

// fire reloads after the debounce. The reload logs the result and swaps
// on success; a failure raises a config-invalid attention item until the
// next good reload.
func (cw *configWatcher) fire() {
	files := cw.changed
	cw.changed, cw.timer = nil, nil
	slices.Sort(files)
	err := cw.c.reloadWith(cw.c.ctx, "watch", files)
	cw.sync()
	if cw.reloaded != nil {
		cw.reloaded(files, err)
	}
}
