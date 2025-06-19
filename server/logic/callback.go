package logic

import (
	"encoding/json"
	"github.com/lizc2003/vue-ssr-v8go/server/common/util"
)

func SendMessageCallback(mtype int64, param1 int64, param2 string) {
	switch mtype {
	case 10:
		bOK := false
		var result RenderResult
		err := json.Unmarshal(util.UnsafeStr2Bytes(param2), &result)
		if err == nil {
			bOK = true
		} else {
			result.Html = err.Error()
		}
		ThisServer.RenderMgr.SendResult(param1, bOK, result)
	case 11:
		ThisServer.RenderMgr.SendResult(param1, false,
			RenderResult{Html: param2})
	}
}
