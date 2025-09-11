package logic

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/lizc2003/vue-ssr-v8go/server/common/alarm"
	"github.com/lizc2003/vue-ssr-v8go/server/common/tlog"
	"github.com/lizc2003/vue-ssr-v8go/server/common/util"
	"github.com/lizc2003/vue-ssr-v8go/server/v8"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

type Config struct {
	Host        string      `toml:"server_host"`
	Env         string      `toml:"env"`
	AlarmUrl    string      `toml:"alarm_url"`
	AlarmSecret string      `toml:"alarm_secret"`
	Log         tlog.Config `toml:"Log"`
	VmConfig    v8.VmConfig `toml:"V8vm"`
	SsrConfig   SSRConfig   `toml:"SSR"`
	Proxy       ProxyConfig `toml:"Proxy"`
}

type SSRConfig struct {
	DistDir         string   `toml:"dist_dir"`
	Timeout         int      `toml:"timeout"`
	ResponseHeaders []string `toml:"response_headers"`
	Origin          string   `toml:"origin"`
	OriginRewrite   string   `toml:"origin_rewrite"`
}

type Server struct {
	RenderMgr       *RenderMgr
	VmMgr           *v8.VmMgr
	Origin          string
	SsrTime         time.Duration
	ResponseHeaders map[string]string
}

var ThisServer *Server

func RunServer(c *Config) {
	if c.AlarmUrl != "" && c.AlarmSecret != "" {
		alarm.NewDefaultRobot(c.AlarmUrl, c.AlarmSecret)
	}

	err := InitReverseProxy(c.Proxy.Locations)
	if err != nil {
		tlog.Fatal(err.Error())
		return
	}

	distPath, err := getDistPath(c.SsrConfig.DistDir)
	if err != nil {
		tlog.Fatal(err.Error())
		return
	}
	publicDir := distPath + PublicPath
	serverDir := distPath + ServerPath

	ssrTimeout := int32(c.SsrConfig.Timeout)
	if ssrTimeout < 1 {
		ssrTimeout = 1
	} else if ssrTimeout > 120 {
		ssrTimeout = 120
	}
	if c.VmConfig.DeleteDelayTime > ssrTimeout {
		c.VmConfig.DeleteDelayTime = ssrTimeout
	}

	originRewrite, err := getOriginRewrite(c)
	if err != nil {
		tlog.Fatal(err.Error())
		return
	}
	vmMgr, err := v8.NewVmMgr(c.Env, serverDir, SendMessageCallback, &c.VmConfig, originRewrite)
	if err != nil {
		tlog.Fatal(err.Error())
		return
	}
	vmMgr.DumpHeapDir = c.Log.Dir
	os.MkdirAll(vmMgr.DumpHeapDir, 0755)

	renderMgr, err := NewRenderMgr(c.Env, publicDir)
	if err != nil {
		tlog.Fatal(err.Error())
		return
	}

	originJson, _ := json.Marshal(c.SsrConfig.Origin)

	ThisServer = &Server{
		RenderMgr:       renderMgr,
		VmMgr:           vmMgr,
		Origin:          string(originJson),
		SsrTime:         time.Duration(ssrTimeout) * time.Second,
		ResponseHeaders: getResponseHeaders(c.SsrConfig.ResponseHeaders),
	}

	go runDumpSignalRoutine()

	fmt.Printf("At %s, the server was started on port %s.\n",
		util.FormatTime(time.Now()),
		strings.Split(c.Host, ":")[1])
	util.GraceHttpServe(c.Host, GetHttpHandler(c.Env, publicDir))
}

func runDumpSignalRoutine() {
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGUSR2)
	for {
		sig := <-ch
		if sig == syscall.SIGUSR2 {
			ThisServer.VmMgr.SignalDumpHeap()
		}
	}
}

func getResponseHeaders(headers []string) map[string]string {
	headersMap := make(map[string]string)
	for _, header := range headers {
		headerParts := strings.SplitN(header, ":", 2)
		if len(headerParts) != 2 {
			continue
		}
		headersMap[strings.TrimSpace(headerParts[0])] = strings.TrimSpace(headerParts[1])
	}
	return headersMap
}

func getOriginRewrite(c *Config) (*v8.OriginRewrite, error) {
	if c.SsrConfig.Origin == "" {
		return nil, errors.New("ssr.origin is empty")
	}

	originUrl, err := v8.ParseUrl(c.SsrConfig.Origin)
	if err != nil {
		return nil, err
	}

	ret := &v8.OriginRewrite{
		OriginHost: originUrl.Host,
	}

	if c.SsrConfig.OriginRewrite != "" {
		rewriteUrl, err := v8.ParseUrl(c.SsrConfig.OriginRewrite)
		if err != nil {
			return nil, err
		}
		ret.RewriteUrl = rewriteUrl
	}

	return ret, nil
}
