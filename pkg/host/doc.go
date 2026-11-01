// Package host is the session host: it spawns one agent, proxies ACP between
// the controller and the agent, journals output and replays it on reconnect.
// It must not import store or control.
package host
