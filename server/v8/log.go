package v8

import (
	"github.com/lizc2003/v8go"
	"github.com/lizc2003/vue-ssr-v8go/server/common/tlog"
	"strconv"
)

// log object: console.log(JSON.stringify(obj))

type ConsoleObj struct {
}

func newConsoleObj() ConsoleObj {
	return ConsoleObj{}
}

func (this ConsoleObj) ConsoleAPIMessage(msg v8go.ConsoleAPIMessage) {
	level := tlog.DEBUG
	switch msg.ErrorLevel {
	case v8go.ErrorLevelLog:
		level = tlog.INFO
	case v8go.ErrorLevelInfo:
		level = tlog.INFO
	case v8go.ErrorLevelWarning:
		level = tlog.WARN
	case v8go.ErrorLevelError:
		level = tlog.ERROR
	default:
		level = tlog.DEBUG
	}

	line := strconv.FormatInt(int64(msg.LineNumber), 10) + ":" +
		strconv.FormatInt(int64(msg.ColumnNumber), 10)
	tlog.Log(level, msg.Url, line, msg.Message)
}
