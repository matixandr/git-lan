// Command git-lan is a zero-config peer-to-peer git collaboration tool for the
// local network. It is designed to be invoked as a git extension: `git lan ...`.
package main

import (
	"fmt"
	"os"

	"github.com/matixandr/git-lan/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
