package logic

import (
	"os"
	"path/filepath"
	"time"
)

const (
	DistPath     = "/dist/"
	PublicPath   = "public"
	ServerPath   = "server"
	IndexName    = "index.html"
	NotfoundName = "404.html"
	ManifestName = ".vite/ssr-manifest.json"

	RenderTimeout = 15 * time.Second
)

var (
	ForwardHeaders = []string{
		"Cookie",
		"User-Agent",
		"X-Forwarded-For",
	}
)

func getDistPath() string {
	basepath, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		return "." + DistPath
	}
	return basepath + DistPath
}
