package logic

import (
	"github.com/lizc2003/vue-ssr-v8go/server/common/defs"
	"github.com/lizc2003/vue-ssr-v8go/server/common/tlog"
	"github.com/lizc2003/vue-ssr-v8go/server/common/util"
	"os"
	"sync"
)

type Render struct {
	end      chan struct{}
	renderId int64
	result   RenderResult
	bOK      bool
}

type RenderMgr struct {
	mutex         sync.Mutex
	renders       map[int64]*Render
	maxId         int64
	indexFileName string
	indexHtml     string
}

func NewRenderMgr(env string, publicDir string) (*RenderMgr, error) {
	var indexHtml string
	indexFileName := publicDir + "/" + IndexName
	content, err := os.ReadFile(indexFileName)
	if err != nil {
		return nil, err
	}
	if env != defs.EnvDev {
		indexHtml = string(content)
	}

	return &RenderMgr{
		renders:       make(map[int64]*Render),
		indexFileName: indexFileName,
		indexHtml:     indexHtml,
	}, nil
}

func (this *RenderMgr) NewRender() *Render {
	req := &Render{
		end: make(chan struct{}),
	}

	this.mutex.Lock()
	this.maxId++
	req.renderId = this.maxId
	this.renders[req.renderId] = req
	this.mutex.Unlock()
	return req
}

func (this *RenderMgr) CloseRender(renderId int64) {
	this.mutex.Lock()
	if _, ok := this.renders[renderId]; ok {
		delete(this.renders, renderId)
	}
	this.mutex.Unlock()
}

func (this *RenderMgr) SendResult(renderId int64, bOK bool, result RenderResult) {
	this.mutex.Lock()
	if req, ok := this.renders[renderId]; ok {
		req.result = result
		req.bOK = bOK
		close(req.end)
		delete(this.renders, renderId)
	}
	this.mutex.Unlock()
}

func (this *RenderMgr) GetIndexHtml() string {
	if this.indexHtml != "" {
		return this.indexHtml
	}

	content, err := os.ReadFile(this.indexFileName)
	if err != nil {
		tlog.Error(err)
		return ""
	}
	return util.UnsafeBytes2Str(content)
}
