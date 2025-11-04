package logic

import (
	"errors"
	"maps"
	"os"
	"path/filepath"
	"strings"
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

func getResponseHeaders(url string) map[string]string {
	headers := ThisServer.ResponseHeaders
	bAllow := false
	for _, p := range ThisServer.AllowIframePaths {
		if strings.HasPrefix(url, p) {
			bAllow = true
			break
		}
	}
	if !bAllow {
		headers = maps.Clone(headers)
		headers["Content-Security-Policy"] = "form-action 'self'; frame-ancestors 'self';"
	}
	return headers
}
