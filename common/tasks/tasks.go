package tasks

import (
	"fmt"
	"runtime/debug"

	"golang.org/x/sync/errgroup"
)

type Group struct {
	errGroup   errgroup.Group
	HandleCrit func(err error)
}

func (t *Group) Go(fn func() error) {
	t.errGroup.Go(func() error {
		defer func() {
			if err := recover(); err != nil {
				debug.PrintStack()
				t.HandleCrit(fmt.Errorf("panic: %v", err))
			}
		}()
		return fn()
	})
}

func (t *Group) Wait() error {
	return t.errGroup.Wait()
}
