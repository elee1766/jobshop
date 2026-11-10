package session

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"net"
	"time"

	host "github.com/gfx-labs/acphost/pkg/acphost"
)

// connectWait is how long Connect polls a host that is not listening yet.
const connectWait = 15 * time.Second

// Connect dials the host in dir and returns its connection, reader and
// welcome. It polls while a new host starts and stops early when the host
// recorded a failed spawn or exited.
func Connect(ctx context.Context, dir string) (net.Conn, *bufio.Reader, host.Welcome, error) {
	deadline := time.Now().Add(connectWait)
	// A new host listens within milliseconds: poll quickly at first, then
	// back off to the old 100ms.
	wait := 5 * time.Millisecond
	for {
		nc, br, wel, err := host.Connect(ctx, dir)
		if err == nil {
			return nc, br, wel, nil
		}
		// The host wrote exit.json with spawn_failed and is gone:
		// retrying until the deadline would only hide why (acphost#100).
		var se *host.SpawnError
		if errors.As(err, &se) {
			return nil, nil, wel, fmt.Errorf("connect to host: %w", err)
		}
		if time.Now().After(deadline) || ctx.Err() != nil {
			return nil, nil, wel, fmt.Errorf("connect to host: %w", err)
		}
		if ex, exited := host.ReadExit(dir); exited && !HostAlive(dir) {
			return nil, nil, wel, fmt.Errorf("host exited: %s", ex.Reason)
		}
		time.Sleep(wait)
		wait = min(2*wait, 100*time.Millisecond)
	}
}
