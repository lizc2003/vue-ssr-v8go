package logic

import (
	"fmt"
	"github.com/lizc2003/vue-ssr-v8go/server/common/alarm"
	"github.com/lizc2003/vue-ssr-v8go/server/common/tlog"
	"github.com/lizc2003/vue-ssr-v8go/server/common/util"
	"github.com/lizc2003/vue-ssr-v8go/server/v8"
	"strings"
	"time"
)

type Config struct {
	Host         string       `toml:"server_host"`
	Env          string       `toml:"env"`
	AlarmSecret  string       `toml:"alarm_secret"`
	DistDir      string       `toml:"dist_dir"`
	AssetsPrefix string       `toml:"assets_prefix"`
	Log          tlog.Config  `toml:"Log"`
	VmConfig     v8.VmConfig  `toml:"V8vm"`
	ApiConfig    v8.ApiConfig `toml:"Api"`
	Proxy        ProxyConfig  `toml:"Proxy"`
}

type Server struct {
	RenderMgr *RenderMgr
	VmMgr     *v8.VmMgr
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

	distPath, err := getDistPath(c.DistDir)
	if err != nil {
		tlog.Fatal(err.Error())
		return
	}
	publicDir := distPath + PublicPath
	serverDir := distPath + ServerPath

	vmMgr, err := v8.NewVmMgr(c.Env, serverDir, SendEventCallback, &c.VmConfig, &c.ApiConfig)
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
		RenderMgr: renderMgr,
		VmMgr:     vmMgr,
	}

	fmt.Printf("At %s, the server was started on port %s.\n",
		util.FormatTime(time.Now()),
		strings.Split(c.Host, ":")[1])
	util.GraceHttpServe(c.Host, GetHttpHandler(c.Env, publicDir, c.AssetsPrefix))
}
