package conflict

import "testing"

func TestReportNotRiskyWhenPeerClean(t *testing.T) {
	r := Detect("kasia", []string{"a.go"}, nil, 0)
	if r.Risky() {
		t.Error("should not be risky when peer has no modifications")
	}
	if r.Lines() != nil {
		t.Error("expected no warning lines")
	}
}

func TestReportRiskyCountOnly(t *testing.T) {
	r := Detect("bartek", []string{"a.go", "b.go"}, nil, 3)
	if !r.Risky() {
		t.Fatal("expected risky report")
	}
	lines := r.Lines()
	if len(lines) == 0 {
		t.Fatal("expected warning lines")
	}
}

func TestReportOverlapDetected(t *testing.T) {
	r := Detect("maciek",
		[]string{"main.go", "util.go"},
		[]string{"util.go", "readme.md"},
		0,
	)
	if len(r.Overlap) != 1 || r.Overlap[0] != "util.go" {
		t.Fatalf("expected overlap [util.go], got %v", r.Overlap)
	}
	if r.PeerModified != 2 {
		t.Errorf("peer modified should fall back to len(peerDirty)=2, got %d", r.PeerModified)
	}
}
