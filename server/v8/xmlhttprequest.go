package v8

import (
	"encoding/json"
	"fmt"
	"github.com/lizc2003/vue-ssr-v8go/server/common/alarm"
	"github.com/lizc2003/vue-ssr-v8go/server/common/tlog"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

type XhrConfig struct {
	Hosts []string
}

type xhrCmd struct {
	Cmd     string            `json:"cmd"`
	XhrId   int               `json:"xhr_id"`
	Url     string            `json:"url"`
	Method  string            `json:"method"`
	Headers map[string]string `json:"headers"`
	Post    string            `json:"post"`
	Timeout int               `json:"timeout"`
	worker  *Worker
	aborted bool
}

type xmlHttpRequestMgr struct {
	mutex sync.Mutex
	queue chan *xhrCmd
	reqs  map[int]*xhrCmd
	maxId int
}

func NewXmlHttpRequestMgr(maxThreadCount int32, c *XhrConfig) *xmlHttpRequestMgr {
	if maxThreadCount > MaxXhrThreadCount {
		maxThreadCount = MaxXhrThreadCount
	}

	httpClient := newHttpClient()
	queue := make(chan *xhrCmd, maxThreadCount*2)
	reqs := make(map[int]*xhrCmd)
	mgr := &xmlHttpRequestMgr{
		queue: queue,
		reqs:  reqs,
	}

	for i := int32(0); i < maxThreadCount; i++ {
		go func() {
			for req := range queue {
				performXhr(req, httpClient, c)

				mgr.mutex.Lock()
				delete(mgr.reqs, req.XhrId)
				mgr.mutex.Unlock()
			}
		}()
	}
	return mgr
}

func (this *xmlHttpRequestMgr) open(req *xhrCmd) int {
	this.mutex.Lock()
	this.maxId++
	req.XhrId = this.maxId
	this.reqs[req.XhrId] = req
	this.mutex.Unlock()

	beginTime := time.Now()
	this.queue <- req
	tlog.Infof("xhr %d: %s, wait time: %v", req.XhrId, req.Url, time.Since(beginTime))
	return req.XhrId
}

func (this *xmlHttpRequestMgr) abort(xhrId int) {
	this.mutex.Lock()
	if req, ok := this.reqs[xhrId]; ok {
		req.aborted = true
	}
	this.mutex.Unlock()
}

func performXhr(req *xhrCmd, client *http.Client, c *XhrConfig) {
	worker := req.worker
	evt := xhrEvent{XhrId: req.XhrId}

	if req.aborted {
		sendXhrFinishEvent(worker, &evt)
		return
	}

	evt.Event = "onstart"
	sendXhrEvent(worker, &evt)

	url := req.Url
	var request *http.Request
	var err error
	if len(req.Post) == 0 {
		request, err = http.NewRequest(req.Method, url, nil)
	} else {
		request, err = http.NewRequest(req.Method, url, strings.NewReader(req.Post))
		if _, ok := req.Headers["Content-Type"]; !ok {
			c := req.Post[0]
			if c == '{' || c == '[' {
				req.Headers["Content-Type"] = "application/json;charset=UTF-8"
			} else {
				req.Headers["Content-Type"] = "application/x-www-form-urlencoded"
			}
		}
	}
	if err != nil {
		sendXhrErrorEvent(worker, &evt, err)
		return
	}

	// fixme
	//request.Host = this.internalApiHost

	for k, v := range req.Headers {
		if k == "SSR-Headers" {
			if v != "" {
				var headers map[string]string
				err := json.Unmarshal([]byte(v), &headers)
				if err == nil {
					for kk, vv := range headers {
						if vv != "" {
							kk = strings.ReplaceAll(kk, "_", "-")
							tlog.Debugf("ssr header %s: %s", kk, vv)
							request.Header.Set(kk, vv)
						}
					}
				}
			}
		} else {
			request.Header.Set(k, v)
		}
	}
	if req.aborted {
		sendXhrFinishEvent(worker, &evt)
		return
	}

	resp, err := client.Do(request)
	if req.aborted {
		sendXhrFinishEvent(worker, &evt)
		return
	}
	if err != nil {
		sendXhrErrorEvent(worker, &evt, err)
		return
	}
	if resp == nil {
		err = fmt.Errorf("response is nil: %s", url)
		sendXhrErrorEvent(worker, &evt, err)
		return
	}

	evt.Event = "onheader"
	evt.Status = int32(resp.StatusCode)
	evt.Headers = make(map[string]string)
	for k, v := range resp.Header {
		evt.Headers[k] = strings.Join(v, "&")
	}
	sendXhrEvent(worker, &evt)

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if req.aborted {
		sendXhrFinishEvent(worker, &evt)
		return
	}
	if err != nil {
		sendXhrErrorEvent(worker, &evt, err)
		return
	}

	evt.Event = "onend"
	evt.Response = string(body)
	sendXhrEvent(worker, &evt)

	sendXhrFinishEvent(worker, &evt)
}

func sendXhrFinishEvent(w *Worker, evt *xhrEvent) {
	evt.Event = "onfinish"
	sendXhrEvent(w, evt)
}

func sendXhrErrorEvent(w *Worker, evt *xhrEvent, err error) {
	tlog.Error(err)
	go alarm.SendAlert(err.Error())

	evt.Event = "onerror"
	evt.Error = err.Error()
	sendXhrEvent(w, evt)

	sendXhrFinishEvent(w, evt)
}

func sendXhrEvent(w *Worker, evt *xhrEvent) {
	err := w.SendXhrEvent(evt)
	if err != nil {
		tlog.Error(err)
	}
	evt.Reset()
}

func handleXMLHttpRequestCmd(w *Worker, msg string) string {
	var req xhrCmd
	err := json.Unmarshal([]byte(msg), &req)
	if err != nil {
		tlog.Error(err)
		return ""
	}
	req.worker = w

	switch req.Cmd {
	case "open":
		xhrId := ThisVmMgr.xhrMgr.open(&req)
		return strconv.FormatInt(int64(xhrId), 10)
	case "abort":
		ThisVmMgr.xhrMgr.abort(req.XhrId)
		return ""
	}

	tlog.Errorf("unknown xhr cmd: %s", req.Cmd)
	return ""
}
