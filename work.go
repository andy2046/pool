package pool

import (
	"sync"

	opentracing "github.com/opentracing/opentracing-go"
)

type (
	// Job represents the job to be run.
	Job struct {
		Data interface{}
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
		pool chan<- chan Job

		jobPool <-chan struct{}

		// worker's job channel for job assignment
		basket chan Job

		// done channel signals the worker to stop
		done <-chan struct{}

		errors chan error

		wg *sync.WaitGroup

		// closed is true when worker received a signal to stop
		closed bool

		mu *sync.Mutex

		tracer opentracing.Tracer
	}
)

// NewWorker creates a worker.
func NewWorker(done <-chan struct{}, workerPool chan<- chan Job, wg *sync.WaitGroup,
	jobPool <-chan struct{}, errors chan error, tracer opentracing.Tracer) *Worker {
	return &Worker{
		pool:    workerPool,
		jobPool: jobPool,
		basket:  make(chan Job, 1),
		done:    done,
		errors:  errors,
		wg:      wg,
		mu:      &sync.Mutex{},
		tracer:  tracer,
	}
}

// Start pushes the worker into worker queue, listens for signal to stop.
func (w *Worker) Start(handler JobHandler) {
	if w.tracer != nil {
		span := w.tracer.StartSpan("Start")
		defer span.Finish()
		span.SetTag("component", "Worker")
	}

	go func() {
		var span opentracing.Span
		for {
			select {
			case <-w.jobPool:
				if w.tracer != nil {
					span = w.tracer.StartSpan("JobHandler")
					span.SetTag("component", "Worker")
				}
				// worker has received a job request
				w.pool <- w.basket
				job := <-w.basket
				if err := handler(job); err != nil {
					logger.Printf("Error handling job -> %s \n", err.Error())
					if w.errors != nil {
						w.errors <- err
					}
				}
				if span != nil {
					span.Finish()
				}
			case <-w.done:
				// worker has received a signal to stop
				w.done = nil
				close(w.basket)
				w.wg.Done()
				w.mu.Lock()
				if !w.closed {
					w.closed = true
				}
				w.mu.Unlock()
				if verbose() {
					logger.Println("worker closed")
				}
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
