package logic

func SendMessageCallback(mtype int64, param1 int64, param2 string, param3 string, param4 string, param5 string) {
	switch mtype {
	case 10:
		result := RenderResult{
			Html:    param2,
			Meta:    param3,
			State:   param4,
			Modules: param5,
		}

		bOK := true
		if result.Html == "" {
			bOK = false
			result.Html = "no render result"
		}
		ThisServer.RenderMgr.SendResult(param1, bOK, result)
	case 11:
		ThisServer.RenderMgr.SendResult(param1, false,
			RenderResult{Html: param2})
	}
}
