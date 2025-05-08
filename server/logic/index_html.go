package logic

import (
	"github.com/lizc2003/vue-ssr-v8go/server/common/defs"
	"github.com/lizc2003/vue-ssr-v8go/server/common/tlog"
	"github.com/lizc2003/vue-ssr-v8go/server/common/util"
	"net/http"
	"os"
	"strings"
)

type IndexHtml struct {
	indexFileName string
	indexHtml     string
	metaBegin     int
	metaEnd       int
	notfoundHtml  string
	SsrManifest   string
}

func NewIndexHtml(env string, publicDir string) (*IndexHtml, error) {
	var indexHtml string
	metaBegin := 0
	metaEnd := 0
	indexFileName := publicDir + "/" + IndexName
	content, err := os.ReadFile(indexFileName)
	if err != nil {
		return nil, err
	}
	if env != defs.EnvDev {
		indexHtml = string(content)
		metaBegin, metaEnd = getMetaPosition(indexHtml)
	}

	var notfoundHtml string
	content, err = os.ReadFile(publicDir + "/" + NotfoundName)
	if err == nil {
		notfoundHtml = string(content)
	} else {
		notfoundHtml = `<!DOCTYPE html><html lang="en"><head></head><body><h1>Page Not Found</h1></body></html>`
	}

	var manifest string
	content, err = os.ReadFile(publicDir + "/" + ManifestName)
	if err == nil {
		manifest = string(content)
	}

	return &IndexHtml{
		indexFileName: indexFileName,
		indexHtml:     indexHtml,
		metaBegin:     metaBegin,
		metaEnd:       metaEnd,
		notfoundHtml:  notfoundHtml,
		SsrManifest:   manifest,
	}, nil
}

func (this *IndexHtml) GetIndexHtml(result RenderResult, renderErr error) (int, string, error) {
	err := renderErr
	if err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "Error: 404") {
			return http.StatusNotFound, this.notfoundHtml, ErrorPageNotFound
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
		if result.PreloadLinks != "" {
			indexHtml = strings.Replace(indexHtml, "<!--preload-links-->", result.PreloadLinks, 1)
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
