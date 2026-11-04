// Package ingest receives webhook deliveries and keeps them in a spool
// directory until the controller drains them (#318).
package ingest

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"time"
)

// Sources.
const (
	SourceGitHub = "github"
	SourceLinear = "linear"
)

// Delivery is one spooled webhook delivery.
type Delivery struct {
	Source     string    `json:"source"`
	DeliveryID string    `json:"delivery_id"`
	Kind       string    `json:"kind"`
	ReceivedAt time.Time `json:"received_at"`
	// Headers are the signature and event headers only.
	Headers map[string]string `json:"headers"`
	Payload json.RawMessage   `json:"payload"`
}

// Name is the spool file name of d.
func (d Delivery) Name() string { return d.Source + "-" + d.DeliveryID + ".json" }

// validID keeps delivery ids safe as file name parts.
var validID = regexp.MustCompile(`^[A-Za-z0-9._-]{1,128}$`)

// ErrBadDelivery is a delivery that cannot be named in the spool.
var ErrBadDelivery = errors.New("ingest: bad delivery")

// Spool is a directory of deliveries, one file each. Temp files start
// with a dot, so readers never see a partial write.
type Spool struct{ Dir string }

// Write stores d. It returns false when a delivery with the same source
// and id is already spooled; the file is not touched then.
func (s Spool) Write(d Delivery) (bool, error) {
	if (d.Source != SourceGitHub && d.Source != SourceLinear) || !validID.MatchString(d.DeliveryID) || strings.HasPrefix(d.DeliveryID, ".") {
		return false, fmt.Errorf("%w: source %q id %q", ErrBadDelivery, d.Source, d.DeliveryID)
	}
	final := filepath.Join(s.Dir, d.Name())
	if _, err := os.Lstat(final); err == nil {
		return false, nil
	}
	b, err := json.Marshal(d)
	if err != nil {
		return false, err
	}
	if err := os.MkdirAll(s.Dir, 0o700); err != nil {
		return false, err
	}
	f, err := os.CreateTemp(s.Dir, ".tmp-*")
	if err != nil {
		return false, err
	}
	tmp := f.Name()
	defer func() { _ = os.Remove(tmp) }()
	if _, err := f.Write(b); err != nil {
		_ = f.Close()
		return false, err
	}
	if err := f.Sync(); err != nil {
		_ = f.Close()
		return false, err
	}
	if err := f.Close(); err != nil {
		return false, err
	}
	// Link, not rename: it fails on an existing name, so two concurrent
	// writes of one delivery keep the first file.
	if err := os.Link(tmp, final); err != nil {
		if errors.Is(err, os.ErrExist) {
			return false, nil
		}
		return false, err
	}
	return true, syncDir(s.Dir)
}

// List returns the names of spooled deliveries, oldest name first. A
// missing dir is an empty spool.
func (s Spool) List() ([]string, error) {
	ents, err := os.ReadDir(s.Dir)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var out []string
	for _, e := range ents {
		n := e.Name()
		if e.Type().IsRegular() && !strings.HasPrefix(n, ".") && strings.HasSuffix(n, ".json") {
			out = append(out, n)
		}
	}
	slices.Sort(out)
	return out, nil
}

// Read decodes the delivery in file name.
func (s Spool) Read(name string) (Delivery, error) {
	var d Delivery
	if filepath.Base(name) != name || strings.HasPrefix(name, ".") {
		return d, fmt.Errorf("%w: name %q", ErrBadDelivery, name)
	}
	b, err := os.ReadFile(filepath.Join(s.Dir, name))
	if err != nil {
		return d, err
	}
	if err := json.Unmarshal(b, &d); err != nil {
		return d, fmt.Errorf("ingest: %s: %w", name, err)
	}
	return d, nil
}

// Remove deletes file name. A missing file is not an error.
func (s Spool) Remove(name string) error {
	if filepath.Base(name) != name || strings.HasPrefix(name, ".") {
		return fmt.Errorf("%w: name %q", ErrBadDelivery, name)
	}
	if err := os.Remove(filepath.Join(s.Dir, name)); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return syncDir(s.Dir)
}

func syncDir(dir string) error {
	d, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer func() { _ = d.Close() }()
	return d.Sync()
}
