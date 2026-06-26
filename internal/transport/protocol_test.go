package transport

import "testing"

func TestParseTarget(t *testing.T) {
	cases := []struct {
		in        string
		peer      string
		repo      string
		wantErr   bool
		roundTrip string
	}{
		{in: "maciek-laptop/dotfiles", peer: "maciek-laptop", repo: "dotfiles", roundTrip: "maciek-laptop/dotfiles"},
		{in: "bartek-pc", peer: "bartek-pc", repo: "", roundTrip: "bartek-pc"},
		{in: "  kasia/repo  ", peer: "kasia", repo: "repo", roundTrip: "kasia/repo"},
		{in: "", wantErr: true},
		{in: "/repo", wantErr: true},
	}
	for _, c := range cases {
		got, err := ParseTarget(c.in)
		if c.wantErr {
			if err == nil {
				t.Errorf("ParseTarget(%q) expected error", c.in)
			}
			continue
		}
		if err != nil {
			t.Errorf("ParseTarget(%q) unexpected error: %v", c.in, err)
			continue
		}
		if got.Peer != c.peer || got.Repo != c.repo {
			t.Errorf("ParseTarget(%q) = %+v, want peer=%q repo=%q", c.in, got, c.peer, c.repo)
		}
		if got.String() != c.roundTrip {
			t.Errorf("String() = %q, want %q", got.String(), c.roundTrip)
		}
	}
}

func TestGitDaemonURL(t *testing.T) {
	got := gitDaemonURL("127.0.0.1", 9418, "/myrepo")
	want := "git://127.0.0.1:9418/myrepo"
	if got != want {
		t.Errorf("gitDaemonURL = %q, want %q", got, want)
	}
}
