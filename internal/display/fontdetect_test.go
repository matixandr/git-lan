package display

import (
	"strings"
	"testing"
)

func TestTerminalProfileKey(t *testing.T) {
	// Clear everything the function inspects so each case is isolated.
	vars := []string{
		"WT_PROFILE_ID", "WT_SESSION", "TERM_PROGRAM", "TERM",
		"VTE_VERSION", "PSModulePath",
	}
	for _, v := range vars {
		t.Setenv(v, "")
	}

	cases := []struct {
		name string
		set  map[string]string
		want string
	}{
		{"windows terminal", map[string]string{"WT_PROFILE_ID": "abc-guid"}, "windows-terminal:abc-guid"},
		{"vscode", map[string]string{"TERM_PROGRAM": "vscode"}, "vscode:"},
		{"wezterm", map[string]string{"TERM_PROGRAM": "WezTerm"}, "wezterm:"},
		{"iterm2", map[string]string{"TERM_PROGRAM": "iTerm.app"}, "iterm2:"},
		{"kitty", map[string]string{"TERM": "xterm-kitty"}, "kitty:"},
		{"alacritty", map[string]string{"TERM": "alacritty"}, "alacritty:"},
		{"gnome", map[string]string{"VTE_VERSION": "6800"}, "gnome-terminal:6800"},
		{"generic", map[string]string{"TERM": "xterm-256color"}, "xterm-256color:"},
		{"unknown", map[string]string{}, "unknown:"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			for _, v := range vars {
				t.Setenv(v, "")
			}
			for k, v := range c.set {
				t.Setenv(k, v)
			}
			if got := terminalProfileKey(); got != c.want {
				t.Errorf("terminalProfileKey() = %q, want %q", got, c.want)
			}
		})
	}
}

func TestProfileKeyPrecedenceWTBeatsTerm(t *testing.T) {
	t.Setenv("TERM", "xterm-256color")
	t.Setenv("WT_PROFILE_ID", "office-guid")
	got := terminalProfileKey()
	if !strings.HasPrefix(got, "windows-terminal:") {
		t.Errorf("WT_PROFILE_ID should win, got %q", got)
	}
}
