package pool

import (
	"log"
	"sync"
)

type (
	// Job represents the job to be run.
	Job struct {
		Name     string `json:"name,omitempty"`
		ID       int32  `json:"id,omitempty"`
		Key      string `json:"key,omitempty"`
		WorkLoad []byte `json:"workload"`
	}

	// JobHandler completes the job.
	JobHandler func(Job) error

	// IWorker is the Worker interface.
	IWorker interface {
		Start(JobHandler)
		Closed() bool
	}

	// Worker represents the worker that executes the job.
	Worker struct {
		// worker pool the worker belongs to
		pool chan chan Job

		jobPool chan struct{}

		// worker's job channel for job assignment
		basket chan Job

		// done channel signals the worker to stop
		done chan struct{}

		wg *sync.WaitGroup

		// closed is true when worker received a signal to stop
		closed bool

		mu *sync.Mutex
	}
)

// NewWorker creates a worker.
func NewWorker(done chan struct{}, workerPool chan chan Job, wg *sync.WaitGroup, jobPool chan struct{}) *Worker {
	return &Worker{
		pool:    workerPool,
		jobPool: jobPool,
		basket:  make(chan Job, 1),
		done:    done,
		wg:      wg,
		mu:      &sync.Mutex{},
	}
}

// Start pushes the worker into worker queue, listens for signal to stop.
func (w *Worker) Start(handler JobHandler) {
	go func() {
		for {
			select {
			case <-w.jobPool:
				// worker has received a job request
				w.pool <- w.basket
				job := <-w.basket
				if err := handler(job); err != nil {
					log.Printf("Error handling job -> %s \n", err.Error())
				}
			case <-w.done:
				// worker has received a signal to stop
				w.done = nil
				w.mu.Lock()
				if !w.closed {
					w.closed = true
				}
				w.mu.Unlock()
				close(w.basket)
				w.wg.Done()
				log.Println("worker closed")
				return
			}
		}
	}()
}

// Closed returns true if worker received a signal to stop.
func (w *Worker) Closed() bool {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.closed
}
