package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/matixandr/git-lan/internal/display"
	"github.com/matixandr/git-lan/internal/git"
	"github.com/matixandr/git-lan/internal/session"
	"github.com/spf13/cobra"
)

var (
	flagJoinPassword string
	flagJoinToken    string
)

var sessionJoinCmd = &cobra.Command{
	Use:   "join <peer>[/repo] [dir]",
	Short: "Join a session by cloning its repository",
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
		// Supply whichever credential the user gave. A locked session needs one
		// of them; an open session ignores both.
		client.Token = flagJoinToken
		if flagJoinPassword != "" {
			pw := flagJoinPassword
			client.DeriveSeed = func(salt []byte) []byte {
				return session.DeriveSeed(pw, salt)
			}
		}

		url, stop, err := client.Bridge(repo)
		if err != nil {
			return err
		}
		defer stop()

		out := cmd.OutOrStdout()
		if peer.Locked {
			fmt.Fprintf(out, "%s session is locked - authenticating...\n", display.Icons.Lock)
		}
		fmt.Fprintf(out, "%s joining %s/%s ...\n", display.Icons.Branch, peer.Instance, repo)
		if err := git.Clone(url, dest, cmd.OutOrStderr()); err != nil {
			return fmt.Errorf("join failed (wrong password/token or access denied?): %w", err)
		}
		fmt.Fprintf(out, "%s joined. cloned into %s\n", display.Icons.Success, dest)
		return nil
	},
}

func init() {
	sessionJoinCmd.Flags().StringVar(&flagJoinPassword, "password", "", "session password")
	sessionJoinCmd.Flags().StringVar(&flagJoinToken, "token", "", "one-time invite token")
	sessionCmd.AddCommand(sessionJoinCmd)
}
