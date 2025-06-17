package main

import (
	"github.com/lizc2003/vue-ssr-v8go/server/common/defs"
	"github.com/lizc2003/vue-ssr-v8go/server/common/tlog"
	"github.com/lizc2003/vue-ssr-v8go/server/common/util"
	"github.com/lizc2003/vue-ssr-v8go/server/logic"
)

func main() {
	var c logic.Config
	bOK, _ := util.NewConfig("./conf-dev.toml", &c)
	if !bOK {
		return
	}

	tlog.Init(&c.Log, defs.App, "")
	logic.RunServer(&c)
	tlog.Close()
}
