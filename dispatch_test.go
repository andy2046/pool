package pool_test

import (
	"sync"
	"testing"
	"time"

	"github.com/andy2046/pool"
)

var _ pool.IDispatcher = &pool.Dispatcher{}

func TestDispatcher(t *testing.T) {
	done := make(chan struct{})
	wgPool := &sync.WaitGroup{}
	wgPool.Add(1)
	numWorkers := 5
	jobQueue := make(chan pool.Job, 10)
	howManyJobs := 2
	i := 0
	wgJobHandler := &sync.WaitGroup{}
	wgJobHandler.Add(howManyJobs)
	jobHandler := func(j pool.Job) error {
		i++
		wgJobHandler.Done()
		return nil
	}

	dispatcher := pool.NewDispatcher(done, wgPool, numWorkers, jobQueue, jobHandler)
	dispatcher.Run()
	if closed := dispatcher.Closed(); closed {
		t.Fatal("dispatcher should not be closed")
	}

	for range pool.Range(howManyJobs) {
		jobQueue <- pool.Job{}
	}

	wgJobHandler.Wait()

	done <- struct{}{}
	// sleep 1 second for done channel to finish
	time.Sleep(time.Duration(1) * time.Second)

	if closed := dispatcher.Closed(); !closed {
		t.Fatal("dispatcher should be closed")
	}

	if i != howManyJobs {
		t.Fatal("job should be handled in job handler")
	}

}
