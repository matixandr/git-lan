package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"time"

	"github.com/matixandr/git-lan/internal/discovery"
	"github.com/matixandr/git-lan/internal/display"
	"github.com/spf13/cobra"
)

var flagStatusInterval time.Duration

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Live dashboard of peers on the network (Ctrl+C to exit)",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runStatus(cmd)
	},
}

func runStatus(cmd *cobra.Command) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	svc, err := discovery.Start(ctx, 0, nil)
	if err != nil {
		return err
	}
	defer svc.Stop()

	out := cmd.OutOrStdout()
	ticker := time.NewTicker(flagStatusInterval)
	defer ticker.Stop()

	drawDashboard(out, svc.Peers())
	for {
		select {
		case <-ctx.Done():
			fmt.Fprintln(out, "\nbye.")
			return nil
		case <-ticker.C:
			drawDashboard(out, svc.Peers())
		}
	}
}

func drawDashboard(out io.Writer, peers []discovery.Peer) {
	th := display.Active
	// Clear screen, home cursor.
	fmt.Fprint(out, "\x1b[2J\x1b[H")

	header := fmt.Sprintf("%s git-lan  ·  %d peer(s)  ·  %s",
		display.Icons.Peer, len(peers), time.Now().Format("15:04:05"))
	fmt.Fprintln(out, th.Heading.Render(header))
	fmt.Fprintln(out, th.Muted.Render("refreshing every "+flagStatusInterval.String()+" - Ctrl+C to exit"))
	fmt.Fprintln(out)

	if len(peers) == 0 {
		fmt.Fprintf(out, "%s No peers visible right now.\n", display.Icons.Offline)
		return
	}
	for _, p := range peers {
		fmt.Fprintln(out, formatPeerRow(p))
		if p.Modified > 0 {
			fmt.Fprintf(out, "    %s %d modified file(s)\n",
				th.Coding.Render(display.Icons.Coding), p.Modified)
		}
	}
}

func init() {
	statusCmd.Flags().DurationVar(&flagStatusInterval, "interval", 5*time.Second, "dashboard refresh interval")
	rootCmd.AddCommand(statusCmd)
}
