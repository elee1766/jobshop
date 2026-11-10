package desk

import (
	"errors"
	"fmt"
	"io/fs"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/gfx-labs/jobshop/pkg/transport/rpc"
)

// Artifacts (#508): list, read and ack deliverables. Attachments are read
// from the dir artifact.show returns; Desk runs beside the controller.

// artifactList proxies artifact.list. Query: unread, trigger, pool.
func (s *Server) artifactList(r *http.Request) (any, error) {
	q := r.URL.Query()
	p := rpc.ArtifactListParams{Unread: q.Get("unread") == "1" || q.Get("unread") == "true", Trigger: q.Get("trigger"), Pool: q.Get("pool")}
	var res []rpc.Artifact
	if err := s.ctl.Call(r.Context(), rpc.MethodArtifactList, p, &res); err != nil {
		return nil, err
	}
	return res, nil
}

func (s *Server) artifactShow(r *http.Request) (any, error) {
	var res rpc.ArtifactShowResult
	if err := s.ctl.Call(r.Context(), rpc.MethodArtifactShow, rpc.IDParams{ID: r.PathValue("id")}, &res); err != nil {
		return nil, err
	}
	res.Dir = "" // a host path; the UI uses the files endpoint
	return res, nil
}

func (s *Server) artifactAck(r *http.Request) (any, error) {
	var b struct {
		Note string `json:"note"`
	}
	if err := s.decode(r, &b); err != nil {
		return nil, err
	}
	var res rpc.Artifact
	if err := s.ctl.Call(r.Context(), rpc.MethodArtifactAck, rpc.ArtifactAckParams{ID: r.PathValue("id"), Note: b.Note}, &res); err != nil {
		return nil, err
	}
	return res, nil
}

// inlineTypes are shown in the page; everything else, svg included (it
// can carry script), is served as a download.
var inlineTypes = []string{"image/png", "image/jpeg", "image/gif", "image/webp"}

// artifactFile serves one attachment by its recorded name. The name must
// be in meta.artifact.files, and the file is opened through os.Root, so
// neither a crafted name nor a symlink reaches outside the artifact dir.
func (s *Server) artifactFile(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	if name == "" || name != filepath.Base(name) || strings.ContainsAny(name, `/\`) || name == "." || name == ".." {
		http.Error(w, "bad file name", http.StatusBadRequest)
		return
	}
	var res rpc.ArtifactShowResult
	if err := s.ctl.Call(r.Context(), rpc.MethodArtifactShow, rpc.IDParams{ID: r.PathValue("id")}, &res); err != nil {
		http.Error(w, "artifact not found", http.StatusNotFound)
		return
	}
	i := slices.IndexFunc(res.Meta.Files, func(f rpc.ArtifactFile) bool { return f.Name == name })
	if i < 0 || res.Dir == "" {
		http.Error(w, "file not found", http.StatusNotFound)
		return
	}
	if res.Meta.Files[i].Pruned {
		http.Error(w, "file pruned after retention", http.StatusGone)
		return
	}
	root, err := os.OpenRoot(res.Dir)
	if err != nil {
		http.Error(w, "file not found", http.StatusNotFound)
		return
	}
	defer func() { _ = root.Close() }()
	f, err := root.Open(name)
	if err != nil {
		code := http.StatusBadGateway
		if errors.Is(err, fs.ErrNotExist) {
			code = http.StatusNotFound
		}
		http.Error(w, "file not found", code)
		return
	}
	defer func() { _ = f.Close() }()
	fi, err := f.Stat()
	if err != nil || !fi.Mode().IsRegular() {
		http.Error(w, "file not found", http.StatusNotFound)
		return
	}
	typ := mime.TypeByExtension(filepath.Ext(name))
	disp := "attachment"
	base, _, _ := strings.Cut(typ, ";")
	switch {
	case slices.Contains(inlineTypes, base):
		disp = "inline"
	case strings.HasPrefix(base, "text/") || base == "application/json" || typ == "":
		// Text is shown as text, never rendered as html.
		typ, disp = "text/plain; charset=utf-8", "inline"
	default:
		typ = "application/octet-stream"
	}
	h := w.Header()
	h.Set("Content-Type", typ)
	h.Set("Content-Disposition", fmt.Sprintf("%s; filename=%s", disp, strconv.Quote(name)))
	h.Set("X-Content-Type-Options", "nosniff")
	h.Set("Content-Security-Policy", "default-src 'none'; sandbox")
	h.Set("Cache-Control", "no-store")
	http.ServeContent(w, r, "", fi.ModTime(), f)
}
