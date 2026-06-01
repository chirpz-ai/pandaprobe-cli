// Package version holds build metadata injected at link time via -ldflags.
package version

import "runtime"

// These variables are overridden at build time with:
//
//	-X github.com/chirpz-ai/pandaprobe-cli/internal/version.Version=...
//	-X github.com/chirpz-ai/pandaprobe-cli/internal/version.Commit=...
//	-X github.com/chirpz-ai/pandaprobe-cli/internal/version.Date=...
var (
	// Version is the semantic version of the build (e.g. "0.1.0").
	Version = "dev"
	// Commit is the short git commit the build was produced from.
	Commit = "none"
	// Date is the RFC3339 build timestamp.
	Date = "unknown"
)

// Info captures the full set of build and runtime metadata.
type Info struct {
	Version   string `json:"version"`
	Commit    string `json:"commit"`
	BuildDate string `json:"build_date"`
	GoVersion string `json:"go_version"`
	OS        string `json:"os"`
	Arch      string `json:"arch"`
}

// Get returns the current build information.
func Get() Info {
	return Info{
		Version:   Version,
		Commit:    Commit,
		BuildDate: Date,
		GoVersion: runtime.Version(),
		OS:        runtime.GOOS,
		Arch:      runtime.GOARCH,
	}
}

// UserAgent returns the value used for the HTTP User-Agent header.
func UserAgent() string {
	return "pandaprobe-cli/" + Version
}
