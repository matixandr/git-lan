package cmd

import (
	"fmt"
	"time"

	"github.com/matixandr/git-lan/internal/discovery"
	"github.com/matixandr/git-lan/internal/display"
	"github.com/matixandr/git-lan/internal/session"
	"github.com/spf13/cobra"
)

var flagListTimeout time.Duration

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List git-lan peers on the local network",
	RunE: func(cmd *cobra.Command, args []string) error {
		peers, err := browseFor(flagListTimeout)
		if err != nil {
			return err
		}
		renderPeerList(cmd, peers)
		return nil
	},
}

func renderPeerList(cmd *cobra.Command, peers []discovery.Peer) {
	out := cmd.OutOrStdout()
	if self := activeSession(); self != nil {
		fmt.Fprintln(out, formatSelfSession(self))
	}
	if len(peers) == 0 {
		fmt.Fprintf(out, "%s No peers found on the LAN.\n", display.Icons.Offline)
		return
	}
	for _, p := range peers {
		fmt.Fprintln(out, formatPeerRow(p))
	}
}

// activeSession returns this host's live hosted session, or nil if none is
// running (or the store cannot be read). `list` and `status` filter the host
// out of mDNS results - a host never sees itself - so reading the local store
// is the only way to confirm to the operator that their own session is live
// and advertised on the LAN.
func activeSession() *session.Session {
	store, err := session.Load()
	if err != nil {
		return nil
	}
	return store.Active
}

// formatSelfSession renders the row for this host's own active session, e.g.
//
//	● you              hosting "hackathon" [locked]  (this host)
func formatSelfSession(s *session.Session) string {
	th := display.Active
	name := fmt.Sprintf("%-16s", "you")

	tag := fmt.Sprintf("hosting %q", s.Name)
	if s.HasPassword() {
		tag += " " + display.Icons.Lock
	}
	if s.AllowPush {
		tag += " (push allowed)"
	}

	return fmt.Sprintf("%s %s %s %s",
		th.Online.Render(display.Icons.Online),
		th.Bold.Render(name),
		th.Warning.Render(tag),
		th.Muted.Render("(this host)"),
	)
}

// formatPeerRow renders one peer line, e.g.
//
//	● maciek-laptop   main      abc1234   2 min ago   [locked] hackathon
func formatPeerRow(p discovery.Peer) string {
	th := display.Active
	pres := p.Presence()
	icon, style := presenceGlyph(pres)

	name := fmt.Sprintf("%-16s", p.Instance)
	if pres == discovery.PresenceOffline {
		return fmt.Sprintf("%s %s %s",
			style.Render(icon), style.Render(name),
			th.Muted.Render("offline - last seen "+humanizeSince(p.LastSeen)))
	}

	branch := fmt.Sprintf("%s %-12s", display.Icons.Branch, truncate(p.Branch, 12))
	head := fmt.Sprintf("%-7s", p.Head)
	seen := fmt.Sprintf("%-10s", humanizeSince(p.LastSeen))

	row := fmt.Sprintf("%s %s %s %s %s",
		style.Render(icon),
		th.Bold.Render(name),
		th.Muted.Render(branch),
		head,
		th.Muted.Render(seen),
	)
	if p.Session != "" {
		tag := p.Session
		if p.Locked {
			tag = display.Icons.Lock + " " + tag
		}
		row += "  " + th.Warning.Render(tag)
	}
	return row
}

// presenceGlyph maps a presence to its icon and style.
func presenceGlyph(p discovery.Presence) (icon string, style interface{ Render(...string) string }) {
	th := display.Active
	switch p {
	case discovery.PresenceCoding:
		return display.Icons.Coding, th.Coding
	case discovery.PresenceIdle:
		return display.Icons.Idle, th.Idle
	case discovery.PresenceOffline:
		return display.Icons.Offline, th.Offline
	default:
		return display.Icons.Online, th.Online
	}
}

func truncate(s string, n int) string {
	if len(s) > n {
		return s[:n]
	}
	return s
}

// humanizeSince renders a coarse, friendly elapsed time.
func humanizeSince(t time.Time) string {
	if t.IsZero() {
		return "unknown"
	}
	d := time.Since(t)
	switch {
	case d < 10*time.Second:
		return "just now"
	case d < time.Minute:
		return fmt.Sprintf("%d sec ago", int(d.Seconds()))
	case d < time.Hour:
		return fmt.Sprintf("%d min ago", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%d hr ago", int(d.Hours()))
	default:
		return fmt.Sprintf("%d days ago", int(d.Hours()/24))
	}
}

func init() {
	listCmd.Flags().DurationVar(&flagListTimeout, "timeout", 2*time.Second, "how long to browse for peers")
	rootCmd.AddCommand(listCmd)
}
