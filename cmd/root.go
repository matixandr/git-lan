package cmd

import (
	"github.com/matixandr/git-lan/internal/display"
	"github.com/matixandr/git-lan/pkg/config"
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
	// PersistentPreRun initializes display state once, before any command
	// renders, resolving Nerd Fonts and color from flags + config.
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		cfg, _ := config.Load()

		nerd := display.DetectNerdFonts(display.Options{
			ForceOff:    flagNoNerdFonts,
			ConfigValue: display.ConfigBoolPtr(cfg),
		})
		display.UseIcons(nerd)
		display.Active = display.NewTheme(!flagNoColor)
	},
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
