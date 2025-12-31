package logic

import (
	"errors"
	"fmt"
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
		if MatchPath(url, p) {
			bAllowIframe = true
			break
		}
	}
	for _, p := range ThisServer.AllowSharedArrayBufferPaths {
		if MatchPath(url, p) {
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

func MatchPath(path, pattern string) bool {
	if strings.HasPrefix(path, pattern) {
		return true
	}

	// Special handling
	if pattern == "" {
		return true
	}
	if pattern == "*" && path == "/" {
		return false
	}

	// Preprocessing: merge consecutive **
	pattern = mergeDoubleStars(pattern)

	m, n := len(pattern), len(path)

	// dp[i][j] indicates whether pattern[0:i] matches path[0:j]
	dp := make([][]bool, m+1)
	for i := range dp {
		dp[i] = make([]bool, n+1)
	}

	// Initialization: empty pattern matches empty path
	dp[0][0] = true

	// Handle patterns starting with * or **
	for i := 1; i <= m; i++ {
		if pattern[i-1] == '*' {
			// Check if it's the second * of **
			if i > 1 && pattern[i-2] == '*' {
				dp[i][0] = dp[i-1][0] // Second * of **
			} else {
				dp[i][0] = dp[i-1][0] // Single *
			}
		}
	}

	// Fill DP table
	for i := 1; i <= m; i++ {
		for j := 1; j <= n; j++ {
			if pattern[i-1] == '*' {
				// Check if it's **
				isDoubleStar := i > 1 && pattern[i-2] == '*'
				if isDoubleStar {
					// ** can match any character sequence (including /)
					// Three cases:
					// 1. ** matches empty: dp[i-1][j] (skip **)
					// 2. ** matches one character and continues: dp[i][j-1]
					// 3. ** matches one character and ends: dp[i-1][j-1]
					dp[i][j] = dp[i-1][j] || dp[i][j-1] || dp[i-1][j-1]
				} else {
					// Single *: cannot match path separator
					if path[j-1] != '/' {
						// Three cases:
						// 1. * matches empty: dp[i-1][j]
						// 2. * matches one character and continues: dp[i][j-1]
						// 3. * matches one character and ends: dp[i-1][j-1]
						dp[i][j] = dp[i-1][j] || dp[i][j-1] || dp[i-1][j-1]
					} else {
						// * cannot match /, can only match empty
						dp[i][j] = dp[i-1][j]
					}
				}
			} else if pattern[i-1] == '?' {
				// ? cannot match path separator
				if path[j-1] != '/' {
					dp[i][j] = dp[i-1][j-1]
				}
			} else {
				// Ordinary characters must match exactly
				if pattern[i-1] == path[j-1] {
					dp[i][j] = dp[i-1][j-1]
				}
			}
		}
	}

	// Support prefix matching: success if pattern completely matches any prefix of the path
	for j := 0; j <= n; j++ {
		if dp[m][j] {
			return true
		}
	}

	return false
}

// Merge consecutive ** into a single **
func mergeDoubleStars(pattern string) string {
	var result strings.Builder
	i := 0
	sz := len(pattern)
	for i < sz {
		if pattern[i] == '*' && i+1 < sz && pattern[i+1] == '*' {
			result.WriteString("**")
			i += 2
			for i < sz && pattern[i] == '*' {
				i++
			}
		} else {
			result.WriteByte(pattern[i])
			i++
		}
	}
	return result.String()
}

func TestMatchPath() {
	testCases := []struct {
		pattern string
		path    string
		expect  bool
	}{
		// basic test
		{"*.go", "main.go", true},
		{"*.go", "test.java", false},
		{"test?.go", "test1.go", true},
		{"test?.go", "test.go", false},
		{"ab?c*d", "abYcZd", true},
		{"*", "/", false},

		// path test
		{"/src/", "/src", false},
		{"/src", "/src/main.go", true},
		{"/srx", "/src/main.go", false},
		{"/src/*.go", "/src/main.go", true},
		{"src/*.go", "lib/main.go", false},
		{"/*/test", "/lib/testX", true},

		// ** test
		{"src/**/*.go", "src/main.go", false},
		{"src/**/*.go", "src/a/b/c/main.go", true},
		{"**/*.go", "any/deep/path/main.go", true},

		// mixed test
		{"src/**/test_?.go", "src/a/b/test_1.go", true},
		{"**/*test*", "a/b/c/my_test.go", true},

		// boundary test
		{"", "", true},
		{"*", "", true},
		{"?", "", false},
		{"**", "", true},
	}

	fmt.Println("test:")
	for _, tc := range testCases {
		result := MatchPath(tc.path, tc.pattern)
		status := "✓"
		if result != tc.expect {
			status = "✗"
		}
		fmt.Printf("%s pattern: %-20s path: %-30s result: %v\n",
			status, tc.pattern, tc.path, result)
	}
}
