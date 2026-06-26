package cmd

import (
	"fmt"

	"github.com/matixandr/git-lan/internal/display"
	"github.com/matixandr/git-lan/pkg/config"
	"github.com/spf13/cobra"
)

var flagDetectFonts bool

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Show configuration or re-run terminal detection",
	Long: `Without flags, config prints the resolved configuration and where it
lives on disk. With --detect-fonts it forgets the cached Nerd Fonts result for
the current terminal profile and probes again.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if flagDetectFonts {
			return runDetectFonts(cmd)
		}
		return runShowConfig(cmd)
	},
}

func runDetectFonts(cmd *cobra.Command) error {
	out := cmd.OutOrStdout()
	// Re-detection needs to know the profile name first for a friendly message.
	nerd, profile, err := display.RedetectCurrent()
	if err != nil {
		fmt.Fprintf(out, "Detecting Nerd Fonts support for profile %q...\n", profile)
		fmt.Fprintf(out, "%s could not probe: %v\n", display.Icons.Warning, err)
		return nil
	}
	fmt.Fprintf(out, "Detecting Nerd Fonts support for profile %q...\n", profile)
	if nerd {
		fmt.Fprintf(out, "%s Nerd Fonts supported - cached for this terminal profile.\n", display.Icons.Success)
	} else {
		fmt.Fprintf(out, "%s Nerd Fonts not supported - using fallback icons.\n", display.Icons.Error)
	}
	return nil
}

func runShowConfig(cmd *cobra.Command) error {
	out := cmd.OutOrStdout()
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	path, _ := config.ConfigPath()
	dir, _ := config.ConfigDir()

	th := display.Active
	fmt.Fprintln(out, th.Heading.Render("git-lan configuration"))
	fmt.Fprintf(out, "  config file:  %s\n", path)
	fmt.Fprintf(out, "  data dir:     %s\n", dir)
	fmt.Fprintf(out, "  display name: %s\n", orDefault(cfg.DisplayName, "(hostname)"))
	fmt.Fprintf(out, "  port:         %s\n", portString(cfg.Port))
	fmt.Fprintf(out, "  allow push:   %v\n", cfg.AllowPush)
	fmt.Fprintf(out, "  nerd fonts:   %s\n", nerdString(cfg.NerdFonts))
	return nil
}

func orDefault(s, def string) string {
	if s == "" {
		return def
	}
	return s
}

func portString(p int) string {
	if p == 0 {
		return "(auto)"
	}
	return fmt.Sprintf("%d", p)
}

func nerdString(p *bool) string {
	if p == nil {
		return "(auto-detect)"
	}
	return fmt.Sprintf("%v", *p)
}

func init() {
	configCmd.Flags().BoolVar(&flagDetectFonts, "detect-fonts", false, "re-probe Nerd Fonts support for this terminal")
	rootCmd.AddCommand(configCmd)
}
