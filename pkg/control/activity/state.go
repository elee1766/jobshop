package activity

import (
	"encoding/json"
	"log/slog"
	"os"
	"sort"
	"sync"
	"time"

	"github.com/elee1766/jobshop/pkg/rpc"
)

// File is the cache file name under the controller root.
const File = "activity.json"

// RepoPoll is the per-repo poll state kept across restarts.
type RepoPoll struct {
	EventsETag string `json:"events_etag,omitempty"`
	RunsETag   string `json:"runs_etag,omitempty"`
	// ETags holds one per per-resource endpoint read for a github-app
	// identity (JOB-108), keyed by endpoint name.
	ETags         map[string]string `json:"etags,omitempty"`
	DefaultBranch string            `json:"default_branch,omitempty"`
	Next          time.Time         `json:"next,omitzero"` // X-Poll-Interval
}

// State is the cache file: the bounded window of items plus the poll
// state that lets a restart keep its ETags. It is a rebuildable
// projection of external data, not a store table.
type State struct {
	Repos map[string]*RepoPoll `json:"repos"`
	Teams map[string]time.Time `json:"teams"` // last Linear poll time
	Items []rpc.ActivityItem   `json:"items"`
	seen  map[string]struct{}  // item ids
	errs  map[string]string    // last error logged per source
	mu    *sync.Mutex          // guards Items for List
	path  string
	// added counts items added since load, for Poll's result.
	added int
}

func loadState(path string, log *slog.Logger) *State {
	st := &State{
		Repos: map[string]*RepoPoll{}, Teams: map[string]time.Time{},
		seen: map[string]struct{}{}, errs: map[string]string{}, mu: &sync.Mutex{},
		path: path,
	}
	b, err := os.ReadFile(st.path)
	if err == nil {
		if err := json.Unmarshal(b, st); err != nil {
			log.Warn("activity: cache unreadable, starting empty", "file", st.path, "err", err)
		}
	}
	if st.Repos == nil {
		st.Repos = map[string]*RepoPoll{}
	}
	if st.Teams == nil {
		st.Teams = map[string]time.Time{}
	}
	for _, it := range st.Items {
		st.seen[it.ID] = struct{}{}
	}
	return st
}

// save writes the cache atomically: a temp file renamed over it.
func (s *State) save() error {
	s.mu.Lock()
	b, err := json.Marshal(s)
	s.mu.Unlock()
	if err != nil {
		return err
	}
	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, b, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, s.path)
}

// unseen returns the items not yet in the cache.
func (s *State) unseen(items []rpc.ActivityItem) []rpc.ActivityItem {
	s.mu.Lock()
	defer s.mu.Unlock()
	var out []rpc.ActivityItem
	for _, it := range items {
		if _, ok := s.seen[it.ID]; !ok {
			out = append(out, it)
		}
	}
	return out
}

// add merges new items and trims to the window and per-source bound.
// It returns how many were new.
func (s *State) add(items []rpc.ActivityItem, now time.Time, window time.Duration, maxPer int) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	n := 0
	for _, it := range items {
		if _, ok := s.seen[it.ID]; ok || now.Sub(it.At) > window {
			continue
		}
		s.seen[it.ID] = struct{}{}
		s.Items = append(s.Items, it)
		n++
	}
	s.added += n
	sort.SliceStable(s.Items, func(i, j int) bool { return s.Items[i].At.After(s.Items[j].At) })
	per := map[string]int{}
	kept := s.Items[:0]
	for _, it := range s.Items {
		src := it.Source + ":" + it.Repo
		if now.Sub(it.At) > window || per[src] >= maxPer {
			delete(s.seen, it.ID)
			continue
		}
		per[src]++
		kept = append(kept, it)
	}
	s.Items = kept
	return n
}
