package cmd

import (
	"github.com/spf13/cobra"
)

// Global flags shared by every subcommand.
var (
	flagNoNerdFonts bool
	flagNoColor     bool
	flagVerbose     bool
)

var rootCmd = &cobra.Command{
	Use:   "lan",
	Short: "Zero-config peer-to-peer git collaboration on your LAN",
	Long: `git-lan lets you share git repositories with people on the same network
without a server, an account, or any configuration. Start a session, and
colleagues discover you over mDNS and can clone, push, and pull instantly.

All peer-to-peer traffic is end-to-end encrypted.`,
	SilenceUsage:  true,
	SilenceErrors: true,
	Version:       Version,
}

// Execute runs the root command. It is the single entry point called by main.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	pf := rootCmd.PersistentFlags()
	pf.BoolVar(&flagNoNerdFonts, "no-nerd-fonts", false, "force plain icons, skip Nerd Fonts detection")
	pf.BoolVar(&flagNoColor, "no-color", false, "disable colored output")
	pf.BoolVarP(&flagVerbose, "verbose", "v", false, "verbose logging")
}
