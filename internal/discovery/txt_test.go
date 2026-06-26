package discovery

import "testing"

func TestTXTRoundTrip(t *testing.T) {
	in := Advertisement{
		Repo:     "git-lan",
		Branch:   "main",
		Head:     "abc1234",
		Modified: 3,
		Session:  "hackathon",
		Locked:   true,
		Presence: PresenceCoding,
	}
	out := ParseTXT(in.TXT())
	if out.Repo != in.Repo || out.Branch != in.Branch || out.Head != in.Head {
		t.Fatalf("repo/branch/head mismatch: %+v", out)
	}
	if out.Modified != in.Modified || !out.Locked || out.Session != in.Session {
		t.Fatalf("metadata mismatch: %+v", out)
	}
	if out.Presence != PresenceCoding {
		t.Fatalf("presence = %q, want coding", out.Presence)
	}
}

func TestTXTTruncatesBranchAndHead(t *testing.T) {
	in := Advertisement{
		Branch: "feature/way-too-long-branch-name-here",
		Head:   "abcdef0123456789",
	}
	out := ParseTXT(in.TXT())
	if len(out.Branch) != maxBranchLen {
		t.Errorf("branch len = %d, want %d", len(out.Branch), maxBranchLen)
	}
	if len(out.Head) != maxHeadLen {
		t.Errorf("head len = %d, want %d", len(out.Head), maxHeadLen)
	}
}

func TestProtocolOf(t *testing.T) {
	if got := protocolOf([]string{"repo=x", "v=v1"}); got != "v1" {
		t.Errorf("protocolOf = %q, want v1", got)
	}
	if got := protocolOf([]string{"repo=x"}); got != "" {
		t.Errorf("protocolOf = %q, want empty", got)
	}
}
