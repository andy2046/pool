package pool_test

import (
	"io"
	"io/ioutil"
	"strings"
	"testing"
	"time"

	"github.com/andy2046/pool"
)

var _ pool.IPool = &pool.Pool{}

func TestPool(t *testing.T) {
	done := make(chan struct{})
	jobHandlerGenerator := func() pool.JobHandler {
		return func(pool.Job) error {
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

	if closed := p.Closed(); closed {
		t.Fatal("pool should not be closed")
	}
	if p.Size() != size {
		t.Fatalf("pool size should be %d \n", size)
	}

	p.JobQueue <- pool.Job{}

	p.Undispatch()
	if closed := p.Closed(); closed {
		t.Fatal("pool should not be closed")
	}
	if p.Size() != size-1 {
		t.Fatalf("pool size should be %d \n", size-1)
	}

	done <- struct{}{}
	// sleep 1 second for done channel to finish
	time.Sleep(time.Duration(1) * time.Second)

	if closed := p.Closed(); closed {
		t.Fatal("pool should not be closed")
	}
	if p.Size() != size-2 {
		t.Fatalf("pool size should be %d \n", size-2)
	}

	close(done)
	// sleep 1 second for done channel to finish
	time.Sleep(time.Duration(1) * time.Second)

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
