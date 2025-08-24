package dispatcher

import (
	"context"
	"sync"
)

type compareInfo struct {
	a, b string
}

type dispatcher struct {
	cancel context.CancelCauseFunc
	sink   chan compareInfo
	ctx    context.Context
	wg     sync.WaitGroup
}

func (d *dispatcher) Emit(pathA, pathB string) (shouldContinue bool) {
	select {
	case <-d.ctx.Done():
		return
	default:
		d.sink <- compareInfo{pathA, pathB}
		return true
	}
}

func (d *dispatcher) Close() error {
	close(d.sink)
	d.wg.Wait()
	defer d.cancel(nil)
	return context.Cause(d.ctx)
}

func (d *dispatcher) spawnThread() {
	defer d.wg.Done()
	handler := newBufferedComparer()
	for {
		select {
		case <-d.ctx.Done():
			return
		case info, ok := <-d.sink:
			if !ok {
				return
			}
			err := handler.DispatchCompareRequest(info)
			if err != nil {
				d.cancel(err)
			}
		}
	}
}

func Run(options Options) Emitter {
	if options.MetadataOnly {
		return noopEmitter{}
	}
	ctx, cancel := context.WithCancelCause(options.Context)
	self := &dispatcher{
		ctx:    ctx,
		sink:   make(chan compareInfo, options.QueueSize),
		cancel: cancel,
	}
	self.wg.Add(options.Threads)
	for range options.Threads {
		go self.spawnThread()
	}
	return self
}
