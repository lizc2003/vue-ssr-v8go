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
	Host         string       `toml:"server_host"`
	Env          string       `toml:"env"`
	Log          tlog.Config  `toml:"Log"`
	AlarmSecret  string       `toml:"alarm_secret"`
	AssetsPrefix string       `toml:"assets_prefix"`
	SsrHeaders   []string     `toml:"ssr_headers"`
	VmConfig     v8.VmConfig  `toml:"Vm"`
	XhrConfig    v8.XhrConfig `toml:"Xhr"`
	Proxy        ProxyConfig  `toml:"Proxy"`
}

type Server struct {
	RenderMgr  *RenderMgr
	VmMgr      *v8.VmMgr
	SsrHeaders []string
}

var ThisServer *Server

func RunServer(c *Config) {
	if c.AlarmSecret != "" {
		alarm.NewDefaultRobotFeiShu(c.AlarmSecret)
	}

	err := InitReverseProxy(c.Proxy.Locations)
	if err != nil {
		tlog.Fatal(err.Error())
		return
	}

	distPath := getDistPath()
	publicDir := distPath + ClientPath
	serverDir := distPath + ServerPath

	vmMgr, err := v8.NewVmMgr(c.Env, serverDir, SendEventCallback, &c.VmConfig, &c.XhrConfig)
	if err != nil {
		tlog.Fatal(err.Error())
		return
	}

	renderMgr, err := NewRenderMgr(c.Env, publicDir)
	if err != nil {
		tlog.Fatal(err.Error())
		return
	}

	ThisServer = &Server{
		RenderMgr:  renderMgr,
		VmMgr:      vmMgr,
		SsrHeaders: getSsrHeaders(c.SsrHeaders),
	}

	fmt.Printf("At %s, the server was started on port %s.\n",
		util.FormatTime(time.Now()),
		strings.Split(c.Host, ":")[1])
	util.GraceHttpServe(c.Host, GetHttpHandler(c.Env, publicDir, c.AssetsPrefix))
}
