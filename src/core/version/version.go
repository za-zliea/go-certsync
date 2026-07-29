// Package version holds the build version string injected via -ldflags.
package version

// Version is overwritten at build time via:
//   -ldflags "-X certsync/src/core/version.Version=<buildTime>-<gitShort>"
var Version = "dev"
