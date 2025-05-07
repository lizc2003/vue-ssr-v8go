package logic

import (
	"github.com/lizc2003/vue-ssr-v8go/server/common/defs"
	"github.com/lizc2003/vue-ssr-v8go/server/common/tlog"
	"github.com/lizc2003/vue-ssr-v8go/server/common/util"
	"os"
	"strings"
)

type IndexHtml struct {
	indexFileName string
	indexHtml     string
	metaBegin     int
	metaEnd       int
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
		SsrManifest:   manifest,
	}, nil
}

func (this *IndexHtml) GetIndexHtml(result RenderResult, renderErr error) string {
	indexHtml, metaBegin, metaEnd := this.getRawIndexHtml()
	if renderErr == nil {
		if result.Meta != "" {
			indexHtml = indexHtml[:metaBegin] + result.Meta + indexHtml[metaEnd:]
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
	return indexHtml
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
	metaEnd += metaBegin + 1
	if metaBegin <= 0 {
		return 0, 0
	}
	return metaBegin, metaEnd + 15
}
