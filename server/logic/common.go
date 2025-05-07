package logic

import (
	"os"
	"path/filepath"
	"time"
)

const (
	DistPath     = "/dist/"
	ClientPath   = "public"
	ServerPath   = "server"
	IndexName    = "index.html"
	ManifestName = ".vite/ssr-manifest.json"

	RenderTimeout = 30 * time.Second
)

func getDistPath() string {
	basepath, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		return "." + DistPath
	}
	return basepath + DistPath
}

func getSsrHeaders(headers []string) []string {
	needHeaders := []string{
		"Cookie",
		"Authorization",
		"User-Agent",
		"X-Forwarded-For",
	}
	var extraHeaders []string

	for _, h := range headers {
		found := false
		for _, h2 := range needHeaders {
			if h == h2 {
				found = true
				break
			}
		}
		if !found {
			extraHeaders = append(extraHeaders, h)
		}
	}
	return append(needHeaders, extraHeaders...)
}
