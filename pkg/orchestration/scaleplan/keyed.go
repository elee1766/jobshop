package scaleplan

import (
	"maps"
	"slices"
)

// ForKeys plans a pool keyed by repo (JOB-207). desired is the pool's
// total. ready holds the ready tickets per repo key, with an entry for
// every key the pool has, "" (the default sessions) included. sessions
// groups the open sessions by key. A group whose key is not in ready
// (dropped from the pool) wants none, so it is cordoned and drains.
//
// desired is split over the keys by demand (ready plus working
// sessions), one at a time in key order, so no repo starves while
// another has work queued. What is left (min_sessions, a stabilization
// hold) keeps the sessions each key has in rotation, then goes to the
// default key. Each key is then planned by ForPool. starts lists the key
// of each session to start, interleaved, so a host cap granting fewer
// starts than wanted still spreads them over the keys.
func ForKeys(desired int, ready map[string]int, sessions map[string][]Session) (plan Plan, starts []string) {
	keys := slices.Sorted(maps.Keys(ready))
	demand := map[string]int{}
	active := map[string]int{}
	for _, k := range keys {
		demand[k] = ready[k]
		for _, s := range sessions[k] {
			if !s.Idle {
				demand[k]++
			}
			if s.Cordon == "" {
				active[k]++
			}
		}
	}
	split := map[string]int{}
	left := desired
	fill := func(want map[string]int) {
		for left > 0 {
			gave := false
			for _, k := range keys {
				if left > 0 && split[k] < want[k] {
					split[k]++
					left--
					gave = true
				}
			}
			if !gave {
				return
			}
		}
	}
	fill(demand)
	fill(active)
	if _, ok := ready[""]; ok {
		split[""] += left
	}
	perKey := map[string]int{}
	all := slices.Sorted(maps.Keys(sessions))
	all = slices.Compact(slices.Sorted(slices.Values(append(all, keys...))))
	for _, k := range all {
		kp := ForPool(split[k], sessions[k])
		plan.Cordon = append(plan.Cordon, kp.Cordon...)
		plan.Uncordon = append(plan.Uncordon, kp.Uncordon...)
		plan.Start += kp.Start
		perKey[k] = kp.Start
	}
	for len(starts) < plan.Start {
		for _, k := range all {
			if perKey[k] > 0 {
				starts = append(starts, k)
				perKey[k]--
			}
		}
	}
	return plan, starts
}
