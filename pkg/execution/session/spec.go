package session

// This file is jobshop's session spec on top of acphost. It was
// acphost.Spec plus the conversation export setup (#45); that export was
// removed by #530, so a spec.json from an older controller may still
// carry conv_export, which is ignored. The JSON layout matches the
// spec.json pkg/host wrote, so older hosts read and adopt cleanly.

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/elee1766/acphost/pkg/acphost"
)

// Spec is acphost.Spec, stored flat in spec.json.
type Spec struct {
	acphost.Spec
}

// Write stores s in dir as acphost.SpecFile.
func Write(dir string, s Spec) error {
	b, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, acphost.SpecFile), b, 0o600)
}

// Read loads the spec from dir. acphost.ReadSpec checks the protocol.
func Read(dir string) (Spec, error) {
	base, err := acphost.ReadSpec(dir)
	if err != nil {
		return Spec{}, err
	}
	var s Spec
	b, err := os.ReadFile(filepath.Join(dir, acphost.SpecFile))
	if err != nil {
		return Spec{}, err
	}
	if err := json.Unmarshal(b, &s); err != nil {
		return Spec{}, fmt.Errorf("%s: %w", acphost.SpecFile, err)
	}
	s.Spec = base
	return s, nil
}

// HostOptions returns the acphost.Main options for the session in dir.
// There are none since conversation export was removed; it stays so
// detached hosts (acphost.MaybeChild in cli.Main, which passes "": its
// options apply to every host) and in-process hosts have one place to
// add them.
func HostOptions(string) []acphost.Option { return nil }
