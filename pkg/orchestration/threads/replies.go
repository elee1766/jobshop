package threads

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gfx-labs/jobshop/pkg/domain"
)

// Replies answers a thread.reply wait (waitengine.ReplySource): a human
// comment on the thread after the wait started, from From when set. It
// reads through the subscription source, so the same filters (bots,
// ignore, authors) apply; they live in the controller's config, which is
// why this is not wait engine code.
type Replies struct {
	E   *Engine
	Src Source
}

// Reply implements waitengine.ReplySource.
func (r Replies) Reply(ctx context.Context, cd domain.WaitCond, since time.Time) (bool, string, error) {
	th := Thread{Linear: cd.Thread}
	if repo, num, ok := strings.Cut(cd.Thread, "#"); ok {
		n, _ := strconv.Atoi(num)
		th = Thread{Repo: repo, Number: n, PR: true}
	}
	cs, err := r.Src.Comments(ctx, th, since)
	if err != nil {
		return false, "", err
	}
	for _, cm := range cs {
		if !cm.At.After(since) || r.E.Ignored(cm) {
			continue
		}
		if cd.From != "" && !strings.EqualFold(cd.From, cm.Author) && (!cm.Linear || !r.E.LinearMatch(cd.From, cm)) {
			continue
		}
		return true, fmt.Sprintf("%s replied on %s", cm.Author, cd.Thread), nil
	}
	return false, "", nil
}
