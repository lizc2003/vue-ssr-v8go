package v8

import (
	"errors"
	"github.com/lizc2003/vue-ssr-v8go/server/common/alarm"
	"github.com/lizc2003/vue-ssr-v8go/server/common/tlog"
	"os"
	"sync/atomic"
	"time"
)

const (
	MaxVmInstances = 2000
	MinVmInstances = 1
	MaxVmLiftTime  = 24 * 3600 // seconds
	MaxXhrThreads  = 4000
	MinXhrThreads  = 2

	VmDeleteDelayTime    = 150 * time.Second
	VmAcquireTimeout     = 5 // seconds
	ProcessExitThreshold = 1000
)

var ErrorNoVm = errors.New("the v8 instance cannot be acquired.")

type VmConfig struct {
	UseStrict        bool  `toml:"use_strict"`
	MaxInstances     int32 `toml:"max_instances"`
	InstanceLifetime int32 `toml:"instance_lifetime"`
	XhrThreads       int32 `toml:"xmlhttprequest_threads"`
}

type VmMgr struct {
	callback SendEventCallback
	xhrMgr   *XmlHttpRequestMgr
	workers  chan *Worker

	vmLifetime         int64
	vmMaxInstances     int32
	vmCurrentInstances int32
	vmAcquireFailCount int32
}

var ThisVmMgr *VmMgr

func NewVmMgr(env string, serverDir string, callback SendEventCallback, vc *VmConfig, ac *ApiConfig) (*VmMgr, error) {
	err := initVm(env, serverDir, vc.UseStrict)
	if err != nil {
		return nil, err
	}

	xhrMgr, err := NewXmlHttpRequestMgr(vc.XhrThreads, ac)
	if err != nil {
		return nil, err
	}

	vmMaxInstances := vc.MaxInstances
	if vmMaxInstances < MinVmInstances {
		vmMaxInstances = MinVmInstances
	} else if vmMaxInstances > MaxVmInstances {
		vmMaxInstances = MaxVmInstances
	}

	vmLifetime := vc.InstanceLifetime
	if vmLifetime < 0 {
		vmLifetime = 0
	} else if vmLifetime > MaxVmLiftTime {
		vmLifetime = MaxVmLiftTime
	}

	workers := make(chan *Worker, vmMaxInstances+100)
	ThisVmMgr = &VmMgr{
		callback:           callback,
		xhrMgr:             xhrMgr,
		workers:            workers,
		vmLifetime:         int64(vmLifetime),
		vmMaxInstances:     vmMaxInstances,
		vmCurrentInstances: 0,
	}

	return ThisVmMgr, nil
}

func (this *VmMgr) Execute(code string, scriptName string) error {
	w := this.acquireWorker()

	if w == nil {
		errMsg := ErrorNoVm.Error()
		tlog.Error(errMsg)
		alarm.SendAlert(errMsg)
		return ErrorNoVm
	}
	err := w.Execute(code, scriptName)

	// tlog.Debug(w.Execute(`console.debug(dumpObject(globalThis))`, "test.js"))

	this.releaseWorker(w)

	if err != nil {
		tlog.Error(err)
	}
	return err
}

func (this *VmMgr) acquireWorker() *Worker {
	var busyWorkers []*Worker
	reqStartTime := time.Now().Unix()
	for {
		var ret *Worker
		bNewFailed := false
		bReachMax := false
		select {
		case worker := <-this.workers:
			if worker.Acquire() {
				ret = worker
			} else {
				busyWorkers = append(busyWorkers, worker)
			}
		default:
			if this.vmCurrentInstances < this.vmMaxInstances {
				worker, err := NewWorker(this.callback)
				if err == nil {
					atomic.AddInt32(&this.vmCurrentInstances, 1)
					worker.SetExpireTime(time.Now().Unix() + this.vmLifetime)
					worker.Acquire()
					ret = worker
				} else {
					tlog.Error(err)
					bNewFailed = true
				}
			} else {
				bReachMax = true
			}
		}

		if ret != nil || bNewFailed {
			for _, w := range busyWorkers {
				this.workers <- w
			}
			return ret
		}

		if bReachMax {
			if len(busyWorkers) > 0 {
				for _, w := range busyWorkers {
					this.workers <- w
				}
				busyWorkers = busyWorkers[:0]
			}
			time.Sleep(10 * time.Millisecond)
		}

		if time.Now().Unix()-reqStartTime > VmAcquireTimeout {
			failCount := atomic.AddInt32(&this.vmAcquireFailCount, 1)
			if failCount >= ProcessExitThreshold {
				errMsg := "too many failures acquiring v8 instance, exit!"
				tlog.Error(errMsg)
				alarm.SendAlert(errMsg)
				os.Exit(1)
			}
			return nil
		}
	}
}

func (this *VmMgr) releaseWorker(worker *Worker) {
	if worker != nil {
		worker.Release()

		if time.Now().Unix() >= worker.GetExpireTime() {
			atomic.AddInt32(&this.vmCurrentInstances, -1)

			go func(w *Worker) {
				time.Sleep(VmDeleteDelayTime)
				w.Dispose()
			}(worker)
		} else {
			this.workers <- worker
		}
	}
}
