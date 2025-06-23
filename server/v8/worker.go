package v8

import "C"
import (
	"encoding/json"
	"github.com/lizc2003/v8go"
	"github.com/lizc2003/vue-ssr-v8go/server/common/tlog"
	"github.com/lizc2003/vue-ssr-v8go/server/common/util"
	"os"
	"strings"
	"sync"
	"time"
)

type xhrEvent struct {
	XhrId    int               `json:"xhr_id"`
	Event    string            `json:"event"`
	Error    string            `json:"error,omitempty"`
	Status   int32             `json:"status,omitempty"`
	Headers  map[string]string `json:"headers,omitempty"`
	Response string            `json:"response,omitempty"`
	renderId int64
}

func (this *xhrEvent) Reset() {
	this.Event = ""
	this.Error = ""
	this.Status = 0
	this.Headers = nil
	this.Response = ""
}

func (this *xhrEvent) Clone() *xhrEvent {
	evt := *this
	return &evt
}

type SendMessageCallback func(mtype int64, param1 int64, param2 string, param3 string, param4 string, param5 string)

type Worker struct {
	Id              int64
	isolate         *v8go.Isolate
	inspectorClient *v8go.InspectorClient
	inspector       *v8go.Inspector
	v8ctx           *v8go.Context

	disposed   bool
	running    bool
	mutex      sync.Mutex
	evtQueue   []*xhrEvent
	callback   SendMessageCallback
	expireTime int64

	lastUsedHeap  uint64
	checkHeapTime int64
}

func NewWorker(callback SendMessageCallback, workerId int64) (*Worker, error) {
	isolate := v8go.NewIsolate()
	client := v8go.NewInspectorClient(newConsoleObj())
	inspector := v8go.NewInspector(isolate, client)
	v8ctx := v8go.NewContext(isolate)
	inspector.ContextCreated(v8ctx)

	worker := &Worker{
		Id:              workerId,
		isolate:         isolate,
		inspectorClient: client,
		inspector:       inspector,
		v8ctx:           v8ctx,
		callback:        callback,
	}

	script, err := isolate.CompileUnboundScript(gInitJs, gInitJsName, v8go.CompileOptions{CachedData: gInitJsCache})
	if err != nil {
		goto ERROR
	}
	_, err = script.Run(v8ctx)
	if err != nil {
		goto ERROR
	}

	if gServerJsCache != nil {
		script, err = isolate.CompileUnboundScript(gServerJs, gServerJsName, v8go.CompileOptions{CachedData: gServerJsCache})
		if err != nil {
			goto ERROR
		}
		_, err = script.Run(v8ctx)
		if err != nil {
			goto ERROR
		}
	} else if gServerFileName != "" {
		var content []byte
		content, err = os.ReadFile(gServerFileName)
		if err != nil {
			goto ERROR
		}
		_, err = v8ctx.RunScript(util.UnsafeBytes2Str(content), gServerJsName)
		if err != nil {
			goto ERROR
		}
	}

	err = setFunctionCallback(worker)
	if err != nil {
		goto ERROR
	}

	return worker, nil

ERROR:
	worker.Dispose()
	return nil, ToJsError(err)
}

func (this *Worker) Dispose() {
	this.mutex.Lock()
	defer this.mutex.Unlock()

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
	if !this.disposed {
		if this.running || len(this.evtQueue) > 0 {
			this.evtQueue = append(this.evtQueue, evt.Clone())
		} else {
			err = doSendXhrEvent(this, evt)
		}
	}
	this.mutex.Unlock()
	return err
}

func (this *Worker) CheckHeap() bool {
	if time.Now().Unix() > this.checkHeapTime {
		this.checkHeapTime = time.Now().Unix() + CheckHeapInterval
		heapSize := this.isolate.GetHeapStatistics().UsedHeapSize
		if heapSize > CheckHeapSize &&
			heapSize > this.lastUsedHeap*CheckHeapGrowRatio/100 {
			this.lastUsedHeap = heapSize
			this.isolate.FullGC()
			tlog.Infof("worker %d trigger gc, used heap size: %dM", this.Id, heapSize/1024/1024)
			return true
		}
	}
	return false
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
		var ret *v8go.Value
		args := info.Args()
		if len(args) >= 1 {
			ret, _ = v8go.NewValue(w.isolate,
				handleXMLHttpRequestCmd(w, args[0].String()),
			)
		}
		info.Release()
		return ret
	})
	v8goOT.Set("handleXhrCmd", xhrCmd)

	sendMessage := v8go.NewFunctionTemplate(w.isolate, func(info *v8go.FunctionCallbackInfo) *v8go.Value {
		if w.callback != nil {
			args := info.Args()
			if len(args) >= 6 {
				w.callback(args[0].Integer(), args[1].Integer(),
					args[2].String(),
					args[3].String(),
					args[4].String(),
					args[5].String(),
				)
			}
		}
		info.Release()
		return nil
	})
	v8goOT.Set("sendMessage", sendMessage)

	v8goObj, err := v8goOT.NewInstance(w.v8ctx)
	if err != nil {
		return err
	}
	return w.v8ctx.Global().Set("v8goGo", v8goObj)
}

func doSendXhrEvent(w *Worker, evt *xhrEvent) error {
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
		err = ToJsError(err)
	}

	if err != nil {
		tlog.Errorf("xhr %d-%d send %s, error: %v", evt.renderId, evt.XhrId, evt.Event, err)
	} else {
		tlog.Debugf("xhr %d-%d send %s ok", evt.renderId, evt.XhrId, evt.Event)
	}
	return err
}
