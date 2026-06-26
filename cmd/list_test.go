package cmd

import (
	"strings"
	"testing"

	"github.com/matixandr/git-lan/internal/display"
	"github.com/matixandr/git-lan/internal/session"
)

// withPlainDisplay swaps in the no-color theme and ASCII fallback icons so
// formatted rows are deterministic, restoring the originals afterward.
func withPlainDisplay(t *testing.T) {
	t.Helper()
	theme, icons := display.Active, display.Icons
	display.Active = display.NewTheme(false)
	display.Icons = display.FallbackIcons
	t.Cleanup(func() {
		display.Active = theme
		display.Icons = icons
	})
}

func TestFormatSelfSession(t *testing.T) {
	withPlainDisplay(t)

	tests := []struct {
		name        string
		sess        *session.Session
		wantContain []string
		wantAbsent  []string
	}{
		{
			name:        "locked and push-allowed",
			sess:        &session.Session{Name: "hackathon", PasswordHash: "x", AllowPush: true},
			wantContain: []string{"you", `hosting "hackathon"`, display.FallbackIcons.Lock, "(push allowed)", "(this host)"},
		},
		{
			name:        "open session, no push",
			sess:        &session.Session{Name: "demo"},
			wantContain: []string{`hosting "demo"`, "(this host)"},
			wantAbsent:  []string{display.FallbackIcons.Lock, "(push allowed)"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			row := formatSelfSession(tt.sess)
			for _, want := range tt.wantContain {
				if !strings.Contains(row, want) {
					t.Errorf("row %q missing %q", row, want)
				}
			}
			for _, absent := range tt.wantAbsent {
				if strings.Contains(row, absent) {
					t.Errorf("row %q unexpectedly contains %q", row, absent)
				}
			}
		})
	}
}
