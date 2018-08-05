package pool_test

import (
	"sync"
	"testing"

	"github.com/andy2046/pool"
)

var _ pool.IWorker = &pool.Worker{}

func TestWorker(t *testing.T) {
	done := make(chan struct{})
	workerPool := make(chan chan pool.Job, 1)
	jobPool := make(chan struct{}, 1)
	howManyJobs := 2
	wg := &sync.WaitGroup{}
	wg.Add(1)
	i := 0
	jobHandler := func(j pool.Job) error {
		i++
		return nil
	}

	worker := pool.NewWorker(done, workerPool, wg, jobPool)
	worker.Start(jobHandler)
	if closed := worker.Closed(); closed {
		t.Fatal("worker should not be closed")
	}

	var basket chan pool.Job
	for range pool.Range(howManyJobs) {
		jobPool <- struct{}{}
		basket = <-workerPool
		basket <- pool.Job{}
	}

	done <- struct{}{}
	wg.Wait()
	if i != howManyJobs {
		t.Fatal("job should be handled in job handler")
	}
	if _, open := <-basket; open {
		t.Fatal("basket job channel should be closed")
	}
	if closed := worker.Closed(); !closed {
		t.Fatal("worker should be closed")
	}

}
