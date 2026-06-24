package exam_quiz

import (
	"embed"
	"io/fs"
	"log/slog"
)

//go:embed web/dist
var webDist embed.FS

// GetDistFS returns the embedded web distribution filesystem
func GetDistFS() fs.FS {
	distFS, err := fs.Sub(webDist, "web/dist")
	if err != nil {
		slog.Warn("failed to get web dist fs", "error", err)
		return nil
	}
	return distFS
}
