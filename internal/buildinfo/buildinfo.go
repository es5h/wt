package buildinfo

import "runtime/debug"

const (
	// ModulePath is the canonical Go module path used for release installs.
	ModulePath = "github.com/es5h/wt"
)

// Version can be overridden at build time via -ldflags.
var Version = "dev"

// EffectiveVersion returns a user-facing version string.
//
// For local/dev builds, Version may remain "dev".
// For module installs (e.g. go install ...@latest), we infer the tagged
// module version from build info when ldflags are not provided.
func EffectiveVersion() string {
	if Version != "" && Version != "dev" {
		return Version
	}

	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "dev"
	}
	if v := info.Main.Version; v != "" && v != "(devel)" {
		return v
	}
	return "dev"
}
