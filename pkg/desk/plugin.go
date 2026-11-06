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
	// Icon, Key and Order tell a nav how to list a view (JOB-86): an icon
	// name the nav knows, the letter after "g" that jumps to it, and its
	// position (1 and up, lower first; unset lists last).
	Icon  string `json:"icon,omitempty"`
	Key   string `json:"key,omitempty"`
	Order int    `json:"order,omitempty"`
	// Placement says where the built-in app docks a nav (JOB-86).
	Placement *Placement `json:"placement,omitempty"`
	// Match selects targets: kind, label, workflow, pool for drawer and
	// action; type (MIME) and ext for artifact.
	Match map[string][]string `json:"match,omitempty"`
}

// Placement is where the app docks a nav contribution: Wide above the
// app's breakpoint, Narrow below it. Each is left, right, top, bottom or
// none (the nav positions itself); unset means left and bottom.
type Placement struct {
	Wide   string `json:"wide,omitempty"`
	Narrow string `json:"narrow,omitempty"`
}

var placements = []string{"left", "right", "top", "bottom", "none"}

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

// PluginSlots are the built-in slot names. session (#535) is the session
// panel's body: the app owns the frame (side column on desktop, full
// screen on a phone) and the last session contribution renders the chat
// inside it. app (JOB-85) is the whole page: the shell mounts one, the
// built-in desk plugin unless another plugin contributes app. Any other
// slot is plugin-defined (JOB-87): <plugin>:<name>, hosted by the plugin
// named before the colon through api.slot.
var PluginSlots = []string{"view", "drawer", "artifact", "action", "header", "session", "app", "nav"}

var (
	pluginNameRe = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{0,62}$`)
	// A plugin-defined slot: the host plugin's name, a colon, a name.
	pluginSlotRe = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{0,62}:[a-z0-9][a-z0-9-]{0,62}$`)
	// Custom element names need a hyphen.
	elementRe = regexp.MustCompile(`^[a-z][a-z0-9]*-[a-z0-9-]*$`)
	viewIDRe  = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{0,62}$`)
	// Icon names are lucide's: lowercase words joined by hyphens.
	iconRe = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{0,40}$`)
	keyRe  = regexp.MustCompile(`^[a-z0-9]$`)
)

// maxOrder bounds Contribution.Order so a typo cannot overflow a sort.
const maxOrder = 10000

func (m PluginManifest) validate() error {
	if !pluginNameRe.MatchString(m.Name) {
		return fmt.Errorf("name %q: want lowercase letters, digits and hyphens", m.Name)
	}
	if !fs.ValidPath(m.Entry) || m.Entry == "." || !strings.HasSuffix(m.Entry, ".js") {
		return fmt.Errorf("entry %q: want a relative .js path inside the plugin", m.Entry)
	}
	for i, c := range m.Contributes {
		if !slices.Contains(PluginSlots, c.Slot) && !pluginSlotRe.MatchString(c.Slot) {
			return fmt.Errorf("contributes[%d]: slot %q: want one of %s, or <plugin>:<name>", i, c.Slot, strings.Join(PluginSlots, ", "))
		}
		if !elementRe.MatchString(c.Element) {
			return fmt.Errorf("contributes[%d]: element %q is not a custom element name", i, c.Element)
		}
		if c.Slot == "view" && !viewIDRe.MatchString(c.ID) {
			return fmt.Errorf("contributes[%d]: view id %q: want lowercase letters, digits and hyphens", i, c.ID)
		}
		if err := c.validateNav(); err != nil {
			return fmt.Errorf("contributes[%d]: %w", i, err)
		}
	}
	return nil
}

// validateNav checks the fields a nav reads: icon, key and order on a
// view, placement on a nav. On another slot they are an error, so a
// wrong slot name does not silently drop them.
func (c Contribution) validateNav() error {
	if c.Slot != "view" {
		switch {
		case c.Icon != "":
			return fmt.Errorf("icon is only for slot view, not %q", c.Slot)
		case c.Key != "":
			return fmt.Errorf("key is only for slot view, not %q", c.Slot)
		case c.Order != 0:
			return fmt.Errorf("order is only for slot view, not %q", c.Slot)
		}
	}
	if c.Icon != "" && !iconRe.MatchString(c.Icon) {
		return fmt.Errorf("icon %q: want lowercase letters, digits and hyphens", c.Icon)
	}
	if c.Key != "" && !keyRe.MatchString(c.Key) {
		return fmt.Errorf("key %q: want one lowercase letter or digit", c.Key)
	}
	if c.Order < 0 || c.Order > maxOrder {
		return fmt.Errorf("order %d: want 0 to %d", c.Order, maxOrder)
	}
	if c.Placement == nil {
		return nil
	}
	if c.Slot != "nav" {
		return fmt.Errorf("placement is only for slot nav, not %q", c.Slot)
	}
	for _, p := range []struct{ name, v string }{{"wide", c.Placement.Wide}, {"narrow", c.Placement.Narrow}} {
		if p.v != "" && !slices.Contains(placements, p.v) {
			return fmt.Errorf("placement.%s %q: want one of %s", p.name, p.v, strings.Join(placements, ", "))
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
