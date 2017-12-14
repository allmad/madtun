package madtun_common

import (
	"sync"
)

type WaitGroup struct {
	errChan   chan error
	waitGroup sync.WaitGroup
}

func NewWaitGroup() *WaitGroup {
	return &WaitGroup{
		errChan:   make(chan error),
		waitGroup: sync.WaitGroup{},
	}
}

func (w *WaitGroup) Run(fn func() error) {
	w.waitGroup.Add(1)
	go func() {
		defer w.waitGroup.Done()
		if err := fn(); err != nil {
			w.errChan <- err
		}
	}()
}

func (w *WaitGroup) Wait() error {
	go func() {
		w.waitGroup.Wait()
		w.errChan <- nil
	}()
	return <-w.errChan
}
