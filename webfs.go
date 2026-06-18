package exam_quiz

import (
	"embed"
	"io/fs"
	"log"
)

//go:embed web/dist
var webDist embed.FS

// GetDistFS returns the embedded web distribution filesystem
func GetDistFS() fs.FS {
	distFS, err := fs.Sub(webDist, "web/dist")
	if err != nil {
		log.Printf("warning: failed to get web dist fs: %v", err)
		return nil
	}
	return distFS
}
