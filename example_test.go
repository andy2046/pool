package pool_test

import (
	"github.com/andy2046/pool"
)

func ExamplePool() {
	done := make(chan struct{})
	jobHandlerGenerator := func() pool.JobHandler {
		return func(pool.Job) error {
			return nil
		}
	}
	size := 3
	opt := func(c *pool.Config) error {
		c.InitPoolNum = size
		return nil
	}

	p := pool.New(done, jobHandlerGenerator, opt)
	p.Start()

	p.JobQueue <- pool.Job{}

	p.Undispatch()

	done <- struct{}{}

	close(done)
}
