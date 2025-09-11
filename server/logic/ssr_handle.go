package logic

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/lizc2003/vue-ssr-v8go/server/common/alarm"
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
	if err == ErrorPageRedirect {
		http.Redirect(writer, request, indexHtml, statusCode)
	} else {
		util.WriteHtmlResponse(writer, statusCode, indexHtml, ThisServer.ResponseHeaders)
	}

	elapse := time.Since(beginTime)
	if err != nil {
		if err == ErrorSsrOff {
			tlog.Infof("request %d finish(%d): %s, elapse: %v, ssr off", render.renderId, render.workerId, url, elapse)
		} else if err == ErrorPageNotFound {
			tlog.Infof("request %d finish(%d): %s, elapse: %v, page not found", render.renderId, render.workerId, url, elapse)
		} else if err == ErrorPageRedirect {
			tlog.Infof("request %d finish(%d): %s, elapse: %v, page redirect %d %s", render.renderId, render.workerId, url, elapse, statusCode, indexHtml)
		} else {
			errMsg := fmt.Sprintf("request %d finish(%d): %s, elapse: %v, ssr error: %v", render.renderId, render.workerId, url, elapse, err)
			tlog.Error(errMsg)
			if err == ErrorRenderTimeout {
				alarm.SendAlert(errMsg)
			}
		}
	} else {
		tlog.Infof("request %d finish(%d): %s, elapse: %v", render.renderId, render.workerId, url, elapse)
	}
}

func ssrRender(render *Render, url string, ssrHeaders map[string]string) (RenderResult, error) {
	ssrHeadersJson, _ := json.Marshal(ssrHeaders)
	urlJson, _ := json.Marshal(url)

	var jsCode strings.Builder
	jsCode.Grow(renderJsLength + len(ssrHeadersJson) + len(urlJson) + len(ThisServer.Origin) + 64)
	jsCode.WriteString(renderJsPart1)
	jsCode.WriteString(`{renderId:`)
	jsCode.WriteString(strconv.FormatInt(render.renderId, 10))
	jsCode.WriteString(`,url:`)
	jsCode.Write(urlJson)
	jsCode.WriteString(`,origin:`)
	jsCode.WriteString(ThisServer.Origin)
	jsCode.WriteString(`,ssrHeaders:`)
	jsCode.Write(ssrHeadersJson)
	jsCode.WriteString(`}`)
	jsCode.WriteString(renderJsPart2)

	workerId, err := ThisServer.VmMgr.Execute(jsCode.String(), renderJsName)
	render.workerId = workerId
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
