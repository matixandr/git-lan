package cmd

import (
	"fmt"

	"github.com/matixandr/git-lan/internal/display"
	"github.com/matixandr/git-lan/internal/security"
	"github.com/spf13/cobra"
)

var trustCmd = &cobra.Command{
	Use:   "trust",
	Short: "Manage the ring of trusted peer fingerprints",
}

var trustAddCmd = &cobra.Command{
	Use:   "add <hostname> <fingerprint>",
	Short: "Pin a hostname to a fingerprint",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		ring, err := security.LoadTrust()
		if err != nil {
			return err
		}
		host, fp := args[0], args[1]
		if existing, ok := ring.Get(host); ok && existing.Fingerprint != fp {
			fmt.Fprintf(cmd.OutOrStdout(), "%s overwriting existing pin for %s (was %s)\n",
				display.Icons.Warning, host, existing.Fingerprint)
		}
		ring.Add(host, fp)
		if err := ring.Save(); err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "%s trusted %s → %s\n", display.Icons.Success, host, fp)
		return nil
	},
}

var trustRemoveCmd = &cobra.Command{
	Use:   "remove <hostname>",
	Short: "Remove a pinned hostname",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ring, err := security.LoadTrust()
		if err != nil {
			return err
		}
		if !ring.Remove(args[0]) {
			return fmt.Errorf("%s is not in the trust ring", args[0])
		}
		if err := ring.Save(); err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "%s removed %s\n", display.Icons.Success, args[0])
		return nil
	},
}

var trustListCmd = &cobra.Command{
	Use:   "list",
	Short: "List trusted peers",
	RunE: func(cmd *cobra.Command, args []string) error {
		ring, err := security.LoadTrust()
		if err != nil {
			return err
		}
		out := cmd.OutOrStdout()
		entries := ring.List()
		if len(entries) == 0 {
			fmt.Fprintf(out, "%s no trusted peers yet.\n", display.Icons.Offline)
			return nil
		}
		th := display.Active
		for _, e := range entries {
			fmt.Fprintf(out, "%s %s\n   %s  (added %s)\n",
				th.Online.Render(display.Icons.Success),
				th.Bold.Render(e.Hostname),
				th.Muted.Render(e.Fingerprint),
				e.AddedAt.Format("2006-01-02"))
		}
		return nil
	},
}

func init() {
	trustCmd.AddCommand(trustAddCmd, trustRemoveCmd, trustListCmd)
	rootCmd.AddCommand(trustCmd)
}
