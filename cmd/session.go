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
	"github.com/matixandr/git-lan/internal/git"
	"github.com/matixandr/git-lan/internal/security"
	"github.com/matixandr/git-lan/internal/session"
	"github.com/matixandr/git-lan/internal/transport"
	"github.com/matixandr/git-lan/pkg/config"
	"github.com/spf13/cobra"
)

var (
	flagSessionName     string
	flagSessionPassword string
	flagSessionAllow    bool
)

var sessionCmd = &cobra.Command{
	Use:   "session",
	Short: "Create, join, and manage collaboration sessions",
}

var sessionCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Start sharing the current repository (runs until Ctrl+C)",
	RunE:  func(cmd *cobra.Command, args []string) error { return runSessionCreate(cmd) },
}

func runSessionCreate(cmd *cobra.Command) error {
	out := cmd.OutOrStdout()
	repo, err := git.Open("")
	if err != nil {
		return err
	}
	name := flagSessionName
	if name == "" {
		name = repo.Name()
	}

	sess, err := session.New(name, flagSessionPassword, repo.Root, flagSessionAllow)
	if err != nil {
		return err
	}
	id, err := security.LoadOrCreateIdentity()
	if err != nil {
		return err
	}

	// Encrypted transport.
	srv := transport.NewServer(id.Private(), repo.Root, flagSessionAllow)
	if flagVerbose {
		srv.Log = func(f string, a ...any) { fmt.Fprintf(os.Stderr, "[server] "+f+"\n", a...) }
	}
	srv.Verify = approvePeer(out)
	if sess.HasPassword() {
		// Derive the gate seed in-process from the password just supplied; it
		// is never written to disk.
		srv.RequireAuth = true
		srv.Salt = sess.Salt
		srv.Seed = session.DeriveSeed(flagSessionPassword, sess.Salt)
	}

	cfg, _ := config.Load()
	if err := srv.Listen(cfg.Port); err != nil {
		return err
	}
	sess.Port = srv.Port()

	// Persist as the active session.
	store := &session.Store{Active: sess}
	if err := store.Save(); err != nil {
		return err
	}
	defer clearActiveSession()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	go func() { _ = srv.Serve(ctx) }()
	defer srv.Shutdown()

	// mDNS announce + browse.
	ad := advertisementFor(repo, name, sess.HasPassword())
	svc, err := discovery.Start(ctx, srv.Port(), &ad)
	if err != nil {
		return err
	}
	defer svc.Stop()

	printSessionBanner(out, sess, id, srv.Port())
	heartbeat(ctx, svc, repo, name, sess.HasPassword())
	fmt.Fprintln(out, "\nsession ended.")
	return nil
}

// heartbeat refreshes the advertised metadata so peers see live branch, HEAD,
// modified-file count, and presence until the context is canceled.
func heartbeat(ctx context.Context, svc *discovery.Service, repo *git.Repo, name string, locked bool) {
	t := time.NewTicker(20 * time.Second)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			svc.UpdateAdvertisement(advertisementFor(repo, name, locked))
		}
	}
}

func advertisementFor(repo *git.Repo, name string, locked bool) discovery.Advertisement {
	branch, _ := repo.Branch()
	presence := discovery.PresenceOnline
	if !repo.IsClean() {
		presence = discovery.PresenceCoding
	}
	return discovery.Advertisement{
		Repo:     repo.Name(),
		Branch:   branch,
		Head:     repo.Head(),
		Modified: repo.ModifiedCount(),
		Session:  name,
		Locked:   locked,
		Presence: presence,
	}
}

func printSessionBanner(out io.Writer, sess *session.Session, id *security.Identity, port int) {
	th := display.Active
	lock := ""
	if sess.HasPassword() {
		lock = " " + display.Icons.Lock
	}
	fmt.Fprintln(out, th.Heading.Render(fmt.Sprintf("%s session \"%s\"%s is live", display.Icons.Success, sess.Name, lock)))
	fmt.Fprintf(out, "  repo:        %s\n", sess.RepoRoot)
	fmt.Fprintf(out, "  port:        %d (encrypted)\n", port)
	fmt.Fprintf(out, "  push:        %v\n", sess.AllowPush)
	fmt.Fprintf(out, "  fingerprint: %s\n", th.Muted.Render(id.Fingerprint()))
	fmt.Fprintln(out, th.Muted.Render("  Ctrl+C to stop. `git lan session invite` to mint a join token."))
}

func clearActiveSession() {
	store, err := session.Load()
	if err != nil {
		return
	}
	store.Active = nil
	_ = store.Save()
}

func init() {
	sessionCreateCmd.Flags().StringVar(&flagSessionName, "name", "", "session name (defaults to repo name)")
	sessionCreateCmd.Flags().StringVar(&flagSessionPassword, "password", "", "require a password to join")
	sessionCreateCmd.Flags().BoolVar(&flagSessionAllow, "allow-push", false, "allow peers to push")
	sessionCmd.AddCommand(sessionCreateCmd)
	rootCmd.AddCommand(sessionCmd)
}
