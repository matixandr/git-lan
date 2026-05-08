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
			// Trust-ring persistence is added with the trust commands; for now
			// trusting behaves like accepting for this connection.
			return nil
		default:
			return fmt.Errorf("rejected by host")
		}
	}
}
