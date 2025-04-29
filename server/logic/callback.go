package logic

import (
	"encoding/json"
	"github.com/lizc2003/vue-ssr-v8go/server/common/util"
)

func SendEventCallback(renderId int64, evt string, msg string) {
	switch evt {
	case "render_ok":
		bOK := false
		var result RenderResult
		err := json.Unmarshal(util.UnsafeStr2Bytes(msg), &result)
		if err == nil {
			bOK = true
		} else {
			result.Html = err.Error()
		}
		ThisServer.RenderMgr.SendResult(renderId, bOK, result)
	case "render_fail":
		ThisServer.RenderMgr.SendResult(renderId, false,
			RenderResult{Html: msg})
	}
}
