package dispatcher

import (
	"sync"
)

type compareInfo struct {
	a, b string
}

type dispatcher struct {
	lifecycle lifecycle
	sink      chan compareInfo
	wg        sync.WaitGroup
}

func (d *dispatcher) Emit(pathA, pathB string) (shouldContinue bool) {
	select {
	case <-d.lifecycle.Context().Done():
		return
	default:
		d.sink <- compareInfo{pathA, pathB}
		return true
	}
}

func (d *dispatcher) Close() error {
	close(d.sink)
	d.wg.Wait()
	defer d.lifecycle.Close()
	return d.lifecycle.GetError()
}

func (d *dispatcher) spawnThread() {
	defer d.wg.Done()
	handler := newBufferedComparer()
	for {
		select {
		case <-d.lifecycle.Context().Done():
			return
		case info, ok := <-d.sink:
			if !ok {
				return
			}
			err := handler.DispatchCompareRequest(info)
			if err != nil {
				d.lifecycle.EmitError(err)
			}
		}
	}
}

func Run(options Options) Emitter {
	if options.MetadataOnly {
		return noopEmitter{}
	}
	var lifecycle lifecycle
	if options.Failfast {
		lifecycle = newFailFastLifecycle(options.Context)
	} else {
		lifecycle = newCollectingLifecycle(options.Context)
	}
	self := &dispatcher{
		lifecycle: lifecycle,
		sink:      make(chan compareInfo, options.QueueSize),
	}
	self.wg.Add(options.Threads)
	for range options.Threads {
		go self.spawnThread()
	}
	return self
}
