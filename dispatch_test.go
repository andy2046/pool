package pool_test

import (
	"io"
	"io/ioutil"
	"strings"
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
	mu := &sync.Mutex{}
	wgJobHandler := &sync.WaitGroup{}
	wgJobHandler.Add(howManyJobs)
	jobHandler := func(j pool.Job) error {
		mu.Lock()
		defer mu.Unlock()
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
	// sleep for done channel to finish
	time.Sleep(1 * time.Second)

	if closed := dispatcher.Closed(); !closed {
		t.Fatal("dispatcher should be closed")
	}

	if i != howManyJobs {
		t.Fatal("job should be handled in job handler")
	}

}

func BenchmarkDispatcher(b *testing.B) {
	name := "dispatcher"
	job := pool.Job{
		Name: "pool",
		Key:  "test",
	}
	done := make(chan struct{})
	wgPool := &sync.WaitGroup{}
	wgPool.Add(1)
	numWorkers := 5
	jobQueue := make(chan pool.Job, 10)
	howManyJobs := 2
	i := 0
	mu := &sync.Mutex{}
	jobHandler := func(j pool.Job) error {
		mu.Lock()
		defer mu.Unlock()
		i++
		io.Copy(ioutil.Discard, strings.NewReader(j.Name+j.Key))
		return nil
	}
	dispatcher := pool.NewDispatcher(done, wgPool, numWorkers, jobQueue, jobHandler)
	dispatcher.Run()

	b.Run(name, func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for range pool.Range(howManyJobs) {
				j := job
				jobQueue <- j
			}
		}
	})
}
