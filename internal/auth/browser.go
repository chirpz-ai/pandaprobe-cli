package auth

import (
	"fmt"
	"os/exec"
	"runtime"
)

// OpenBrowser opens url in the user's default browser. It is the default
// Options.Open; tests inject their own opener.
func OpenBrowser(url string) error {
	var cmd string
	var args []string
	switch runtime.GOOS {
	case "darwin":
		cmd, args = "open", []string{url}
	case "windows":
		cmd, args = "rundll32", []string{"url.dll,FileProtocolHandler", url}
	default: // linux, *bsd, etc.
		cmd, args = "xdg-open", []string{url}
	}
	if _, err := exec.LookPath(cmd); err != nil {
		return fmt.Errorf("cannot open browser (%s not found); re-run with --no-browser", cmd)
	}
	return exec.Command(cmd, args...).Start()
}
