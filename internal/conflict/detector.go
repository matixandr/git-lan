// Package conflict provides early warnings about overlapping work before a
// push, so two people editing the same repo on a LAN find out before they
// stomp on each other rather than after.
package conflict

import (
	"fmt"
	"sort"
)

// Report summarizes the risk of pushing into a peer who also has uncommitted
// work. git-lan keeps mDNS metadata minimal, so the remote side is known only
// by its modified-file count; when the peer also shares filenames (same LAN,
// same repo), the overlap set is reported precisely.
type Report struct {
	LocalDirty   []string // files we are about to push that are dirty locally
	PeerModified int      // count of files the peer reports as modified
	Overlap      []string // filenames dirty on both sides, when known
	PeerName     string
}

// Risky reports whether the push warrants a warning.
func (r Report) Risky() bool {
	return r.PeerModified > 0 && len(r.LocalDirty) > 0
}

// Lines renders the report as human-readable warning lines (empty if not risky).
func (r Report) Lines() []string {
	if !r.Risky() {
		return nil
	}
	var out []string
	out = append(out, fmt.Sprintf("%s has %d uncommitted file(s) while you push %d local change(s).",
		r.PeerName, r.PeerModified, len(r.LocalDirty)))
	if len(r.Overlap) > 0 {
		out = append(out, "Both of you have edited:")
		for _, f := range r.Overlap {
			out = append(out, "  - "+f)
		}
		out = append(out, "Pushing now may overwrite their work. Coordinate first.")
	} else {
		out = append(out, "Filenames are not shared over mDNS, so overlap can't be confirmed - check before pushing.")
	}
	return out
}

// Detect builds a report from the local dirty set, the peer's modified count,
// and (optionally) the peer's dirty filenames if they were obtained out of band.
func Detect(peerName string, localDirty, peerDirty []string, peerModified int) Report {
	r := Report{
		LocalDirty:   dedupSorted(localDirty),
		PeerModified: peerModified,
		PeerName:     peerName,
	}
	if len(peerDirty) > 0 {
		r.Overlap = intersect(localDirty, peerDirty)
		if peerModified == 0 {
			r.PeerModified = len(peerDirty)
		}
	}
	return r
}

func intersect(a, b []string) []string {
	set := make(map[string]struct{}, len(b))
	for _, x := range b {
		set[x] = struct{}{}
	}
	var out []string
	seen := map[string]struct{}{}
	for _, x := range a {
		if _, ok := set[x]; ok {
			if _, dup := seen[x]; !dup {
				out = append(out, x)
				seen[x] = struct{}{}
			}
		}
	}
	sort.Strings(out)
	return out
}

func dedupSorted(in []string) []string {
	seen := map[string]struct{}{}
	var out []string
	for _, x := range in {
		if _, ok := seen[x]; !ok {
			seen[x] = struct{}{}
			out = append(out, x)
		}
	}
	sort.Strings(out)
	return out
}
