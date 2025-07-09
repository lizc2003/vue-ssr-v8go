package logic

import (
	"encoding/json"
	"fmt"
	"github.com/lizc2003/vue-ssr-v8go/server/common/defs"
	"github.com/lizc2003/vue-ssr-v8go/server/common/tlog"
	"github.com/lizc2003/vue-ssr-v8go/server/common/util"
	"net/http"
	"os"
	"strings"
)

type IndexHtml struct {
	indexFileName    string
	indexHtml        string
	metaBegin        int
	metaEnd          int
	notfoundHtml     string
	manifestFileName string
	ssrManifest      map[string][]string
}

func NewIndexHtml(env string, publicDir string) (*IndexHtml, error) {
	indexFileName := publicDir + "/" + IndexName
	indexContent, err := os.ReadFile(indexFileName)
	if err != nil {
		return nil, err
	}

	var indexHtml string
	metaBegin := 0
	metaEnd := 0

	manifestFileName := publicDir + "/" + ManifestName
	var ssrManifest map[string][]string

	if env != defs.EnvDev {
		indexHtml = string(indexContent)
		metaBegin, metaEnd = getMetaPosition(indexHtml)

		ssrManifest = getRawManifest(manifestFileName)
		if ssrManifest == nil {
			ssrManifest = make(map[string][]string)
		}
	}

	var notfoundHtml string
	content, err := os.ReadFile(publicDir + "/" + NotfoundName)
	if err == nil {
		notfoundHtml = string(content)
	} else {
		notfoundHtml = `<!DOCTYPE html><html lang="en"><head></head><body><h1>Page Not Found</h1></body></html>`
	}

	return &IndexHtml{
		indexFileName:    indexFileName,
		indexHtml:        indexHtml,
		metaBegin:        metaBegin,
		metaEnd:          metaEnd,
		notfoundHtml:     notfoundHtml,
		manifestFileName: manifestFileName,
		ssrManifest:      ssrManifest,
	}, nil
}

func (this *IndexHtml) GetIndexHtml(result RenderResult, renderErr error) (int, string, error) {
	err := renderErr
	if err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "Error: 404 ") {
			return http.StatusNotFound, this.notfoundHtml, ErrorPageNotFound
		} else if idx := strings.Index(errMsg, "Error: 301 "); idx >= 0 {
			return 301, getRedirectUrl(errMsg, idx), ErrorPageRedirect
		} else if idx2 := strings.Index(errMsg, "Error: 302 "); idx2 >= 0 {
			return 302, getRedirectUrl(errMsg, idx2), ErrorPageRedirect
		} else if strings.Contains(errMsg, "Error: ssr-off") {
			err = ErrorSsrOff
		}
	}

	indexHtml, metaBegin, metaEnd := this.getRawIndexHtml()
	if renderErr == nil {
		if result.Meta != "" && metaBegin > 0 {
			s1 := indexHtml[:metaBegin]
			s2 := indexHtml[metaEnd:]
			var sb strings.Builder
			sb.Grow(len(s1) + len(s2) + len(result.Meta))
			sb.WriteString(s1)
			sb.WriteString(result.Meta)
			sb.WriteString(s2)
			indexHtml = sb.String()
		}
		if result.Modules != "" {
			preloadLinks := this.getPreloadLinks(result.Modules)
			if preloadLinks != "" {
				indexHtml = strings.Replace(indexHtml, "<!--preload-links-->", preloadLinks, 1)
			}
		}
		if result.State != "" {
			state := "window.__INITIAL_STATE__ = " + result.State
			indexHtml = strings.Replace(indexHtml, "<!--app-state-->", state, 1)
		}
		indexHtml = strings.Replace(indexHtml, "<!--app-html-->", result.Html, 1)
	}
	return http.StatusOK, indexHtml, err
}

func (this *IndexHtml) getRawIndexHtml() (string, int, int) {
	if this.indexHtml != "" {
		return this.indexHtml, this.metaBegin, this.metaEnd
	}

	content, err := os.ReadFile(this.indexFileName)
	if err != nil {
		tlog.Error(err)
		return "", 0, 0
	}
	indexHtml := util.UnsafeBytes2Str(content)
	metaBegin, metaEnd := getMetaPosition(indexHtml)
	return indexHtml, metaBegin, metaEnd
}

func getMetaPosition(indexHtml string) (int, int) {
	metaBegin := strings.Index(indexHtml, "<!--meta-begin-->")
	metaEnd := strings.Index(indexHtml[metaBegin+1:], "<!--meta-end-->")
	if metaBegin <= 0 || metaEnd < 0 {
		return 0, 0
	}

	metaEnd += metaBegin + 1
	return metaBegin, metaEnd + 15
}

func (this *IndexHtml) getSsrManifest() map[string][]string {
	if this.ssrManifest != nil {
		return this.ssrManifest
	}
	return getRawManifest(this.manifestFileName)
}

func getRawManifest(fileName string) map[string][]string {
	var manifest map[string][]string
	content, err := os.ReadFile(fileName)
	if err == nil {
		var mfs map[string]any
		err = json.Unmarshal(content, &mfs)
		if err == nil {
			manifest = make(map[string][]string, len(mfs))
			for k, v := range mfs {
				if arr, ok := v.([]any); ok {
					var files []string
					for _, v2 := range arr {
						if str, ok := v2.(string); ok {
							files = append(files, str)
						}
					}
					manifest[k] = files
				}
			}
		}
	}
	return manifest
}
func (this *IndexHtml) getPreloadLinks(_modules string) string {
	var modules []string
	err := json.Unmarshal(util.UnsafeStr2Bytes(_modules), &modules)
	if err != nil {
		tlog.Error(err)
		return ""
	}
	if len(modules) == 0 {
		return ""
	}

	manifest := this.getSsrManifest()
	if len(manifest) == 0 {
		return ""
	}

	var sb strings.Builder
	var seen = make(map[string]bool)
	for _, module := range modules {
		if files, ok := manifest[module]; ok {
			for _, file := range files {
				if !seen[file] {
					seen[file] = true
					if files2, ok := manifest[basename(file)]; ok {
						for _, depFile := range files2 {
							sb.WriteString(renderPreloadLink(depFile))
							seen[depFile] = true
						}
					}
					sb.WriteString(renderPreloadLink(file))
				}
			}
		}
	}
	return sb.String()
}

func renderPreloadLink(file string) string {
	idx := strings.LastIndex(file, ".")
	if idx <= 0 {
		return ""
	}

	ext := file[idx+1:]
	switch ext {
	case "js":
		return fmt.Sprintf(`<link rel="modulepreload" crossorigin href="%s">`, file)
	case "css":
		return fmt.Sprintf(`<link rel="stylesheet" href="%s">`, file)
	case "woff":
		return fmt.Sprintf(`<link rel="preload" href="%s" as="font" type="font/woff" crossorigin>`, file)
	case "woff2":
		return fmt.Sprintf(`<link rel="preload" href="%s" as="font" type="font/woff2" crossorigin>`, file)
	case "gif":
		return fmt.Sprintf(`<link rel="preload" href="%s" as="image" type="image/gif">`, file)
	case "jpg", "jpeg":
		return fmt.Sprintf(`<link rel="preload" href="%s" as="image" type="image/jpeg">`, file)
	case "png":
		return fmt.Sprintf(`<link rel="preload" href="%s" as="image" type="image/png">`, file)
	default:
		return ""
	}
}

func basename(str string) string {
	return str[strings.LastIndex(str, "/")+1:]
}

func getRedirectUrl(s string, idx int) string {
	url := s[idx+len("Error: 301 "):]
	if url == "" {
		url = "/"
	} else {
		url = strings.Split(url, "\n")[0]
	}
	return url
}
