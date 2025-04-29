package logic

import (
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/render"
	"github.com/lizc2003/vue-ssr-v8go/server/common/tlog"
	"github.com/lizc2003/vue-ssr-v8go/server/common/util"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var ErrorRenderTimeout = errors.New("render timeout.")

func HandleSsrRequest(c *gin.Context) {
	reqURL := c.Request.URL
	url := reqURL.Path
	if len(reqURL.RawQuery) > 0 {
		url += "?"
		url += reqURL.RawQuery
	}

	ssrHeaders := make(map[string]string)
	for _, k := range ThisServer.SsrHeaders {
		v := c.GetHeader(k)
		if v == "" && k == "X-Forwarded-For" {
			v = c.ClientIP()
		}
		if v != "" {
			ssrHeaders[strings.ReplaceAll(k, "-", "_")] = v
		}
	}

	tlog.Infof("request: %s", url)

	result, err := ssrRender(url, ssrHeaders)
	indexHtml := ThisServer.RenderMgr.GetIndexHtml()
	if err == nil {
		indexHtml = strings.Replace(indexHtml, "<!--app-html-->", result.Html, 1)
	} else {
		tlog.Errorf("ssr render failed: %s", err.Error())
	}

	c.Render(http.StatusOK, render.Data{
		ContentType: "text/html; charset=utf-8",
		Data:        util.UnsafeStr2Bytes(indexHtml),
	})
}

func ssrRender(url string, ssrHeaders map[string]string) (RenderResult, error) {
	req := ThisServer.RenderMgr.NewRender()

	ssrHeadersJson, _ := json.Marshal(ssrHeaders)
	urlJson, _ := json.Marshal(url)

	var jsCode strings.Builder
	jsCode.Grow(renderJsLength + len(ssrHeadersJson) + len(urlJson) + 64)
	jsCode.WriteString(renderJsPart1)
	jsCode.WriteString(`{renderId:`)
	jsCode.WriteString(strconv.FormatInt(req.renderId, 10))
	jsCode.WriteString(`,url:`)
	jsCode.Write(urlJson)
	jsCode.WriteString(`,ssrHeaders:`)
	jsCode.Write(ssrHeadersJson)
	jsCode.WriteString(`}`)
	jsCode.WriteString(renderJsPart2)

	err := ThisServer.VmMgr.Execute(jsCode.String(), renderJsName)
	if err == nil {
		select {
		case <-req.end:
		case <-time.After(RenderTimeout):
			err = ErrorRenderTimeout
		}
	}
	ThisServer.RenderMgr.CloseRender(req.renderId)

	return req.result, err
}
