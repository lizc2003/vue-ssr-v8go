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
			ssrHeaders[k] = v
		}
	}

	render := ThisServer.RenderMgr.NewRender()
	tlog.Infof("request %d: %s", render.renderId, url)

	beginTime := time.Now()

	result, err := ssrRender(render, url, ssrHeaders)
	statusCode, indexHtml, err := ThisServer.RenderMgr.IndexHtml.GetIndexHtml(result, err)
	util.WriteHtmlResponse(writer, statusCode, indexHtml)

	elapse := time.Since(beginTime)
	if err != nil {
		if err == ErrorSsrOff {
			tlog.Infof("request %d finish: %s, elapse: %v, ssr off", render.renderId, url, elapse)
		} else if err == ErrorPageNotFound {
			tlog.Infof("request %d finish: %s, elapse: %v, page not found", render.renderId, url, elapse)
		} else {
			tlog.Errorf("request %d finish: %s, elapse: %v, ssr error: %v", render.renderId, url, elapse, err)
		}
	} else {
		tlog.Infof("request %d finish: %s, elapse: %v", render.renderId, url, elapse)
	}
}

func ssrRender(render *Render, url string, ssrHeaders map[string]string) (RenderResult, error) {
	ssrHeadersJson, _ := json.Marshal(ssrHeaders)
	urlJson, _ := json.Marshal(url)

	var jsCode strings.Builder
	jsCode.Grow(renderJsLength + len(ssrHeadersJson) + len(urlJson) + 64)
	jsCode.WriteString(renderJsPart1)
	jsCode.WriteString(`{renderId:`)
	jsCode.WriteString(strconv.FormatInt(render.renderId, 10))
	jsCode.WriteString(`,url:`)
	jsCode.Write(urlJson)
	jsCode.WriteString(`,ssrHeaders:`)
	jsCode.Write(ssrHeadersJson)
	jsCode.WriteString(`}`)
	jsCode.WriteString(renderJsPart2)

	err := ThisServer.VmMgr.Execute(jsCode.String(), renderJsName)
	if err == nil {
		select {
		case <-render.end:
			if !render.bOK {
				err = errors.New(render.result.Html)
			}
		case <-time.After(ThisServer.SsrTime):
			err = ErrorRenderTimeout
		}
	}
	ThisServer.RenderMgr.CloseRender(render.renderId)

	return render.result, err
}
