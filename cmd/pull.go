package cmd

import (
	"fmt"
	"time"

	"github.com/matixandr/git-lan/internal/display"
	"github.com/matixandr/git-lan/internal/git"
	"github.com/spf13/cobra"
)

var pullCmd = &cobra.Command{
	Use:   "pull <peer>[/repo] [branch]",
	Short: "Pull from a peer into the current repository",
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

		client, err := clientForPeer(peer)
		if err != nil {
			return err
		}
		url, stop, err := client.Bridge(remoteRepo)
		if err != nil {
			return err
		}
		defer stop()

		fmt.Fprintf(cmd.OutOrStdout(), "%s pulling from %s/%s ...\n",
			display.Icons.Branch, peer.Instance, remoteRepo)
		if err := repo.Pull(url, branch, cmd.OutOrStderr()); err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "%s up to date.\n", display.Icons.Success)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(pullCmd)
}
