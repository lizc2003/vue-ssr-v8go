package v8

import (
	"errors"
	"fmt"
	"github.com/lizc2003/v8go"
	"github.com/lizc2003/vue-ssr-v8go/server/common/alarm"
	"github.com/lizc2003/vue-ssr-v8go/server/common/defs"
	"github.com/lizc2003/vue-ssr-v8go/server/common/tlog"
	"math/rand"
	"os"
	"path"
	"sync"
	"sync/atomic"
	"time"
)

const (
	MaxVmInstances = 1000
	MinVmInstances = 1
	MaxVmLiftTime  = 24 * 3600 // seconds
	MaxXhrThreads  = 2000
	MinXhrThreads  = 2

	VmAcquireTimeout     = 5 // seconds
	ProcessExitThreshold = 1000

	CheckHeapInterval  = 5 // seconds
	CheckHeapSize      = 512 * 1024 * 1024
	CheckHeapGrowRatio = 120 // 120%
	MinHeapSizeLimit   = (CheckHeapSize / 1024 / 1024) * 150 / 100
)

var ErrorNoVm = errors.New("the v8 instance cannot be acquired.")

type VmConfig struct {
	UseStrict        bool  `toml:"use_strict"`
	HeapSizeLimit    int32 `toml:"heap_size_limit"`
	MaxInstances     int32 `toml:"max_instances"`
	InstanceLifetime int32 `toml:"instance_lifetime"`
	DeleteDelayTime  int32 `toml:"delete_delay_time"`
	XhrThreads       int32 `toml:"xmlhttprequest_threads"`
}

type VmMgr struct {
	callback SendEventCallback
	xhrMgr   *XmlHttpRequestMgr
	workers  chan *Worker

	bDev               bool
	mutex              sync.Mutex
	vmDeleteDelayTime  time.Duration
	vmMaxId            int64
	vmLifetime         int64
	vmMaxInstances     int32
	vmCurrentInstances int32
	vmAcquireFailCount int32

	DumpHeapDir string
	isDumpHeap  int32
}

var ThisVmMgr *VmMgr

func NewVmMgr(env string, serverDir string, callback SendEventCallback, vc *VmConfig, ac *ApiConfig) (*VmMgr, error) {
	bDev := false
	if env == defs.EnvDev {
		bDev = true
	}
	err := initVm(bDev, serverDir, vc.UseStrict, vc.HeapSizeLimit)
	if err != nil {
		return nil, err
	}

	heapSizeLimit := uint64(0)
	{
		iso := v8go.NewIsolate()
		heapSizeLimit = iso.GetHeapStatistics().HeapSizeLimit
		iso.Dispose()
	}
	tlog.Infof("v8 version: %s, heap size limit: %dM", v8go.Version(), heapSizeLimit/1024/1024)

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

	vmDeleteDelayTime := vc.DeleteDelayTime
	if vmDeleteDelayTime <= 0 {
		vmDeleteDelayTime = 1
	}

	workers := make(chan *Worker, vmMaxInstances+100)
	ThisVmMgr = &VmMgr{
		callback:           callback,
		xhrMgr:             xhrMgr,
		workers:            workers,
		bDev:               bDev,
		vmLifetime:         int64(vmLifetime),
		vmMaxInstances:     vmMaxInstances,
		vmDeleteDelayTime:  time.Duration(vmDeleteDelayTime) * time.Second,
		vmCurrentInstances: 0,
	}

	return ThisVmMgr, nil
}

func (this *VmMgr) SignalDumpHeap() {
	atomic.StoreInt32(&this.isDumpHeap, 1)
}

func (this *VmMgr) Execute(code string, scriptName string) (int64, error) {
	w := this.acquireWorker()

	if w == nil {
		errMsg := ErrorNoVm.Error()
		tlog.Error(errMsg)
		alarm.SendAlert(errMsg)
		return 0, ErrorNoVm
	}
	workerId := w.Id
	err := w.Execute(code, scriptName)

	// tlog.Debug(w.Execute(`console.debug(dumpObject(globalThis))`, "test.js"))

	bChecked := w.CheckHeap()
	if (bChecked && this.bDev) ||
		atomic.CompareAndSwapInt32(&this.isDumpHeap, 1, 0) {
		n := rand.Int31n(1000)
		fName := time.Now().Format("20060102150405") + fmt.Sprintf("-%03d.heapsnapshot", n)
		fName = path.Join(this.DumpHeapDir, fName)
		w.isolate.WriteSnapshot(fName, true)
	}

	this.releaseWorker(w)

	if err != nil {
		tlog.Error(err)
	}
	return workerId, err
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
			bReachMax = true
			if this.vmCurrentInstances < this.vmMaxInstances {
				this.mutex.Lock()
				if this.vmCurrentInstances < this.vmMaxInstances {
					atomic.AddInt32(&this.vmCurrentInstances, 1)
					bReachMax = false
				}
				this.mutex.Unlock()
			}

			if !bReachMax {
				workerId := atomic.AddInt64(&this.vmMaxId, 1)
				worker, err := NewWorker(this.callback, workerId)
				if err == nil {
					tlog.Infof("vm created: %d", workerId)
					worker.SetExpireTime(time.Now().Unix() + this.vmLifetime)
					worker.Acquire()
					ret = worker
				} else {
					tlog.Error(err)
					atomic.AddInt32(&this.vmCurrentInstances, -1)
					bNewFailed = true
				}
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
				time.Sleep(this.vmDeleteDelayTime)
				w.Dispose()
				tlog.Infof("vm deleted: %d", w.Id)
			}(worker)
		} else {
			this.workers <- worker
		}
	}
}
