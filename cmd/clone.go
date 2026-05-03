package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/matixandr/git-lan/internal/display"
	"github.com/matixandr/git-lan/internal/git"
	"github.com/spf13/cobra"
)

var cloneCmd = &cobra.Command{
	Use:   "clone <peer>[/repo] [dir]",
	Short: "Clone a repository from a peer over the encrypted transport",
	Args:  cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		peer, repo, err := resolvePeer(args[0], 3*time.Second)
		if err != nil {
			return err
		}

		dest := repo
		if len(args) == 2 {
			dest = args[1]
		}
		if _, err := os.Stat(dest); err == nil {
			return fmt.Errorf("destination %q already exists", dest)
		}

		client, err := clientForPeer(peer)
		if err != nil {
			return err
		}
		url, stop, err := client.Bridge(repo)
		if err != nil {
			return err
		}
		defer stop()

		fmt.Fprintf(cmd.OutOrStdout(), "%s cloning %s/%s into %s ...\n",
			display.Icons.Branch, peer.Instance, repo, filepath.Base(dest))
		if err := git.Clone(url, dest, cmd.OutOrStderr()); err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "%s done.\n", display.Icons.Success)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(cloneCmd)
}
