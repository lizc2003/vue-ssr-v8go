package logic

import (
	"errors"
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
	ErrorPageNotFound  = errors.New("page not found")
	ErrorSsrOff        = errors.New("ssr off")
	ErrorRenderTimeout = errors.New("render timeout")

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
