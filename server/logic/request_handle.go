package logic

import (
	"github.com/gin-gonic/gin"
	"github.com/lizc2003/vue-ssr-v8go/server/common/tlog"
	"strings"
)

func HandleSsrRequest(c *gin.Context) {
	reqURL := c.Request.URL
	url := reqURL.Path
	if len(reqURL.RawQuery) > 0 {
		url += "?"
		url += reqURL.RawQuery
	}

	forwardHeaders := make(map[string]string)
	for _, k := range ThisServer.ForwardHeaders {
		v := c.GetHeader(k)
		if v == "" && k == "X-Forwarded-For" {
			v = c.ClientIP()
		}
		if v != "" {
			forwardHeaders[strings.ReplaceAll(k, "-", "_")] = v
		}
	}

	tlog.Infof("http request: %s", url)
}
