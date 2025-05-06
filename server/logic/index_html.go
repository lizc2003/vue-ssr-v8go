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
}

func NewIndexHtml(env string, publicDir string) (*IndexHtml, error) {
	var indexHtml string
	indexFileName := publicDir + "/" + IndexName
	content, err := os.ReadFile(indexFileName)
	if err != nil {
		return nil, err
	}
	if env != defs.EnvDev {
		indexHtml = string(content)
	}

	return &IndexHtml{
		indexFileName: indexFileName,
		indexHtml:     indexHtml,
	}, nil
}

func (this *IndexHtml) GetIndexHtml(result RenderResult, renderErr error) string {
	indexHtml := this.getRawIndexHtml()
	if renderErr == nil {
		if result.State != "" {
			state := "window.__INITIAL_STATE__ = " + result.State
			indexHtml = strings.Replace(indexHtml, "<!--app-state-->", state, 1)
		}
		indexHtml = strings.Replace(indexHtml, "<!--app-html-->", result.Html, 1)
	}
	return indexHtml
}

func (this *IndexHtml) getRawIndexHtml() string {
	if this.indexHtml != "" {
		return this.indexHtml
	}

	content, err := os.ReadFile(this.indexFileName)
	if err != nil {
		tlog.Error(err)
		return ""
	}
	return util.UnsafeBytes2Str(content)
}
