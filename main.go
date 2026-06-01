// Command pandaprobe is an agent-first CLI for the PandaProbe observability platform.
//
// It is designed to be driven by both humans and AI coding agents. By default
// every command emits JSON to stdout, errors as JSON to stderr, uses no
// interactive prompts, and returns meaningful exit codes.
package main

import (
	"os"

	"github.com/chirpz-ai/pandaprobe-cli/cmd"
)

func main() {
	os.Exit(int(cmd.Execute()))
}
