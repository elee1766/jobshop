package sandbox

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"

	"github.com/gfx-labs/acphost/pkg/sandbox"
	"github.com/gfx-labs/acphost/pkg/xerr"
	"golang.org/x/sys/unix"
)

// CheckMounts fails with the first sandbox mount source (or the cwd) the
// mapped uid cannot use, naming the path and its mode. Without it the
// agent fails at its first open and the pool crash-loops into
// quarantine. With a sandbox bwrap binds as the controller (acphost#162,
// JOB-53) and the agent sees only the mount itself, so only the source's
// own mode and ACL count; its host-side parents (a 0700 home) do not.
// Without a sandbox the agent opens host paths, so every parent counts.
func CheckMounts(s *sandbox.Sandbox, cwd string, uid int, gids []int) error {
	type need struct {
		path string
		mode uint32
	}
	var ns []need
	if s != nil {
		for _, m := range s.RO {
			if _, err := os.Stat(m.Src); m.SkipMissing && err != nil {
				continue
			}
			ns = append(ns, need{m.Src, unix.R_OK})
		}
		for _, m := range s.RW {
			ns = append(ns, need{m.Src, unix.R_OK | unix.W_OK})
		}
	}
	ns = append(ns, need{cwd, unix.R_OK | unix.W_OK})
	check := reachable
	if s != nil {
		check = usable
	}
	for _, n := range ns {
		if err := check(n.path, n.mode, uid, gids); err != nil {
			return xerr.Permanentf("isolation: mapped uid %d cannot use %w", uid, err)
		}
	}
	return nil
}

// reachable checks uid can search every directory above p and has want
// on p itself.
func reachable(p string, want uint32, uid int, gids []int) error {
	p = filepath.Clean(p)
	var dirs []string
	for d := filepath.Dir(p); ; d = filepath.Dir(d) {
		dirs = append(dirs, d)
		if d == filepath.Dir(d) {
			break
		}
	}
	slices.Reverse(dirs)
	for _, d := range dirs {
		if err := Allowed(d, unix.X_OK, uid, gids); err != nil {
			return err
		}
	}
	return usable(p, want, uid, gids)
}

// usable checks uid has want on p itself (and search, for a dir).
func usable(p string, want uint32, uid int, gids []int) error {
	st, err := os.Stat(p)
	if err != nil {
		return err
	}
	if st.IsDir() {
		want |= unix.X_OK
	}
	return Allowed(p, want, uid, gids)
}

// Allowed evaluates p's mode bits and POSIX ACL (access check order of
// acl(5)) for uid and gids.
func Allowed(p string, want uint32, uid int, gids []int) error {
	var st unix.Stat_t
	if err := unix.Stat(p, &st); err != nil {
		return err
	}
	deny := fmt.Errorf("%s (mode %04o, owner %d:%d)", p, st.Mode&0o7777, st.Uid, st.Gid)
	ok := func(perm uint32) error {
		if perm&want == want {
			return nil
		}
		return deny
	}
	if int(st.Uid) == uid {
		return ok(st.Mode >> 6 & 7)
	}
	inGroup := func(g uint32) bool { return slices.Contains(gids, int(g)) }
	buf := make([]byte, 1024)
	n, err := unix.Getxattr(p, aclAccessKey, buf)
	if err != nil || n < 4 {
		// No ACL: plain mode bits.
		if inGroup(st.Gid) {
			return ok(st.Mode >> 3 & 7)
		}
		return ok(st.Mode & 7)
	}
	mask, other := uint32(7), uint32(0)
	var user *uint32
	var groups []uint32
	matched := false
	for i := 4; i+8 <= n; i += 8 {
		tag := binary.LittleEndian.Uint16(buf[i:])
		perm := uint32(binary.LittleEndian.Uint16(buf[i+2:]))
		id := binary.LittleEndian.Uint32(buf[i+4:])
		switch tag {
		case aclUser:
			if int(id) == uid {
				user = &perm
			}
		case aclGroupObj:
			if inGroup(st.Gid) {
				groups, matched = append(groups, perm), true
			}
		case aclGroup:
			if inGroup(id) {
				groups, matched = append(groups, perm), true
			}
		case aclMask:
			mask = perm
		case aclOther:
			other = perm
		}
	}
	if user != nil {
		return ok(*user & mask)
	}
	if matched {
		for _, g := range groups {
			if g&mask&want == want {
				return nil
			}
		}
		return deny
	}
	return ok(other)
}

// CopyCredential copies src to dst through a temp file, so the agent never
// reads a half-written login. 0660: with a default ACL the group bits
// are the mask, so 0600 would shut out the mapped uid.
func CopyCredential(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() { _ = in.Close() }()
	tmp, err := os.CreateTemp(filepath.Dir(dst), ".cred-*")
	if err != nil {
		return err
	}
	defer func() { _ = os.Remove(tmp.Name()) }()
	if _, err := io.Copy(tmp, in); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Chmod(0o660); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), dst)
}
