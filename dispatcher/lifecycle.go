package dispatcher

import (
	"context"
	"errors"

	"github.com/decatile/treecmp/util"
)

type lifecycle interface {
	Context() context.Context
	EmitError(err error)
	GetError() error
	Close()
}

type failfastLifecycle struct {
	ctx    context.Context
	cancel context.CancelCauseFunc
}

func (f *failfastLifecycle) Context() context.Context {
	return f.ctx
}

func (f *failfastLifecycle) EmitError(err error) {
	f.cancel(err)
}

func (f *failfastLifecycle) GetError() error {
	return context.Cause(f.ctx)
}

func (f *failfastLifecycle) Close() {
	f.cancel(nil)
}

func newFailFastLifecycle(ctx context.Context) lifecycle {
	ctx, cancel := context.WithCancelCause(ctx)
	return &failfastLifecycle{ctx, cancel}
}

type collectingLifecycle struct {
	ctx       context.Context
	collector *util.Collector[error]
}

func (f *collectingLifecycle) Context() context.Context {
	return f.ctx
}

func (f *collectingLifecycle) EmitError(err error) {
	f.collector.Add(err)
}

func (f *collectingLifecycle) GetError() error {
	return errors.Join(f.collector.Values()...)
}

func (f *collectingLifecycle) Close() {

}

func newCollectingLifecycle(ctx context.Context) lifecycle {
	return &collectingLifecycle{ctx, util.NewCollector[error]()}
}
