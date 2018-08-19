package pool_test

import (
	"io"
	"io/ioutil"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/andy2046/pool"
)

var _ pool.IPool = &pool.Pool{}

func TestPool(t *testing.T) {
	done := make(chan struct{})
	mu := &sync.Mutex{}
	NumJobs := 50
	expected := make([]int, NumJobs)
	for i := range pool.Range(NumJobs) {
		expected[i] = i
	}
	result := make([]int, NumJobs)
	jobHandlerGenerator := func() pool.JobHandler {
		return func(j pool.Job) error {
			mu.Lock()
			defer mu.Unlock()
			result[int(j.ID)] = int(j.ID)
			time.Sleep(100 * time.Millisecond)
			return nil
		}
	}
	size := 3
	opt := func(c *pool.Config) error {
		c.InitPoolNum = size
		c.WorkerNum = 2
		return nil
	}

	p := pool.New(done, jobHandlerGenerator, opt)
	p.Start()

	if closed := p.Closed(); closed {
		t.Fatal("pool should not be closed")
	}
	if p.Size() != size {
		t.Fatalf("pool size should be %d \n", size)
	}

	for i := range pool.Range(NumJobs) {
		p.JobQueue <- pool.Job{
			ID: int32(i),
		}
	}

	p.Undispatch()
	if closed := p.Closed(); closed {
		t.Fatal("pool should not be closed")
	}
	if p.Size() != size-1 {
		t.Fatalf("pool size should be %d \n", size-1)
	}

	done <- struct{}{}
	// sleep for done channel to finish
	time.Sleep(1 * time.Second)

	if closed := p.Closed(); closed {
		t.Fatal("pool should not be closed")
	}
	if p.Size() != size-2 {
		t.Fatalf("pool size should be %d \n", size-2)
	}

	close(done)
	// sleep for done channel to finish
	time.Sleep(5 * time.Second)
	if !reflect.DeepEqual(result, expected) {
		t.Fatal("pool should finish all the jobs before exiting")
	}

	if closed := p.Closed(); !closed {
		t.Fatal("pool should be closed")
	}
	if p.Size() != 0 {
		t.Fatalf("pool size should be %d \n", 0)
	}

}

func BenchmarkPool(b *testing.B) {
	name := "pool"
	job := pool.Job{
		Name: "pool",
		Key:  "test",
	}
	done := make(chan struct{})
	jobHandlerGenerator := func() pool.JobHandler {
		return func(j pool.Job) error {
			io.Copy(ioutil.Discard, strings.NewReader(j.Name+j.Key))
			return nil
		}
	}
	size := 3
	opt := func(c *pool.Config) error {
		c.InitPoolNum = size
		c.WorkerNum = 5
		return nil
	}
	p := pool.New(done, jobHandlerGenerator, opt)
	p.Start()

	b.Run(name, func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			j := job
			p.JobQueue <- j
		}
	})
}
