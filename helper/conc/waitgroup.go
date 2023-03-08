// Inspired by https://github.com/sourcegraph/conc
// but not implement recover panic and other features
// just a simple wait group, add context and logger support
// used to improve coroutine management and shutdown operations

package conc

import (
	"container/list"
	"context"
	"sync"

	"github.com/dogechain-lab/dogechain/helper/nocopy"
)

const (
	ErrContextDone = "context is done"
)

type WaitGroup interface {
	// Go() add a goroutine to wait group
	Go(f func(context.Context))

	// Wait() wait all goroutine finish, repanic the first panic
	Wait()

	// Close() cancel all goroutine
	Close()
}

type waitGroup struct {
	nocopy.NoCopy

	ctx    context.Context
	cancel context.CancelFunc

	wg   sync.WaitGroup
	onec sync.Once

	mux       sync.Mutex
	recovered *list.List
}

func NewWaitGroup(ctx context.Context) WaitGroup {
	ctx, cancel := context.WithCancel(ctx)

	return &waitGroup{
		ctx:       ctx,
		cancel:    cancel,
		recovered: list.New(),
	}
}

func (wgroup *waitGroup) Go(f func(context.Context)) {
	// if context is done, panic
	select {
	case <-wgroup.ctx.Done():
		panic(ErrContextDone)
	default:
	}

	wgroup.wg.Add(1)

	go func(wgroup *waitGroup) {
		defer wgroup.wg.Done()
		defer func() {
			if val := recover(); val != nil {
				wgroup.mux.Lock()
				wgroup.recovered.PushBack(val)
				wgroup.mux.Unlock()
			}
		}()

		f(wgroup.ctx)
	}(wgroup)
}

func (wgroup *waitGroup) Wait() {
	wgroup.wg.Wait()

	// repanic the first panic
	val := wgroup.recovered.Front()
	if val != nil && val.Value != nil {
		panic(val.Value)
	}
}

func (wgroup *waitGroup) Close() {
	wgroup.onec.Do(wgroup.cancel)
}
