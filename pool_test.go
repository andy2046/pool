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
	mu := &sync.RWMutex{}
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
			i := j.Data.(int)
			result[i] = i
			time.Sleep(100 * time.Millisecond)
			return nil
		}
	}
	size := 3
	opt := func(c *pool.Config) error {
		c.InitPoolNum = size
		c.WorkerNum = 2
		c.Errors = true
		return nil
	}

	p := pool.New(done, jobHandlerGenerator, opt)
	p.Start()

	go func() {
		for err := range p.Errors {
			t.Logf("TestPool error: %v", err)
		}
	}()

	if closed := p.Closed(); closed {
		t.Fatal("pool should not be closed")
	}
	if p.Size() != size {
		t.Fatalf("pool size should be %d \n", size)
	}

	for i := range pool.Range(NumJobs) {
		p.JobQueue <- pool.Job{
			Data: i,
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
	time.Sleep(1 * time.Second)

	if closed := p.Closed(); closed {
		t.Fatal("pool should not be closed")
	}
	if p.Size() != size-2 {
		t.Fatalf("pool size should be %d \n", size-2)
	}

	close(done)
	// wait for jobs to finish
	for {
		time.Sleep(1 * time.Second)
		if p.Closed() {
			break
		}
	}
	mu.RLock()
	if !reflect.DeepEqual(result, expected) {
		mu.RUnlock()
		t.Fatal("pool should finish all the jobs before exiting")
	}
	mu.RUnlock()

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
		Data: "test",
	}
	done := make(chan struct{})
	jobHandlerGenerator := func() pool.JobHandler {
		return func(j pool.Job) error {
			io.Copy(ioutil.Discard, strings.NewReader(j.Data.(string)))
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

	b.ResetTimer()
	b.Run(name, func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			j := job
			p.JobQueue <- j
		}
	})
}
