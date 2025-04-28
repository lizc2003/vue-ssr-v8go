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
	MaxXhrThreadCount    = 1000
	MaxVmLiftTime        = 24 * 3600 // seconds
	VmDeleteDelayTime    = 2 * time.Minute
	VmRequireTimeout     = 8 // seconds
	ProcessExitThreshold = 1000
)

var ErrorNoVm = errors.New("The VM instance cannot be acquired.")

type VmConfig struct {
	VmMaxCount int32
	VmLifeTime int32 // seconds
}

type VmMgr struct {
	callback         SendEventCallback
	xhrMgr           *xmlHttpRequestMgr
	workers          chan *Worker
	vmLifeTime       int64
	vmMaxCount       int32
	vmCurrentCount   int32
	unavailableCount int32
}

var ThisVmMgr *VmMgr

func NewVmMgr(env string, callback SendEventCallback, vc *VmConfig, xc *XhrConfig) (*VmMgr, error) {
	err := initVm(env)
	if err != nil {
		return nil, err
	}

	vmMaxCount := vc.VmMaxCount
	if vmMaxCount < 10 {
		vmMaxCount = 10
	}

	vmLifeTime := vc.VmLifeTime
	if vmLifeTime < 0 {
		vmLifeTime = 0
	} else if vmLifeTime > MaxVmLiftTime {
		vmLifeTime = MaxVmLiftTime
	}

	workers := make(chan *Worker, vmMaxCount+100)

	ThisVmMgr = &VmMgr{
		callback:       callback,
		xhrMgr:         NewXmlHttpRequestMgr(vmMaxCount*2, xc),
		workers:        workers,
		vmLifeTime:     int64(vmLifeTime),
		vmMaxCount:     vmMaxCount,
		vmCurrentCount: 0,
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
		bNewFail := false
		select {
		case worker := <-this.workers:
			if worker.Acquire() {
				ret = worker
			} else {
				busyWorkers = append(busyWorkers, worker)
			}
		default:
			if this.vmCurrentCount < this.vmMaxCount {
				atomic.AddInt32(&this.vmCurrentCount, 1)
				worker, err := NewWorker(this.callback)
				if err == nil {
					worker.SetExpireTime(time.Now().Unix() + this.vmLifeTime)
					worker.Acquire()
					ret = worker
				} else {
					tlog.Error(err)
					atomic.AddInt32(&this.vmCurrentCount, -1)
					bNewFail = true
				}
			} else {
				bNewFail = true
			}
		}

		if ret != nil {
			for _, w := range busyWorkers {
				this.workers <- w
			}
			return ret
		} else if bNewFail {
			if len(busyWorkers) > 0 {
				for _, w := range busyWorkers {
					this.workers <- w
				}
				busyWorkers = busyWorkers[:0]
			}
			time.Sleep(10 * time.Millisecond)
		}

		if time.Now().Unix()-reqStartTime > VmRequireTimeout {
			errCount := atomic.AddInt32(&this.unavailableCount, 1)
			if errCount >= ProcessExitThreshold {
				errMsg := "v8 vm unavailable too many times, exit!"
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
			atomic.AddInt32(&this.vmCurrentCount, -1)

			go func(w *Worker) {
				time.Sleep(VmDeleteDelayTime)
				w.Dispose()
			}(worker)
		} else {
			this.workers <- worker
		}
	}
}
