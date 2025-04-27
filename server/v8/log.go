package v8

import (
	"fmt"
	"github.com/lizc2003/vue-ssr-v8go/server/common/tlog"
	"github.com/lizc2003/vue-ssr-v8go/server/common/util"
	"github.com/tommie/v8go"
	"strconv"
	"strings"
	"time"
)

// log object: console.log(JSON.stringify(obj))

type ConsoleObj struct {
	level tlog.LEVEL
}

func newConsoleObj() *ConsoleObj {
	return &ConsoleObj{level: tlog.Level()}
}

func (this *ConsoleObj) ConsoleAPIMessage(msg v8go.ConsoleAPIMessage) {
	var levelTxt string
	level := tlog.DEBUG
	switch msg.ErrorLevel {
	case v8go.ErrorLevelLog:
		level = tlog.FATAL
		levelTxt = "LOG"
	case v8go.ErrorLevelInfo:
		level = tlog.INFO
		levelTxt = "INFO"
	case v8go.ErrorLevelWarning:
		level = tlog.WARN
		levelTxt = "WARN"
	case v8go.ErrorLevelError:
		level = tlog.ERROR
		levelTxt = "ERROR"
	default:
		level = tlog.DEBUG
		levelTxt = "DEBUG"
	}

	if level >= this.level {
		var sb strings.Builder
		sb.WriteString(util.FormatTime(time.Now()))
		sb.WriteByte(' ')
		sb.WriteString(levelTxt)
		sb.WriteByte(' ')
		sb.WriteString(msg.Url)
		sb.WriteByte(':')
		sb.WriteString(strconv.FormatInt(int64(msg.LineNumber), 10))
		sb.WriteByte(':')
		sb.WriteString(strconv.FormatInt(int64(msg.ColumnNumber), 10))
		sb.WriteByte(' ')
		sb.WriteString(msg.Message)
		fmt.Println(sb.String())
	}
}
