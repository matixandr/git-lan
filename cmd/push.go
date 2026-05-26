package cmd

import (
	"fmt"
	"time"

	"github.com/matixandr/git-lan/internal/conflict"
	"github.com/matixandr/git-lan/internal/discovery"
	"github.com/matixandr/git-lan/internal/display"
	"github.com/matixandr/git-lan/internal/git"
	"github.com/spf13/cobra"
)

var pushCmd = &cobra.Command{
	Use:   "push <peer>[/repo] [branch]",
	Short: "Push the current repository to a peer (peer must approve)",
	Args:  cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		repo, err := git.Open("")
		if err != nil {
			return err
		}
		peer, remoteRepo, err := resolvePeer(args[0], 3*time.Second)
		if err != nil {
			return err
		}
		branch := ""
		if len(args) == 2 {
			branch = args[1]
		}

		// Warn about overlapping dirty files before sending anything (a fuller
		// conflict check arrives with the conflict detector).
		warnOnConflicts(cmd, repo, peer)

		client, err := clientForPeer(peer)
		if err != nil {
			return err
		}
		url, stop, err := client.Bridge(remoteRepo)
		if err != nil {
			return err
		}
		defer stop()

		fmt.Fprintf(cmd.OutOrStdout(), "%s pushing to %s/%s ...\n",
			display.Icons.Branch, peer.Instance, remoteRepo)
		if err := repo.Push(url, branch, cmd.OutOrStderr()); err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "%s pushed.\n", display.Icons.Success)
		return nil
	},
}

// warnOnConflicts prints an early warning if both sides have uncommitted work.
func warnOnConflicts(cmd *cobra.Command, repo *git.Repo, peer discovery.Peer) {
	localDirty, err := repo.ModifiedFiles()
	if err != nil {
		return
	}
	report := conflict.Detect(peer.Instance, localDirty, nil, peer.Modified)
	if !report.Risky() {
		return
	}
	out := cmd.OutOrStdout()
	th := display.Active
	fmt.Fprintf(out, "%s possible conflict\n", th.Warning.Render(display.Icons.Warning))
	for _, line := range report.Lines() {
		fmt.Fprintf(out, "  %s\n", th.Warning.Render(line))
	}
}

func init() {
	rootCmd.AddCommand(pushCmd)
}
