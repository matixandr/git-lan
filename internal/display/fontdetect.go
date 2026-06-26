package display

import (
	"os"
)

// terminalProfileKey builds a stable identifier for the current terminal
// environment, used as the cache key for Nerd Fonts detection. The same
// physical terminal profile should always map to the same key so a probe runs
// at most once per profile.
//
// The returned key is "<type>:<discriminator>", e.g. "windows-terminal:<guid>",
// "iterm2:", "vscode:", or "xterm-256color:".
func terminalProfileKey() string {
	env := os.Getenv

	switch {
	case env("WT_PROFILE_ID") != "":
		// Windows Terminal assigns a stable GUID per profile - the ideal key.
		return "windows-terminal:" + env("WT_PROFILE_ID")
	case env("TERM_PROGRAM") == "vscode":
		return "vscode:"
	case env("TERM_PROGRAM") == "WezTerm":
		return "wezterm:"
	case env("TERM_PROGRAM") == "iTerm.app":
		return "iterm2:"
	case env("TERM") == "xterm-kitty":
		return "kitty:"
	case env("TERM") == "alacritty":
		return "alacritty:"
	case env("VTE_VERSION") != "":
		return "gnome-terminal:" + env("VTE_VERSION")
	case env("PSModulePath") != "" && env("WT_SESSION") == "":
		// PowerShell outside Windows Terminal (e.g. conhost).
		return "powershell:"
	default:
		if t := env("TERM"); t != "" {
			return t + ":"
		}
		return "unknown:"
	}
}
