package desk

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"mime"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strings"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/rpc"
)

// Desk plugins (#517). A plugin is a directory with plugin.json and ES
// modules that define custom elements. Sources, later wins on the same
// name: the embedded built-ins (dist/plugins), each pack's desk/
// directory, then <root>/desk/plugins.

// PluginManifest is plugin.json.
type PluginManifest struct {
	Name        string         `json:"name"`
	Version     string         `json:"version,omitempty"`
	Entry       string         `json:"entry"`
	API         string         `json:"api,omitempty"`
	Contributes []Contribution `json:"contributes"`
}

// Contribution is one element a plugin puts into a slot.
type Contribution struct {
	Slot    string `json:"slot"`
	Element string `json:"element"`
	// ID and Title name a view (nav entry, ?view=<id>).
	ID    string `json:"id,omitempty"`
	Title string `json:"title,omitempty"`
	// Match selects targets: kind, label, workflow, pool for drawer and
	// action; type (MIME) and ext for artifact.
	Match map[string][]string `json:"match,omitempty"`
}

// Plugin is a resolved plugin as GET /api/plugins lists it.
type Plugin struct {
	PluginManifest
	// Source is "builtin", "pack:<path>" or "root".
	Source string `json:"source"`
	// URL is the base-relative directory its files are served from.
	URL    string         `json:"url"`
	Config map[string]any `json:"config,omitempty"`
	// Overrides is the source this one replaced, if any.
	Overrides string `json:"overrides,omitempty"`
	fsys      fs.FS
}

// PluginSlots are the slot names a contribution may use. session (#535)
// is the session panel's body: the shell owns the frame (side column on
// desktop, full screen on a phone) and the last session contribution
// renders the chat inside it.
var PluginSlots = []string{"view", "drawer", "artifact", "action", "header", "session"}

var (
	pluginNameRe = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{0,62}$`)
	// Custom element names need a hyphen.
	elementRe = regexp.MustCompile(`^[a-z][a-z0-9]*-[a-z0-9-]*$`)
	viewIDRe  = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{0,62}$`)
)

func (m PluginManifest) validate() error {
	if !pluginNameRe.MatchString(m.Name) {
		return fmt.Errorf("name %q: want lowercase letters, digits and hyphens", m.Name)
	}
	if !fs.ValidPath(m.Entry) || m.Entry == "." || !strings.HasSuffix(m.Entry, ".js") {
		return fmt.Errorf("entry %q: want a relative .js path inside the plugin", m.Entry)
	}
	for i, c := range m.Contributes {
		if !slices.Contains(PluginSlots, c.Slot) {
			return fmt.Errorf("contributes[%d]: slot %q: want one of %s", i, c.Slot, strings.Join(PluginSlots, ", "))
		}
		if !elementRe.MatchString(c.Element) {
			return fmt.Errorf("contributes[%d]: element %q is not a custom element name", i, c.Element)
		}
		if c.Slot == "view" && !viewIDRe.MatchString(c.ID) {
			return fmt.Errorf("contributes[%d]: view id %q: want lowercase letters, digits and hyphens", i, c.ID)
		}
	}
	return nil
}

// pluginSource is one directory of plugin directories.
type pluginSource struct {
	name string
	fsys fs.FS
}

// readPlugins reads every <dir>/plugin.json in src. A broken plugin is
// logged and skipped so one bad directory does not hide the rest.
func (s *Server) readPlugins(src pluginSource) []Plugin {
	ents, err := fs.ReadDir(src.fsys, ".")
	if err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			s.log.Warn("desk: plugins", "source", src.name, "err", err)
		}
		return nil
	}
	var out []Plugin
	for _, e := range ents {
		if !e.IsDir() {
			continue
		}
		sub, err := fs.Sub(src.fsys, e.Name())
		if err != nil {
			continue
		}
		b, err := fs.ReadFile(sub, "plugin.json")
		if err != nil {
			continue
		}
		var m PluginManifest
		if err := json.Unmarshal(b, &m); err != nil {
			s.log.Warn("desk: plugin.json", "source", src.name, "dir", e.Name(), "err", err)
			continue
		}
		if err := m.validate(); err != nil {
			s.log.Warn("desk: plugin.json", "source", src.name, "dir", e.Name(), "err", err)
			continue
		}
		if m.Name != e.Name() {
			s.log.Warn("desk: plugin.json", "source", src.name, "dir", e.Name(), "err", "name does not match its directory")
			continue
		}
		if m.Contributes == nil {
			m.Contributes = []Contribution{}
		}
		out = append(out, Plugin{PluginManifest: m, Source: src.name, URL: "plugins/" + m.Name + "/", fsys: sub})
	}
	return out
}

// deskConfig is the desk section of the active config, read through
// config.show so a reload applies without restarting Desk. The pack
// paths come along for the plugin sources.
func (s *Server) deskConfig(r *http.Request) (config.Desk, []string) {
	var show rpc.ConfigShowResult
	if err := s.ctl.Call(r.Context(), rpc.MethodConfigShow, struct{}{}, &show); err != nil {
		s.log.WarnContext(r.Context(), "desk: config.show", "err", err)
		return config.Desk{}, nil
	}
	var d config.Desk
	if raw, ok := show.Config["desk"]; ok {
		b, _ := json.Marshal(raw)
		if err := json.Unmarshal(b, &d); err != nil {
			s.log.WarnContext(r.Context(), "desk: desk config", "err", err)
		}
	}
	var packs []string
	for _, in := range show.Instances {
		packs = append(packs, in.Path)
	}
	return d, packs
}

// dirFS is a directory as an fs.FS that cannot be escaped by symlinks.
// A missing directory is an empty source. Roots stay open per path (one
// per pack and the root dir); one whose directory is gone is reopened.
func (s *Server) dirFS(dir string) fs.FS {
	s.rootsMu.Lock()
	defer s.rootsMu.Unlock()
	if r, ok := s.roots[dir]; ok {
		if _, err := r.Stat("."); err == nil {
			return r.FS()
		}
		_ = r.Close()
		delete(s.roots, dir)
	}
	root, err := os.OpenRoot(dir)
	if err != nil {
		return emptyFS{}
	}
	if s.roots == nil {
		s.roots = map[string]*os.Root{}
	}
	s.roots[dir] = root
	return root.FS()
}

type emptyFS struct{}

func (emptyFS) Open(string) (fs.File, error) { return nil, fs.ErrNotExist }

// plugins resolves the enabled plugins, sorted by name.
func (s *Server) plugins(r *http.Request) []Plugin {
	dc, packs := s.deskConfig(r)
	srcs := []pluginSource{{name: "builtin", fsys: s.builtin}}
	for _, p := range packs {
		dir := p
		if !filepath.IsAbs(dir) && s.cfg.Root != "" {
			dir = filepath.Join(s.cfg.Root, dir)
		}
		if filepath.IsAbs(dir) {
			srcs = append(srcs, pluginSource{name: "pack:" + p, fsys: s.dirFS(filepath.Join(dir, "desk"))})
		}
	}
	if s.cfg.Root != "" {
		srcs = append(srcs, pluginSource{name: "root", fsys: s.dirFS(filepath.Join(s.cfg.Root, "desk", "plugins"))})
	}
	byName := map[string]Plugin{}
	for _, src := range srcs {
		for _, p := range s.readPlugins(src) {
			if old, ok := byName[p.Name]; ok {
				p.Overrides = old.Source
			}
			byName[p.Name] = p
		}
	}
	out := make([]Plugin, 0, len(byName))
	for name, p := range byName {
		pc := dc.Plugins[name]
		if pc.Enabled != nil && !*pc.Enabled {
			continue
		}
		p.Config = pc.Config
		out = append(out, p)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

func (s *Server) pluginList(r *http.Request) (any, error) {
	return s.plugins(r), nil
}

// pluginFile serves GET /plugins/{name}/{file...} from the plugin that
// won its name. Paths are fs.ValidPath, so no "..", and directory
// sources are os.Root, so no symlink escape.
func (s *Server) pluginFile(w http.ResponseWriter, r *http.Request) {
	name, file := r.PathValue("name"), r.PathValue("file")
	if !fs.ValidPath(file) || file == "." {
		http.NotFound(w, r)
		return
	}
	var p *Plugin
	for _, x := range s.plugins(r) {
		if x.Name == name {
			p = &x
			break
		}
	}
	if p == nil {
		http.NotFound(w, r)
		return
	}
	f, err := p.fsys.Open(file)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	defer func() { _ = f.Close() }()
	st, err := f.Stat()
	if err != nil || st.IsDir() {
		http.NotFound(w, r)
		return
	}
	ct := mime.TypeByExtension(path.Ext(file))
	switch path.Ext(file) {
	case ".js", ".mjs":
		ct = "text/javascript; charset=utf-8"
	case ".json":
		ct = "application/json"
	}
	if ct == "" {
		ct = "application/octet-stream"
	}
	h := w.Header()
	h.Set("Content-Type", ct)
	// Plugins change without a Desk rebuild; revalidate every load.
	h.Set("Cache-Control", "no-cache")
	if rs, ok := f.(io.ReadSeeker); ok {
		http.ServeContent(w, r, path.Base(file), st.ModTime(), rs)
		return
	}
	if r.Method != http.MethodHead {
		_, _ = io.Copy(w, f)
	}
}
