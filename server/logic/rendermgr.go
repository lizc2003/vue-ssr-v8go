package logic

import (
	"sync"
)

type Render struct {
	end      chan struct{}
	renderId int64
	result   RenderResult
	bOK      bool
}

type RenderMgr struct {
	mutex     sync.Mutex
	renders   map[int64]*Render
	maxId     int64
	IndexHtml *IndexHtml
}

func NewRenderMgr(env string, publicDir string) (*RenderMgr, error) {
	indexHtml, err := NewIndexHtml(env, publicDir)
	if err != nil {
		return nil, err
	}

	return &RenderMgr{
		renders:   make(map[int64]*Render),
		IndexHtml: indexHtml,
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
