package pool_test

import (
	"log"
	"sync"
	"testing"
	"time"

	"github.com/andy2046/pool"
)

func PoolExample(t *testing.T) {
	done := make(chan struct{})
	mu := &sync.RWMutex{}
	data := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	sum := 0
	jobHandlerGenerator := func() pool.JobHandler {
		return func(j pool.Job) error {
			mu.Lock()
			defer mu.Unlock()
			sum += j.Data.(int)
			return nil
		}
	}
	size := 2
	opt := func(c *pool.Config) error {
		c.InitPoolNum = size
		c.WorkerNum = 5
		return nil
	}

	p := pool.New(done, jobHandlerGenerator, opt)
	p.Start()

	for i := range data {
		p.JobQueue <- pool.Job{
			Data: data[i],
		}
	}

	close(done)

	// sleep for done channel to finish
	time.Sleep(3 * time.Second)
	mu.RLock()
	log.Println(sum) // 55
	if sum != 55 {
		mu.RUnlock()
		t.Fatalf("sum is %d \n", sum)
	}
	mu.RUnlock()
}
