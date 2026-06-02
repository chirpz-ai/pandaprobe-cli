package auth

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"strconv"
	"time"

	"github.com/chirpz-ai/pandaprobe-cli/internal/models"
)

// DefaultTimeout bounds how long Login waits for the browser handshake.
const DefaultTimeout = 3 * time.Minute

// Exchanger swaps a one-time code + PKCE verifier for credentials. *api.Client
// satisfies it; tests provide a stub.
type Exchanger interface {
	ExchangeCode(ctx context.Context, code, verifier string) (*models.CLICredentials, error)
}

// Options configures the login flow.
type Options struct {
	AuthURL    string // web app base, e.g. https://app.pandaprobe.com
	Label      string // human label for the minted key (e.g. hostname)
	CLIVersion string
	NoBrowser  bool
	Timeout    time.Duration
	Open       func(string) error // defaults to OpenBrowser; injected in tests
	Progress   io.Writer          // human progress messages (stderr); nil → discard
}

// Login runs the PKCE browser handshake and returns the minted credentials.
func Login(ctx context.Context, client Exchanger, opts Options) (*models.CLICredentials, error) {
	progress := opts.Progress
	if progress == nil {
		progress = io.Discard
	}
	// say writes a best-effort progress line (errors on a progress stream are
	// not actionable).
	say := func(format string, a ...any) { _, _ = fmt.Fprintf(progress, format, a...) }
	open := opts.Open
	if open == nil {
		open = OpenBrowser
	}
	timeout := opts.Timeout
	if timeout <= 0 {
		timeout = DefaultTimeout
	}

	state, err := RandomState()
	if err != nil {
		return nil, err
	}
	verifier, err := RandomVerifier()
	if err != nil {
		return nil, err
	}
	challenge := Challenge(verifier)

	srv, err := newCallbackServer(state)
	if err != nil {
		return nil, err
	}
	srv.start()
	defer srv.shutdown()

	loginURL, err := buildLoginURL(opts.AuthURL, srv.Port(), state, challenge, opts.Label, opts.CLIVersion)
	if err != nil {
		return nil, err
	}

	if opts.NoBrowser {
		say("Open this URL in your browser to continue:\n\n    %s\n\n", loginURL)
	} else {
		say("Opening your browser to complete login…\n")
		if oerr := open(loginURL); oerr != nil {
			say("Could not open a browser automatically (%v).\nOpen this URL manually:\n\n    %s\n\n", oerr, loginURL)
		} else {
			say("If it didn't open, visit:\n\n    %s\n\n", loginURL)
		}
	}
	say("Waiting for authorization…\n")

	waitCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	code, err := srv.wait(waitCtx)
	if err != nil {
		if waitCtx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("timed out waiting for browser authorization after %s", timeout)
		}
		return nil, err
	}

	creds, err := client.ExchangeCode(ctx, code, verifier)
	if err != nil {
		return nil, err
	}
	return creds, nil
}

func buildLoginURL(authURL string, port int, state, challenge, label, cliVersion string) (string, error) {
	u, err := url.Parse(authURL)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return "", fmt.Errorf("invalid auth URL %q", authURL)
	}
	u.Path = "/cli-login"
	q := url.Values{}
	q.Set("port", strconv.Itoa(port))
	q.Set("state", state)
	q.Set("code_challenge", challenge)
	q.Set("code_challenge_method", "S256")
	if label != "" {
		q.Set("label", label)
	}
	if cliVersion != "" {
		q.Set("cli_version", cliVersion)
	}
	u.RawQuery = q.Encode()
	return u.String(), nil
}
