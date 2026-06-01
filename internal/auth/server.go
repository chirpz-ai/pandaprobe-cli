package auth

import (
	"context"
	"crypto/subtle"
	"fmt"
	"net"
	"net/http"
)

// callbackServer is the loopback HTTP server that receives the browser redirect
// carrying the one-time code. It binds 127.0.0.1 only.
type callbackServer struct {
	state    string
	listener net.Listener
	srv      *http.Server
	result   chan callbackResult
}

type callbackResult struct {
	code string
	err  error
}

// newCallbackServer binds a loopback listener on an OS-chosen port and prepares
// the /callback handler. The caller must call start() then wait().
func newCallbackServer(state string) (*callbackServer, error) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, fmt.Errorf("start loopback server: %w", err)
	}
	cs := &callbackServer{
		state:    state,
		listener: ln,
		result:   make(chan callbackResult, 1),
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/callback", cs.handleCallback)
	cs.srv = &http.Server{Handler: mux}
	return cs, nil
}

// Port returns the chosen loopback port.
func (cs *callbackServer) Port() int {
	return cs.listener.Addr().(*net.TCPAddr).Port
}

func (cs *callbackServer) start() {
	go func() { _ = cs.srv.Serve(cs.listener) }()
}

func (cs *callbackServer) shutdown() {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_ = cs.srv.Shutdown(ctx)
}

func (cs *callbackServer) handleCallback(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	gotState := q.Get("state")
	code := q.Get("code")

	// Constant-time state comparison guards against a malicious local process
	// racing the real browser callback.
	if subtle.ConstantTimeCompare([]byte(gotState), []byte(cs.state)) != 1 {
		cs.fail(w, "state mismatch — the login attempt could not be verified.")
		cs.deliver(callbackResult{err: fmt.Errorf("state mismatch on callback")})
		return
	}
	if errParam := q.Get("error"); errParam != "" {
		msg := q.Get("error_description")
		if msg == "" {
			msg = errParam
		}
		cs.fail(w, "Login failed: "+msg)
		cs.deliver(callbackResult{err: fmt.Errorf("login failed: %s", msg)})
		return
	}
	if code == "" {
		cs.fail(w, "No authorization code was returned.")
		cs.deliver(callbackResult{err: fmt.Errorf("no code in callback")})
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(successPage))
	cs.deliver(callbackResult{code: code})
}

// deliver sends the first result and ignores any subsequent ones (buffered, 1).
func (cs *callbackServer) deliver(res callbackResult) {
	select {
	case cs.result <- res:
	default:
	}
}

func (cs *callbackServer) fail(w http.ResponseWriter, msg string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusBadRequest)
	_, _ = fmt.Fprintf(w, failPage, msg)
}

// wait blocks until the callback fires or ctx is cancelled (timeout/interrupt).
func (cs *callbackServer) wait(ctx context.Context) (string, error) {
	select {
	case res := <-cs.result:
		return res.code, res.err
	case <-ctx.Done():
		return "", ctx.Err()
	}
}

const successPage = `<!doctype html><html><head><meta charset="utf-8">
<title>PandaProbe CLI</title></head>
<body style="font-family:system-ui,sans-serif;text-align:center;padding:4rem">
<h1>&#10003; Logged in</h1>
<p>You can return to your terminal &mdash; this tab can be closed.</p>
</body></html>`

const failPage = `<!doctype html><html><head><meta charset="utf-8">
<title>PandaProbe CLI</title></head>
<body style="font-family:system-ui,sans-serif;text-align:center;padding:4rem">
<h1>Login failed</h1>
<p>%s</p>
<p>Return to your terminal and try again.</p>
</body></html>`
