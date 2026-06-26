package cmd

import (
	"fmt"

	"github.com/matixandr/git-lan/internal/display"
	"github.com/matixandr/git-lan/internal/session"
	"github.com/spf13/cobra"
)

var sessionLeaveCmd = &cobra.Command{
	Use:   "leave",
	Short: "Clear the active session state",
	Long: `Clears the persisted active session. If a session daemon is still
running in another terminal, stop it there with Ctrl+C - leave only removes the
saved state so a stale entry does not linger after a crash.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		store, err := session.Load()
		if err != nil {
			return err
		}
		out := cmd.OutOrStdout()
		if store.Active == nil {
			fmt.Fprintf(out, "%s no active session.\n", display.Icons.Offline)
			return nil
		}
		name := store.Active.Name
		store.Active = nil
		if err := store.Save(); err != nil {
			return err
		}
		fmt.Fprintf(out, "%s left session \"%s\".\n", display.Icons.Success, name)
		return nil
	},
}

func init() {
	sessionCmd.AddCommand(sessionLeaveCmd)
}
