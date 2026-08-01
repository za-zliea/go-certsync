package frontend

import (
	"embed"
	"io/fs"
)

// fsys embeds every file under dist/. The //go:embed directive cannot use
// "..", so this file lives next to dist/ rather than in src/server.
//
//go:embed all:dist
var fsys embed.FS

// FS exposes the embedded frontend build output with the "dist/" prefix
// stripped, so files are served at the root (e.g. "index.html" instead of
// "dist/index.html").
var FS, _ = fs.Sub(fsys, "dist")
