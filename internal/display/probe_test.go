package display

import "testing"

func TestParseCursorReport(t *testing.T) {
	cases := []struct {
		in     string
		col    int
		wantOK bool
	}{
		{"\x1b[24;13R", 13, true},
		{"\x1b[1;1R", 1, true},
		{"\x1b[10;200R", 200, true},
		{"garbage", 0, false},
		{"\x1b[24R", 0, false},  // missing column
		{"\x1b[a;bR", 0, false}, // non-numeric
	}
	for _, c := range cases {
		col, ok := parseCursorReport(c.in)
		if ok != c.wantOK || (ok && col != c.col) {
			t.Errorf("parseCursorReport(%q) = (%d,%v), want (%d,%v)", c.in, col, ok, c.col, c.wantOK)
		}
	}
}
