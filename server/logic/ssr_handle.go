package logic

import (
	"encoding/json"
	"errors"
	"github.com/lizc2003/vue-ssr-v8go/server/common/tlog"
	"github.com/lizc2003/vue-ssr-v8go/server/common/util"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func HandleSsrRequest(writer http.ResponseWriter, request *http.Request) {
	reqURL := request.URL
	url := reqURL.Path
	if len(reqURL.RawQuery) > 0 {
		url += "?"
		url += reqURL.RawQuery
	}

	ssrHeaders := make(map[string]string)
	for _, k := range ForwardHeaders {
		v := request.Header.Get(k)
		if v == "" && k == "X-Forwarded-For" {
			v = util.GetClientIP(request)
		}
		if v != "" {
			ssrHeaders[strings.ReplaceAll(k, "-", "_")] = v
		}
	}

	tlog.Infof("request: %s", url)
	beginTime := time.Now()

	result, err := ssrRender(url, ssrHeaders)
	statusCode, indexHtml, err := ThisServer.RenderMgr.IndexHtml.GetIndexHtml(result, err)
	util.WriteHtmlResponse(writer, statusCode, indexHtml)

	elapse := time.Since(beginTime)
	if err != nil {
		if err == ErrorSsrOff {
			tlog.Infof("request finish: %s, elapse: %v, ssr off", url, elapse)
		} else if err == ErrorPageNotFound {
			tlog.Infof("request finish: %s, elapse: %v, page not found", url, elapse)
		} else {
			tlog.Infof("request finish: %s, elapse: %v, ssr error: %v", url, elapse, err)
		}
	} else {
		tlog.Infof("request finish: %s, elapse: %v", url, elapse)
	}
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
			if !req.bOK {
				err = errors.New(req.result.Html)
			}
		case <-time.After(ThisServer.SsrTime):
			err = ErrorRenderTimeout
		}
	}
	ThisServer.RenderMgr.CloseRender(req.renderId)

	return req.result, err
}
