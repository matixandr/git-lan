package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/matixandr/git-lan/internal/display"
	"github.com/matixandr/git-lan/internal/session"
	"github.com/spf13/cobra"
)

var flagInviteTTL time.Duration

var sessionInviteCmd = &cobra.Command{
	Use:   "invite",
	Short: "Mint a one-time join token for the active session",
	RunE: func(cmd *cobra.Command, args []string) error {
		store, err := session.Load()
		if err != nil {
			return err
		}
		if store.Active == nil {
			return fmt.Errorf("no active session - run `git lan session create` first")
		}
		token, _, err := session.GenerateInvite(store.Active.Secret, flagInviteTTL)
		if err != nil {
			return err
		}
		printInviteBlock(cmd, store.Active.Name, token, flagInviteTTL)
		return nil
	},
}

func printInviteBlock(cmd *cobra.Command, sessionName, token string, ttl time.Duration) {
	out := cmd.OutOrStdout()
	th := display.Active

	width := len(token) + 4
	if width < 40 {
		width = 40
	}
	bar := strings.Repeat("─", width-2)
	pad := func(s string) string {
		if len(s) > width-4 {
			return s
		}
		return s + strings.Repeat(" ", width-4-len(s))
	}

	fmt.Fprintln(out, th.Heading.Render(fmt.Sprintf("%s one-time invite for \"%s\"", display.Icons.Lock, sessionName)))
	fmt.Fprintln(out, "┌"+bar+"┐")
	fmt.Fprintln(out, "│ "+pad(token)+" │")
	fmt.Fprintln(out, "└"+bar+"┘")
	fmt.Fprintf(out, "valid for %s · single use · join with:\n", ttl)
	fmt.Fprintf(out, "  %s\n", th.Bold.Render("git lan session join <peer>/"+sessionName+" --token "+token))
}

func init() {
	sessionInviteCmd.Flags().DurationVar(&flagInviteTTL, "ttl", time.Hour, "how long the invite is valid")
	sessionCmd.AddCommand(sessionInviteCmd)
}
