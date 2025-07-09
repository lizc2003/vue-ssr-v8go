package logic

import (
	"errors"
	"os"
	"path/filepath"
)

const (
	PublicPath   = "public"
	ServerPath   = "server"
	IndexName    = "index.html"
	NotfoundName = "404.html"
	ManifestName = ".vite/ssr-manifest.json"
)

var (
	ErrorPageNotFound  = errors.New("page not found")
	ErrorPageRedirect  = errors.New("page redirect")
	ErrorSsrOff        = errors.New("ssr off")
	ErrorRenderTimeout = errors.New("render timeout")

	ForwardHeaders = []string{
		"Cookie",
		"User-Agent",
		"X-Forwarded-For",
	}
)

func getDistPath(distDir string) (string, error) {
	if distDir == "" {
		return "", errors.New("empty dist dir")
	}

	if distDir[0] != '/' {
		basepath, err := filepath.Abs(filepath.Dir(os.Args[0]))
		if err != nil {
			return "", err
		}
		distDir = basepath + "/" + distDir
	}
	if distDir[len(distDir)-1] != '/' {
		distDir += "/"
	}

	return distDir, nil
}
