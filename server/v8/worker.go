package v8

import "C"
import (
	"encoding/json"
	"github.com/lizc2003/vue-ssr-v8go/server/common/tlog"
	"github.com/tommie/v8go"
	"strings"
	"sync"
)

type xhrEvent struct {
	XhrId    int               `json:"xhr_id"`
	Event    string            `json:"event"`
	Error    string            `json:"error,omitempty"`
	Status   int32             `json:"status,omitempty"`
	Headers  map[string]string `json:"headers,omitempty"`
	Response string            `json:"response,omitempty"`
}

func (this *xhrEvent) Reset() {
	this.Event = ""
	this.Error = ""
	this.Status = 0
	this.Headers = nil
	this.Response = ""
}

type SendEventCallback func(renderId int64, evt string, msg string)

type Worker struct {
	isolate         *v8go.Isolate
	inspectorClient *v8go.InspectorClient
	inspector       *v8go.Inspector
	v8ctx           *v8go.Context

	disposed   bool
	running    bool
	mutex      sync.Mutex
	evtQueue   []*xhrEvent
	callback   SendEventCallback
	expireTime int64
}

func NewWorker(callback SendEventCallback) (*Worker, error) {
	isolate := v8go.NewIsolate()
	client := v8go.NewInspectorClient(newConsoleObj())
	inspector := v8go.NewInspector(isolate, client)
	v8ctx := v8go.NewContext(isolate)
	inspector.ContextCreated(v8ctx)

	_, err := v8ctx.RunScript(gGlobalJs, "init.js")
	if err != nil {
		return nil, err
	}

	w := &Worker{
		isolate:         isolate,
		inspectorClient: client,
		inspector:       inspector,
		v8ctx:           v8ctx,
		callback:        callback,
	}
	err = setFunctionCallback(w)
	if err != nil {
		return nil, err
	}

	return w, nil
}

func (this *Worker) Dispose() {
	if this.disposed {
		return
	}
	this.disposed = true

	this.inspector.ContextDestroyed(this.v8ctx)
	this.v8ctx.Close()
	this.inspector.Dispose()
	this.inspectorClient.Dispose()
	this.isolate.Dispose()
}

func (this *Worker) Acquire() bool {
	bOK := false
	if this.mutex.TryLock() {
		if this.running {
			tlog.Error("v8worker still running")
		} else {
			this.running = true
			bOK = true
		}
		this.mutex.Unlock()
	}
	return bOK
}

func (this *Worker) Release() {
	this.mutex.Lock()
	if len(this.evtQueue) > 0 {
		for _, evt := range this.evtQueue {
			doSendXhrEvent(this, evt)
		}
		this.evtQueue = nil
	}
	this.running = false
	this.mutex.Unlock()
}

func (this *Worker) Execute(code string, scriptName string) error {
	_, err := this.v8ctx.RunScript(code, scriptName)
	if err != nil {
		return ToJsError(err)
	}
	return nil
}

func (this *Worker) SendXhrEvent(evt *xhrEvent) error {
	var err error
	this.mutex.Lock()
	if this.running || len(this.evtQueue) > 0 {
		this.evtQueue = append(this.evtQueue, evt)
	} else {
		err = doSendXhrEvent(this, evt)
	}
	this.mutex.Unlock()
	return err
}

func (this *Worker) SetExpireTime(expireTime int64) {
	this.expireTime = expireTime
}

func (this *Worker) GetExpireTime() int64 {
	return this.expireTime
}

////////////////////////////////////////////

func setFunctionCallback(w *Worker) error {
	v8goOT := v8go.NewObjectTemplate(w.isolate)

	xhrCmd := v8go.NewFunctionTemplate(w.isolate, func(info *v8go.FunctionCallbackInfo) *v8go.Value {
		args := info.Args()
		if len(args) < 1 {
			return nil
		}
		ret, _ := v8go.NewValue(w.isolate,
			handleXMLHttpRequestCmd(w, args[0].String()),
		)
		return ret
	})
	v8goOT.Set("handleXhrCmd", xhrCmd)

	sendEvent := v8go.NewFunctionTemplate(w.isolate, func(info *v8go.FunctionCallbackInfo) *v8go.Value {
		if w.callback != nil {
			args := info.Args()
			if len(args) < 3 {
				return nil
			}
			w.callback(args[0].Integer(), args[1].String(), args[2].String())
		}
		return nil
	})
	v8goOT.Set("sendEvent", sendEvent)

	v8goObj, err := v8goOT.NewInstance(w.v8ctx)
	if err != nil {
		return err
	}
	return w.v8ctx.Global().Set("v8goGo", v8goObj)
}

func doSendXhrEvent(w *Worker, evt any) error {
	s, err := json.Marshal(evt)
	if err != nil {
		tlog.Error(err)
		return err
	}

	var sb strings.Builder
	sb.WriteString("v8goJs.xhrMgr.sendEvent(")
	sb.Write(s)
	sb.WriteByte(')')

	_, err = w.v8ctx.RunScript(sb.String(), "send_xhr_event.js")
	if err != nil {
		return ToJsError(err)
	}
	return nil
}
