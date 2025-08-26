package comparer

import (
	"context"
	"sync"
)

type Comparer struct {
	ctx context.Context
	wg  sync.WaitGroup
	ch  chan Task
}

func (c *Comparer) Compare(file1 string, file2 string) error {
	select {
	case <-c.ctx.Done():
		return c.ctx.Err()
	default:
		c.ch <- Task{file1, file2}
		return nil
	}
}

func (c *Comparer) Close() error {
	close(c.ch)
	c.wg.Wait()
	return nil
}

func New(ctx context.Context, numTasks int, sink chan<- error) *Comparer {
	self := &Comparer{ctx: ctx, ch: make(chan Task)}
	self.wg.Add(numTasks)
	for range numTasks {
		go func() {
			defer self.wg.Done()
			for {
				select {
				case <-self.ctx.Done():
					return
				case task, ok := <-self.ch:
					if !ok {
						return
					}
					err := process(self.ctx, task)
					if err != nil {
						sink <- err
					}
				}
			}
		}()
	}
	return self
}
