package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"

	"github.com/matixandr/git-lan/internal/display"
	"github.com/matixandr/git-lan/internal/security"
	"github.com/matixandr/git-lan/internal/transport"
	"golang.org/x/term"
)

// approveMu serializes interactive approval prompts so concurrent inbound
// connections do not interleave their questions on the terminal.
var approveMu sync.Mutex

// approvePeer returns a transport.VerifyFunc that decides whether to admit an
// inbound peer. The trust ring (wired in a later pass) short-circuits known
// peers; unknown peers trigger an interactive fingerprint prompt. On a
// non-interactive host, unknown peers are rejected - fail safe.
func approvePeer(out io.Writer) transport.VerifyFunc {
	return func(peerID []byte) error {
		fp := security.FingerprintOf(peerID)

		// Auto-accept peers already in the trust ring - no prompt needed.
		if ring, err := security.LoadTrust(); err == nil {
			if host, ok := ring.FingerprintTrusted(fp); ok {
				fmt.Fprintf(out, "%s peer %s (%s) is trusted - accepted.\n",
					display.Icons.Success, host, fp)
				return nil
			}
		}

		approveMu.Lock()
		defer approveMu.Unlock()

		if !term.IsTerminal(int(os.Stdin.Fd())) {
			return fmt.Errorf("unknown peer %s and no terminal to approve", fp)
		}

		th := display.Active
		fmt.Fprintln(out)
		fmt.Fprintf(out, "%s incoming peer wants to connect\n", th.Warning.Render(display.Icons.Warning))
		fmt.Fprintf(out, "  fingerprint: %s\n", th.Bold.Render(fp))
		fmt.Fprint(out, "  [a]ccept once / [r]eject / [t]rust always? ")

		reader := bufio.NewReader(os.Stdin)
		line, _ := reader.ReadString('\n')
		switch strings.ToLower(strings.TrimSpace(line)) {
		case "a", "accept":
			return nil
		case "t", "trust":
			persistTrust(out, fp)
			return nil
		default:
			return fmt.Errorf("rejected by host")
		}
	}
}

// persistTrust pins a fingerprint after the host chose "trust always". The
// inbound side has no reliable hostname, so we prompt for a label.
func persistTrust(out io.Writer, fp string) {
	fmt.Fprint(out, "  name this peer (for `trust list`): ")
	reader := bufio.NewReader(os.Stdin)
	line, _ := reader.ReadString('\n')
	name := strings.TrimSpace(line)
	if name == "" {
		name = "peer-" + fp[len(fp)-6:]
	}
	ring, err := security.LoadTrust()
	if err != nil {
		return
	}
	ring.Add(name, fp)
	_ = ring.Save()
}
