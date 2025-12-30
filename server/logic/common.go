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

	bAllowIframe := false
	bAllowSharedArray := false
	for _, p := range ThisServer.AllowIframePaths {
		if strings.HasPrefix(url, p) {
			bAllowIframe = true
			break
		}
	}
	for _, p := range ThisServer.AllowSharedArrayBufferPaths {
		if strings.HasPrefix(url, p) {
			bAllowSharedArray = true
			break
		}
	}

	if !bAllowIframe || bAllowSharedArray {
		headers = maps.Clone(headers)
	}

	if !bAllowIframe {
		headers["Content-Security-Policy"] = "form-action 'self'; frame-ancestors 'self';"
	}

	if bAllowSharedArray {
		headers["Cross-Origin-Opener-Policy"] = "same-origin"
		headers["Cross-Origin-Embedder-Policy"] = "credentialless"
	}
	return headers
}
