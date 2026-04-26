package display

import (
	"bufio"
	"os"
	"strconv"
	"strings"
	"time"

	"golang.org/x/term"
)

// probeTimeout bounds how long we wait for a terminal's cursor-position reply.
// Terminals that do not implement DSR simply never answer; we must not hang.
const probeTimeout = 200 * time.Millisecond

// probeNerdFonts measures the on-screen width of a Nerd Fonts glyph using the
// terminal's cursor-position report (DSR). A patched font renders the glyph as
// a single cell; an unpatched terminal either renders nothing (0) or a
// replacement that occupies two cells. Only single-width counts as supported.
//
// It returns false without side effects if stdout is not a TTY, if the terminal
// does not answer DSR within probeTimeout, or on any I/O error.
func probeNerdFonts() bool {
	if !term.IsTerminal(int(os.Stdout.Fd())) {
		return false
	}

	// Raw mode on stdin so the DSR reply arrives unbuffered and unechoed.
	old, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return false
	}
	defer term.Restore(int(os.Stdin.Fd()), old)

	reader := bufio.NewReader(os.Stdin)

	// Move to a known column and measure the cursor before the glyph.
	os.Stdout.WriteString("\x1b[s\x1b[1G")
	before, ok := cursorColumn(reader)
	if !ok {
		os.Stdout.WriteString("\x1b[u")
		return false
	}

	// Emit one Nerd Fonts glyph (branch icon) and measure again.
	os.Stdout.WriteString(NerdFontsIcons.Branch)
	after, ok := cursorColumn(reader)

	// Restore cursor and erase whatever we drew on this line.
	os.Stdout.WriteString("\x1b[u\x1b[1K")
	if !ok {
		return false
	}

	// Single-cell advance means the glyph rendered correctly.
	return after-before == 1
}

// cursorColumn issues a DSR query (ESC [6n) and parses the column from the
// terminal's "ESC [ row ; col R" reply, with a timeout.
func cursorColumn(r *bufio.Reader) (int, bool) {
	os.Stdout.WriteString("\x1b[6n")

	type result struct {
		col int
		ok  bool
	}
	ch := make(chan result, 1)
	go func() {
		resp, err := r.ReadString('R')
		if err != nil {
			ch <- result{0, false}
			return
		}
		col, ok := parseCursorReport(resp)
		ch <- result{col, ok}
	}()

	select {
	case res := <-ch:
		return res.col, res.ok
	case <-time.After(probeTimeout):
		return 0, false
	}
}

// parseCursorReport extracts the column from a DSR reply like "\x1b[24;13R".
func parseCursorReport(s string) (int, bool) {
	i := strings.IndexByte(s, '[')
	j := strings.IndexByte(s, 'R')
	if i < 0 || j < 0 || j < i {
		return 0, false
	}
	body := s[i+1 : j]
	parts := strings.Split(body, ";")
	if len(parts) != 2 {
		return 0, false
	}
	col, err := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err != nil {
		return 0, false
	}
	return col, true
}
