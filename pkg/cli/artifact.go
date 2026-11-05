package cli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/elee1766/jobshop/pkg/rpc"
)

// artifactCmd delivers and reads artifacts (#508).
type artifactCmd struct {
	Deliver artifactDeliverCmd `cmd:"" help:"Deliver a report for a human to read in Desk, with optional attachments."`
	List    artifactListCmd    `cmd:"" help:"List artifacts, unread first."`
	Show    artifactShowCmd    `cmd:"" help:"Show an artifact with its comments and history."`
	Ack     artifactAckCmd     `cmd:"" help:"Acknowledge (close) an artifact as read (admin)."`
}

type artifactDeliverCmd struct {
	Title  string   `arg:"" help:"Artifact title."`
	File   string   `short:"f" help:"Markdown body file; - or empty reads stdin."`
	Attach []string `short:"a" help:"File to attach (repeatable); must be inside the session dir or worktree." type:"path"`
	Ticket string   `short:"t" help:"Source ticket (default: the session's current ticket)."`
}

func (c artifactDeliverCmd) Run(g *Globals) error {
	var body []byte
	var err error
	if c.File == "" || c.File == "-" {
		body, err = io.ReadAll(os.Stdin)
	} else {
		body, err = os.ReadFile(c.File)
	}
	if err != nil {
		return err
	}
	files := make([]string, 0, len(c.Attach))
	for _, a := range c.Attach {
		abs, err := filepath.Abs(a)
		if err != nil {
			return err
		}
		files = append(files, abs)
	}
	return g.call(rpc.MethodArtifactDeliver, rpc.ArtifactDeliverParams{Title: c.Title, Body: string(body), Files: files, Ticket: c.Ticket}, nil)
}

type artifactListCmd struct {
	Unread  bool   `help:"Only unread (open) artifacts."`
	Trigger string `help:"Only artifacts from this trigger."`
	Pool    string `help:"Only artifacts from this pool."`
	JSON    bool   `name:"json" help:"Print JSON."`
}

func (c artifactListCmd) Run(g *Globals) error {
	p := rpc.ArtifactListParams{Unread: c.Unread, Trigger: c.Trigger, Pool: c.Pool}
	if c.JSON {
		return g.call(rpc.MethodArtifactList, p, nil)
	}
	var res []rpc.Artifact
	if err := g.call(rpc.MethodArtifactList, p, &res); err != nil {
		return err
	}
	w := tabwriter.NewWriter(stdout, 0, 2, 2, ' ', 0)
	_, _ = fmt.Fprintln(w, "ID\tSTATE\tSOURCE\tAGE\tFILES\tTITLE")
	for _, a := range res {
		state := "unread"
		if a.Read {
			state = "read"
		}
		src := a.Meta.Trigger
		if src == "" {
			src = a.Meta.Pool
		}
		age := time.Since(a.Ticket.CreatedAt).Round(time.Minute)
		_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%d\t%s\n", a.Ticket.ID, state, orDash(src), age, len(a.Meta.Files), a.Ticket.Title)
	}
	return w.Flush()
}

func orDash(s string) string {
	if strings.TrimSpace(s) == "" {
		return "-"
	}
	return s
}

type artifactShowCmd struct {
	ID string `arg:"" help:"Artifact ticket id."`
}

func (c artifactShowCmd) Run(g *Globals) error {
	return g.call(rpc.MethodArtifactShow, rpc.IDParams{ID: c.ID}, nil)
}

type artifactAckCmd struct {
	ID   string `arg:"" help:"Artifact ticket id."`
	Note string `short:"n" help:"Note, added as a comment."`
}

func (c artifactAckCmd) Run(g *Globals) error {
	return g.call(rpc.MethodArtifactAck, rpc.ArtifactAckParams{ID: c.ID, Note: c.Note}, nil)
}
