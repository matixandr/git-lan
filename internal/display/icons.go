// Package display handles everything the user sees: icon sets, Nerd Fonts
// auto-detection (cached per terminal profile), and lipgloss theming.
package display

// IconSet is the full set of glyphs git-lan renders. Two concrete sets exist:
// a Nerd Fonts set using patched-font private-use codepoints, and an ASCII/
// Unicode fallback that renders anywhere. All display code reads from the
// active set via the package-level Icons, never hard-coding a glyph.
type IconSet struct {
	Online  string // peer reachable, clean tree
	Offline string // peer not responding
	Warning string // conflict / MITM / error condition
	Success string // operation completed
	Error   string // hard failure
	Lock    string // session has a password
	Branch  string // git branch indicator
	Commit  string // git commit indicator
	Coding  string // peer has uncommitted changes
	Idle    string // peer online but inactive
	Peer    string // peer / server generic
}

// nf builds a single-glyph string from a Nerd Fonts private-use codepoint.
// Using rune values keeps this source file pure ASCII, so the glyphs cannot be
// corrupted by an editor that lacks a patched font.
func nf(codepoint rune) string { return string(codepoint) }

// NerdFontsIcons uses glyphs from patched Nerd Fonts.
var NerdFontsIcons = IconSet{
	Online:  nf(0xf111), // nf-fa-circle
	Offline: nf(0xf1db), // nf-fa-circle_o
	Warning: nf(0xf071), // nf-fa-warning
	Success: nf(0xf00c), // nf-fa-check
	Error:   nf(0xf00d), // nf-fa-times
	Lock:    nf(0xf023), // nf-fa-lock
	Branch:  nf(0xe0a0), // nf-pl-branch
	Commit:  nf(0xf417), // nf-oct-git_commit
	Coding:  nf(0xf121), // nf-fa-code
	Idle:    nf(0xf017), // nf-fa-clock_o
	Peer:    nf(0xf233), // nf-fa-server
}

// FallbackIcons renders on any terminal without a patched font, using only
// widely-supported Unicode and ASCII.
var FallbackIcons = IconSet{
	Online:  nf(0x25cf), // filled circle
	Offline: nf(0x25cb), // hollow circle
	Warning: nf(0x26a0), // warning sign
	Success: nf(0x2713), // check mark
	Error:   nf(0x2717), // ballot x
	Lock:    "[locked]",
	Branch:  "#",
	Commit:  "*",
	Coding:  "~",
	Idle:    "-",
	Peer:    ">",
}

// Icons is the active icon set, initialized once at startup by the root
// command's pre-run after Nerd Fonts detection. It defaults to the safe
// fallback set so code paths that run before initialization still render.
var Icons = FallbackIcons

// UseIcons selects the active set. nerd true selects the Nerd Fonts set.
func UseIcons(nerd bool) {
	if nerd {
		Icons = NerdFontsIcons
	} else {
		Icons = FallbackIcons
	}
}
