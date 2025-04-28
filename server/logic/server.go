package logic

import (
	"fmt"
	"github.com/lizc2003/vue-ssr-v8go/server/common/alarm"
	"github.com/lizc2003/vue-ssr-v8go/server/common/tlog"
	"github.com/lizc2003/vue-ssr-v8go/server/common/util"
	v8 "github.com/lizc2003/vue-ssr-v8go/server/v8"
	"strings"
	"time"
)

type Config struct {
	Host           string       `toml:"server_host"`
	Env            string       `toml:"env"`
	Log            tlog.Config  `toml:"Log"`
	AlarmSecret    string       `toml:"alarm_secret"`
	AssetsPrefix   string       `toml:"assets_prefix"`
	ForwardHeaders []string     `toml:"forward_headers"`
	VmConfig       v8.VmConfig  `toml:"Vm"`
	XhrConfig      v8.XhrConfig `toml:"Xhr"`
}

type Server struct {
	VmMgr          *v8.VmMgr
	ForwardHeaders []string
}

var ThisServer *Server

func RunServer(c *Config) {
	if c.AlarmSecret != "" {
		alarm.NewDefaultRobotFeiShu(c.AlarmSecret)
	}

	vmMgr, err := v8.NewVmMgr(c.Env, SendEventCallback, &c.VmConfig, &c.XhrConfig)
	if err != nil {
		tlog.Fatal(err.Error())
		return
	}

	ThisServer = &Server{
		VmMgr:          vmMgr,
		ForwardHeaders: getForwardHeaders(c.ForwardHeaders),
	}

	fmt.Printf("At %s, the server was started on port %s.\n",
		util.FormatTime(time.Now()),
		strings.Split(c.Host, ":")[1])
	util.GraceHttpServe(c.Host, GetHttpHandler(c.Env, c.AssetsPrefix))
}
